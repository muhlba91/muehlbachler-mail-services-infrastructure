package install

import (
	"fmt"

	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// SystemDService creates a systemd service file for the given software on the remote server.
// ctx: Pulumi context.
// name: The name of the software (used to locate the service file).
// conn: The remote connection arguments.
// opts: Additional Pulumi resource options.
func SystemDService(
	ctx *pulumi.Context,
	name string,
	conn *remote.ConnectionArgs,
	opts ...pulumi.ResourceOption,
) ([]pulumi.ResourceOption, *string, error) {
	systemdServiceHash, shErr := file.Hash(fmt.Sprintf("./assets/%s/%s.service", name, name))
	if shErr != nil {
		return nil, nil, shErr
	}
	systemdServiceCopy, tyErr := remote.NewCopyToRemote(
		ctx,
		fmt.Sprintf("remote-copy-%s-service", name),
		&remote.CopyToRemoteArgs{
			Source:     pulumi.NewFileAsset(fmt.Sprintf("./assets/%s/%s.service", name, name)),
			RemotePath: pulumi.Sprintf("/etc/systemd/system/%s.service", name),
			Triggers:   pulumi.Array{pulumi.String(*systemdServiceHash)},
			Connection: conn,
		},
		opts...)
	if tyErr != nil {
		return nil, nil, tyErr
	}
	opts = append(opts, pulumi.DependsOn([]pulumi.Resource{systemdServiceCopy}))
	return opts, systemdServiceHash, nil
}
