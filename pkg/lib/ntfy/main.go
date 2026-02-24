package ntfy

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"
	mailConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	ntfyConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/ntfy"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/install"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/google/project"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
)

// Install Ntfy on the remote server via SSH and create necessary resources.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// ntfyConfig: Ntfy configuration.
// mailConfig: Mail configuration.
// dnsConfig: DNS configuration.
// dependsOn: List of Pulumi resources that this installation depends on.
func Install(ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	ntfyConfig *ntfyConf.Config,
	mailConfig *mailConf.Config,
	dnsConfig *dns.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) error {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	dnsErr := createDNSRecords(ctx, mailConfig, dnsConfig, ntfyConfig)
	if dnsErr != nil {
		return dnsErr
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "ntfy", conn, opts...)
	if prepErr != nil {
		return prepErr
	}

	dockerCompose, _ := template.Render("./assets/ntfy/docker-compose.yml.j2", map[string]any{
		"domain": ntfyConfig.Domain.Name,
	})
	dockerComposeCopy, dockerComposeHash, dcErr := install.DockerCompose(
		ctx,
		"ntfy",
		pulumi.String(dockerCompose),
		false,
		conn,
		opts...)
	if dcErr != nil {
		return dcErr
	}

	configFileCopy, configFileHash := createConfig(
		ctx,
		conn,
		ntfyConfig,
		opts...)

	_, cronErr := install.Cron(ctx, "ntfy", conn, opts...)
	if cronErr != nil {
		return cronErr
	}

	opts, systemdServiceHash, shErr := install.SystemDService(ctx, "ntfy", conn, opts...)
	if shErr != nil {
		return shErr
	}

	ntfyVersion := install.Version("./outputs/ntfy_docker-compose.yml", "ntfy", dockerComposeHash)

	installFn, _ := ntfyVersion.ApplyT(func(version string) string {
		ic, _ := template.Render("./assets/ntfy/install.sh.j2", map[string]any{
			"version": version,
			"project": project.GetOrDefault(ctx, nil),
			"bucket": map[string]string{
				"id":   config.BackupBucketID,
				"path": config.BackupBucketPath,
			},
		})
		return ic
	}).(pulumi.StringOutput)
	installTask := pulumi.All(configFileCopy, dockerComposeCopy).
		ApplyT(func(args []any) pulumi.ResourceOption {
			configCopy, _ := args[0].(pulumi.ResourceOption)
			dockerCopy, _ := args[1].(pulumi.ResourceOption)

			cmd, _ := remote.NewCommand(
				ctx,
				"remote-command-install-ntfy",
				&remote.CommandArgs{
					Create: installFn,
					Update: installFn,
					Triggers: pulumi.Array{
						pulumi.String(*systemdServiceHash),
						dockerComposeHash,
						configFileHash,
						ntfyVersion,
					},
					Connection: conn,
				},
				append(opts, configCopy, dockerCopy)...)
			return pulumi.DependsOn([]pulumi.Resource{cmd})
		})

	installTask.ApplyT(func(installT any) error {
		installer, _ := installT.(pulumi.ResourceOption)
		install.Postinstall(ctx, "ntfy", pulumi.Array{}, conn, append(opts, installer)...)
		return nil
	})

	return nil
}
