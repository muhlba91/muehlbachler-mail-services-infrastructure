package mailcow

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/install"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
)

// Install Mailcow on the remote server via SSH and create necessary resources.
// ctx: Pulumi context.
// conn: SSH connection arguments.
// installTask: The installation task output to depend on.
// opts: Additional Pulumi resource options.
func postinstall(ctx *pulumi.Context,
	conn *remote.ConnectionArgs,
	installTask pulumi.Output,
	opts ...pulumi.ResourceOption,
) {
	bodyChecksHash, _ := file.Hash("./assets/mailcow/config/body_checks.pcre")
	bodyChecksCopy := installTask.ApplyT(func(install any) pulumi.ResourceOrInvokeOption {
		installer, _ := install.(pulumi.ResourceOption)
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			"remote-copy-mailcow-postfix-body-checks",
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset("./assets/mailcow/config/body_checks.pcre"),
				RemotePath: pulumi.String("/opt/mailcow/data/conf/postfix/body_checks.pcre"),
				Triggers:   pulumi.Array{pulumi.String(*bodyChecksHash)},
				Connection: conn,
			},
			append(opts, installer)...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	clientHeadersHash, _ := file.Hash("./assets/mailcow/config/client_headers.pcre")
	clientHeadersCopy := installTask.ApplyT(func(install any) pulumi.ResourceOrInvokeOption {
		installer, _ := install.(pulumi.ResourceOption)
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			"remote-copy-mailcow-postfix-client-headers",
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset("./assets/mailcow/config/client_headers.pcre"),
				RemotePath: pulumi.String("/opt/mailcow/data/conf/postfix/client_headers.pcre"),
				Triggers:   pulumi.Array{pulumi.String(*clientHeadersHash)},
				Connection: conn,
			},
			append(opts, installer)...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	postfixExtraHash, _ := file.Hash("./assets/mailcow/config/extra.cf")
	postfixExtraCopy := installTask.ApplyT(func(install any) pulumi.ResourceOrInvokeOption {
		installer, _ := install.(pulumi.ResourceOption)
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			"remote-copy-mailcow-postfix-extra",
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset("./assets/mailcow/config/extra.cf"),
				RemotePath: pulumi.String("/opt/mailcow/data/conf/postfix/extra.cf"),
				Triggers:   pulumi.Array{pulumi.String(*postfixExtraHash)},
				Connection: conn,
			},
			append(opts, installer)...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	pulumi.All(postfixExtraCopy, bodyChecksCopy, clientHeadersCopy, installTask).
		ApplyT(func(args []any) error {
			postfixExtra, _ := args[0].(pulumi.ResourceOption)
			bodyChecks, _ := args[1].(pulumi.ResourceOption)
			clientHeaders, _ := args[2].(pulumi.ResourceOption)
			installer, _ := args[3].(pulumi.ResourceOption)

			install.Postinstall(ctx, "mailcow", pulumi.Array{
				pulumi.String(*postfixExtraHash),
				pulumi.String(*bodyChecksHash),
				pulumi.String(*clientHeadersHash),
			}, conn, append(opts, postfixExtra, bodyChecks, clientHeaders, installer)...)
			return nil
		})
}
