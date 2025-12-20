package install

import (
	"fmt"

	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Prepare executes the preparation script for the given software on the remote server.
// ctx: Pulumi context.
// name: The name of the software (used to locate the preparation script).
// conn: The remote connection arguments.
// opts: Additional Pulumi resource options.
func Prepare(
	ctx *pulumi.Context,
	name string,
	conn *remote.ConnectionArgs,
	opts ...pulumi.ResourceOption,
) ([]pulumi.ResourceOption, error) {
	prepareFn, pErr := file.ReadContents(fmt.Sprintf("./assets/%s/prepare.sh", name))
	if pErr != nil {
		return nil, pErr
	}
	prepare, prepErr := remote.NewCommand(ctx, fmt.Sprintf("remote-command-prepare-%s", name), &remote.CommandArgs{
		Create:     pulumi.StringPtr(prepareFn),
		Connection: conn,
	}, opts...)
	if prepErr != nil {
		return nil, prepErr
	}
	return append(opts, pulumi.DependsOn([]pulumi.Resource{prepare})), nil
}
