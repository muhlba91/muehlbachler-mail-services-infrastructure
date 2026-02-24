package mailcow

import (
	"os"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert/yaml"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"
	mailConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	mcModel "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/mailcow"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/install"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/mail"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
)

// Install Mailcow on the remote server via SSH and create necessary resources.
// ctx: Pulumi context.
// ipv4Address: The public IPv4 address of the server.
// ipv6Address: The public IPv6 address of the server.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// secrets: Mailcow secrets needed for installation.
// mailConfig: Mail configuration.
// dnsConfig: DNS configuration.
// dependsOn: List of Pulumi resources that this installation depends on.
//
//nolint:funlen // Function is long but clear in its purpose.
func Install(ctx *pulumi.Context,
	ipv4Address pulumi.StringOutput,
	ipv6Address pulumi.StringOutput,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	secrets *mcModel.Secrets,
	mailConfig *mailConf.Config,
	dnsConfig *dns.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) error {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "mailcow", conn, opts...)
	if prepErr != nil {
		return prepErr
	}

	dockerCompose, _ := secrets.APIKeyRead.ApplyT(func(key string) string {
		dc, _ := template.Render("./assets/mailcow/docker-compose.override.yml.j2", map[string]any{
			"mailname": mail.Mailname(*mailConfig.Main.Name),
			"apiKey":   key,
		})
		return dc
	}).(pulumi.StringOutput)
	dockerComposeCopy, dockerComposeHash, dcErr := install.DockerCompose(
		ctx,
		"mailcow",
		dockerCompose,
		true,
		conn,
		opts...)
	if dcErr != nil {
		return dcErr
	}

	configFileCopy, configFileHash := createConfig(
		ctx,
		conn,
		ipv4Address,
		ipv6Address,
		secrets,
		mailConfig,
		dnsConfig,
		opts...,
	)

	_, cronErr := install.Cron(ctx, "mailcow", conn, opts...)
	if cronErr != nil {
		return cronErr
	}

	opts, systemdServiceHash, shErr := install.SystemDService(ctx, "mailcow", conn, opts...)
	if shErr != nil {
		return shErr
	}

	mailcowVersion, _ := dockerComposeHash.ApplyT(func(_ string) string {
		data, rErr := os.ReadFile("./outputs/mailcow_docker-compose.override.yml")
		if rErr != nil {
			return ""
		}

		s := strings.ReplaceAll(string(data), "#", "")
		var parsed map[string]any
		if pErr := yaml.Unmarshal([]byte(s), &parsed); pErr != nil {
			return ""
		}

		v, ok := parsed["version"].(string)
		if !ok {
			return ""
		}
		return v
	}).(pulumi.StringOutput)

	//nolint:godox // TODO is required
	// TODO: restore doesn't work automated - https://github.com/mailcow/mailcow-dockerized/pull/5934
	installFn, _ := mailcowVersion.ApplyT(func(version string) string {
		ic, _ := template.Render("./assets/mailcow/install.sh.j2", map[string]any{
			"version": version,
			"bucket": map[string]string{
				"id":   config.BackupBucketID,
				"path": config.BackupBucketPath,
			},
			"dkimSignHeaders": strings.Join(mailConfig.DkimSignHeaders, ":"),
		})
		return ic
	}).(pulumi.StringOutput)
	installFileHash := file.WritePulumi("./outputs/mailcow_install.sh", installFn).
		ApplyT(func(_ string) string {
			hash, _ := file.Hash("./outputs/mailcow_install.sh")
			return *hash
		})
	installFileCopy := installFileHash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			"remote-copy-install-sh",
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset("./outputs/mailcow_install.sh"),
				RemotePath: pulumi.String("/opt/mailcow/install.sh"),
				Triggers:   pulumi.Array{installFileHash},
				Connection: conn,
			},
			opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})
	installTask, _ := pulumi.All(configFileCopy, installFileCopy, dockerComposeCopy).
		ApplyT(func(args []any) pulumi.ResourceOption {
			configCopy, _ := args[0].(pulumi.ResourceOption)
			installCopy, _ := args[1].(pulumi.ResourceOption)
			dockerCopy, _ := args[2].(pulumi.ResourceOption)

			cmd, _ := remote.NewCommand(
				ctx,
				"remote-command-install-mailcow",
				&remote.CommandArgs{
					Create: pulumi.Sprintf("bash /opt/mailcow/install.sh"),
					Update: pulumi.Sprintf("bash /opt/mailcow/install.sh"),
					Triggers: pulumi.Array{
						pulumi.String(*systemdServiceHash),
						dockerComposeHash,
						configFileHash,
						mailcowVersion,
					},
					Connection: conn,
				},
				append(opts, configCopy, installCopy, dockerCopy)...)
			return pulumi.DependsOn([]pulumi.Resource{cmd})
		}).(pulumi.AnyOutput)

	postinstall(ctx, conn, installTask, opts...)

	return nil
}
