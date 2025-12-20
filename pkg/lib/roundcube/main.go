package roundcube

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"
	mailConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	rcConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/roundcube"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/install"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
)

// Install Mailcow on the remote server via SSH and create necessary resources.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// postgresqlUsers: Map of PostgreSQL users needed for Mailcow.
// apiKeyReadWrite: The API key with read and write permissions for Mailcow.
// roundcubeConfig: Configuration for Roundcube installation.
// mailConfig: Configuration for mail services.
// dnsConfig: Configuration for DNS settings.
// dependsOn: List of Pulumi resources that this installation depends on.
func Install(ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	postgresqlUsers map[string]*pulumi.AnyOutput,
	apiKeyReadWrite pulumi.StringOutput,
	roundcubeConfig *rcConf.Config,
	mailConfig *mailConf.Config,
	dnsConfig *dns.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) error {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	dnsErr := createDNSRecords(ctx, mailConfig, dnsConfig, roundcubeConfig)
	if dnsErr != nil {
		return dnsErr
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "roundcube", conn, opts...)
	if prepErr != nil {
		return prepErr
	}

	dockerCompose, _ := template.Render("./assets/roundcube/docker-compose.yml.j2", map[string]any{
		"domain": roundcubeConfig.Domain.Name,
	})
	dockerComposeCopy, dockerComposeHash, dcErr := install.DockerCompose(
		ctx,
		"roundcube",
		pulumi.String(dockerCompose),
		false,
		conn,
		opts...)
	if dcErr != nil {
		return dcErr
	}

	nginxConfCopy, configFileCopy, configFileHash := createConfig(
		ctx,
		conn,
		postgresqlUsers,
		roundcubeConfig,
		mailConfig,
		opts...,
	)
	opts = append(opts, pulumi.DependsOn([]pulumi.Resource{nginxConfCopy}))

	opts, systemdServiceHash, shErr := install.SystemDService(ctx, "roundcube", conn, opts...)
	if shErr != nil {
		return shErr
	}

	roundcubeVersion := install.Version("./outputs/roundcube_docker-compose.yml", "webmail", dockerComposeHash)

	installFn, _ := roundcubeVersion.ApplyT(func(version string) string {
		ic, _ := template.Render("./assets/roundcube/install.sh.j2", map[string]any{
			"version": version,
		})
		return ic
	}).(pulumi.StringOutput)
	installTask := pulumi.All(configFileCopy, dockerComposeCopy).
		ApplyT(func(args []any) pulumi.ResourceOption {
			configCopy, _ := args[0].(pulumi.ResourceOption)
			dockerCopy, _ := args[1].(pulumi.ResourceOption)
			cmd, _ := remote.NewCommand(
				ctx,
				"remote-command-install-roundcube",
				&remote.CommandArgs{
					Create: installFn,
					Update: installFn,
					Triggers: pulumi.Array{
						pulumi.String(*systemdServiceHash),
						dockerComposeHash,
						configFileHash,
						roundcubeVersion,
					},
					Connection: conn,
				},
				append(opts, configCopy, dockerCopy)...)
			return pulumi.DependsOn([]pulumi.Resource{cmd})
		})

	postinstall(ctx, conn, apiKeyReadWrite, mailConfig, installTask, opts...)

	return nil
}
