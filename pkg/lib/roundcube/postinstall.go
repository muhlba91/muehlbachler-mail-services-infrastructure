package roundcube

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	mailConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/install"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/mail"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/storage"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/storage/google"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
)

// postinstall performs post-installation tasks for Roundcube.
// ctx: Pulumi context.
// conn: SSH connection arguments.
// apiKeyReadWrite: The API key with read and write permissions for Mailcow.
// mailConfig: Configuration for mail services.
// installTask: The installation task output to depend on.
// opts: Additional Pulumi resource options.
func postinstall(ctx *pulumi.Context,
	conn *remote.ConnectionArgs,
	apiKeyReadWrite pulumi.StringOutput,
	mailConfig *mailConf.Config,
	installTask pulumi.Output,
	opts ...pulumi.ResourceOption,
) {
	passwordPlugin, _ := apiKeyReadWrite.ApplyT(func(key string) string {
		pp, _ := template.Render("./assets/roundcube/password.inc.php.j2", map[string]any{
			"mailname": mail.Mailname(*mailConfig.Main.Name),
			"apiToken": key,
		})
		return pp
	}).(pulumi.StringOutput)
	passwordPluginHash := google.WriteFileAndUpload(ctx, &storage.WriteFileAndUploadOptions{
		Name:       "roundcube_password.inc.php",
		Content:    passwordPlugin,
		OutputPath: "./outputs",
		BucketID:   config.BucketID,
		BucketPath: config.BucketPath,
		Labels:     config.CommonLabels(),
	}).
		ApplyT(func(_ any) string {
			hash, _ := file.Hash("./outputs/roundcube_password.inc.php")
			return *hash
		})
	passwordPluginCopy := installTask.ApplyT(func(install any) pulumi.ResourceOption {
		installer, _ := install.(pulumi.ResourceOption)
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			"remote-copy-roundcube-password-plugin-conf",
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset("./outputs/roundcube_password.inc.php"),
				RemotePath: pulumi.String("/opt/roundcube/www/plugins/password/config.inc.php"),
				Triggers:   pulumi.Array{passwordPluginHash},
				Connection: conn,
			},
			append(opts, installer)...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	pulumi.All(passwordPluginCopy, installTask).ApplyT(func(args []any) error {
		passwordCopy, _ := args[0].(pulumi.ResourceOption)
		installer, _ := args[1].(pulumi.ResourceOption)

		install.Postinstall(
			ctx,
			"roundcube",
			pulumi.Array{passwordPluginHash},
			conn,
			append(opts, passwordCopy, installer)...)
		return nil
	})
}
