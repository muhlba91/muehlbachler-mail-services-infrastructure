package scaleway

import (
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/scaleway"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/install"
	"github.com/muhlba91/pulumi-shared-library/pkg/model/scaleway/iam/application"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// Install scaleway CLI on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// application: The Scaleway application containing the credentials to be installed on the server.
// scalewayConfig: Configuration for Scaleway, including project information.
// dependsOn: Pulumi resource option to specify dependencies.
func Install(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	application *application.Application,
	scalewayConfig *scaleway.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*remote.Command, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "scaleway", conn, opts...)
	if prepErr != nil {
		return nil, prepErr
	}

	rclone, _ := pulumi.All(application.Key.AccessKey, application.Key.SecretKey).ApplyT(func(args []any) string {
		accessKey, ok1 := args[0].(string)
		secretKey, ok2 := args[1].(string)
		if !ok1 || !ok2 {
			log.Error().Msg("[scaleway][install] failed to cast application keys to string")
		}

		tpl, tErr := template.Render("./assets/scaleway/rclone.conf.j2", map[string]string{
			"accessKey":      accessKey,
			"secretKey":      secretKey,
			"organizationId": scalewayConfig.OrganizationID,
			"defaultRegion":  config.ScalewayDefaultRegion,
		})
		if tErr != nil {
			log.Error().Err(tErr).Msg("[scaleway][install] failed to render credentials template")
		}

		return tpl
	}).(pulumi.StringOutput)
	scalewayRcloneHash := file.WritePulumi("./outputs/scaleway_rclone.conf", rclone).
		ApplyT(func(_ string) string {
			hash, _ := file.Hash("./outputs/scaleway_rclone.conf")
			return *hash
		})
	scalewayRcloneCopy := scalewayRcloneHash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			"remote-copy-scaleway-rclone-conf",
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset("./outputs/scaleway_rclone.conf"),
				RemotePath: pulumi.String("/opt/scaleway/rclone.conf"),
				Triggers:   pulumi.Array{scalewayRcloneHash},
				Connection: conn,
			},
			opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	installFn, iErr := file.ReadContents("./assets/scaleway/install.sh")
	if iErr != nil {
		return nil, iErr
	}
	return remote.NewCommand(ctx, "remote-command-install-scaleway", &remote.CommandArgs{
		Create:     pulumi.StringPtr(installFn),
		Update:     pulumi.StringPtr(installFn),
		Triggers:   pulumi.Array{scalewayRcloneHash},
		Connection: conn,
	}, append(opts, install.CollectResourceOptions([]pulumi.Output{scalewayRcloneCopy})...)...)
}
