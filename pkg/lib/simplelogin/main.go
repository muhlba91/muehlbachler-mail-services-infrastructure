package simplelogin

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"
	mailConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/server"
	simpleloginConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/simplelogin"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/dkim"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/install"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
)

// Install SimpleLogin on the remote server via SSH and create necessary resources.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// postgresqlUsers: Map of PostgreSQL users needed for SimpleLogin.
// simpleloginConfig: Configuration for SimpleLogin installation.
// serverConfig: Configuration of the server where SimpleLogin is installed.
// dependsOn: List of Pulumi resources that this installation depends on.
func Install(ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	postgresqlUsers map[string]*pulumi.AnyOutput,
	simpleloginConfig *simpleloginConf.Config,
	serverConfig *server.Config,
	mailConfig *mailConf.Config,
	dnsConfig *dns.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*dkim.Data, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "simplelogin", conn, opts...)
	if prepErr != nil {
		return nil, prepErr
	}

	dockerCompose, _ := template.Render("./assets/simplelogin/docker-compose.yml.j2", map[string]any{
		"domain": simpleloginConfig.Domain,
	})
	dockerComposeCopy, dockerComposeHash, dcErr := install.DockerCompose(
		ctx,
		"simplelogin",
		pulumi.String(dockerCompose),
		false,
		conn,
		opts...)
	if dcErr != nil {
		return nil, dcErr
	}

	dkimKey, dkimKeyCopy, dkErr := createDKIMConfig(ctx, conn, simpleloginConfig, mailConfig, dnsConfig, opts...)
	if dkErr != nil {
		return nil, dkErr
	}
	envFileCopy, envFileHash := createConfig(ctx, conn, postgresqlUsers, simpleloginConfig, serverConfig, opts...)

	opts, systemdServiceHash, shErr := install.SystemDService(ctx, "simplelogin", conn, opts...)
	if shErr != nil {
		return nil, shErr
	}

	simpleloginVersion := install.Version("./outputs/simplelogin_docker-compose.yml", "app", dockerComposeHash)

	initShHash, ishErr := file.Hash("./assets/simplelogin/init.sh")
	if ishErr != nil {
		return nil, ishErr
	}
	initShCopy, ishcErr := remote.NewCopyToRemote(
		ctx,
		"remote-copy-simplelogin-init-sh",
		&remote.CopyToRemoteArgs{
			Source:     pulumi.NewFileAsset("./assets/simplelogin/init.sh"),
			RemotePath: pulumi.String("/opt/simplelogin/init.sh"),
			Triggers:   pulumi.Array{pulumi.String(*initShHash)},
			Connection: conn,
		},
		opts...)
	if ishcErr != nil {
		return nil, ishcErr
	}

	installFn, _ := simpleloginVersion.ApplyT(func(version string) string {
		ic, _ := template.Render("./assets/simplelogin/install.sh.j2", map[string]any{
			"version": version,
		})
		return ic
	}).(pulumi.StringOutput)
	_ = pulumi.All(dkimKeyCopy, envFileCopy, initShCopy, dockerComposeCopy).
		ApplyT(func(args []any) pulumi.ResourceOption {
			dkimCopy, _ := args[0].(pulumi.ResourceOption)
			envCopy, _ := args[1].(pulumi.ResourceOption)
			initCopy, _ := args[2].(pulumi.ResourceOption)
			dockerCopy, _ := args[3].(pulumi.ResourceOption)

			cmd, _ := remote.NewCommand(
				ctx,
				"remote-command-install-simplelogin",
				&remote.CommandArgs{
					Create: installFn,
					Update: installFn,
					Triggers: pulumi.Array{
						pulumi.String(*systemdServiceHash),
						dockerComposeHash,
						envFileHash,
						pulumi.String(*initShHash),
						simpleloginVersion,
					},
					Connection: conn,
				},
				append(opts, dkimCopy, envCopy, initCopy, dockerCopy)...)
			return pulumi.DependsOn([]pulumi.Resource{cmd})
		})

	return dkimKey, nil
}
