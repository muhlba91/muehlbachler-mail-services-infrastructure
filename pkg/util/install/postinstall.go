package install

import (
	"fmt"

	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Postinstall executes the post-installation script for the given software on the remote server.
// ctx: Pulumi context.
// name: The name of the software (used to locate the preparation script).
// conn: The remote connection arguments.
// opts: Additional Pulumi resource options.
func Postinstall(
	ctx *pulumi.Context,
	name string,
	triggers pulumi.Array,
	conn *remote.ConnectionArgs,
	opts ...pulumi.ResourceOption,
) []pulumi.ResourceOption {
	postinstallFn, _ := file.ReadContents(fmt.Sprintf("./assets/%s/postinstall.sh", name))
	postinstall, _ := remote.NewCommand(ctx, fmt.Sprintf("remote-command-postinstall-%s", name), &remote.CommandArgs{
		Create:     pulumi.StringPtr(postinstallFn),
		Update:     pulumi.StringPtr(postinstallFn),
		Triggers:   triggers,
		Connection: conn,
	}, opts...)
	return append(opts, pulumi.DependsOn([]pulumi.Resource{postinstall}))
}
