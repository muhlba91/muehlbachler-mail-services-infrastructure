package install

import (
	"fmt"

	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// DockerCompose creates a docker-compose file for the given software on the remote server.
// ctx: Pulumi context.
// name: The name of the software (used to locate the service file).
// content: The docker-compose content to be written.
// override: Whether this is an override file.
// conn: The remote connection arguments.
// opts: Additional Pulumi resource options.
func DockerCompose(
	ctx *pulumi.Context,
	name string,
	content pulumi.StringInput,
	override bool,
	conn *remote.ConnectionArgs,
	opts ...pulumi.ResourceOption,
) (pulumi.ResourceOption, *pulumi.StringOutput, error) {
	filename := "docker-compose.yml"
	if override {
		filename = "docker-compose.override.yml"
	}

	dockerComposeHash, _ := file.WritePulumi(fmt.Sprintf("./outputs/%s_%s", name, filename), content).
		ApplyT(func(_ string) string {
			hash, _ := file.Hash(fmt.Sprintf("./outputs/%s_%s", name, filename))
			return *hash
		}).(pulumi.StringOutput)
	dockerComposeCopy, tyErr := remote.NewCopyToRemote(
		ctx,
		fmt.Sprintf("remote-copy-%s-docker-compose", name),
		&remote.CopyToRemoteArgs{
			Source:     pulumi.NewFileAsset(fmt.Sprintf("./outputs/%s_%s", name, filename)),
			RemotePath: pulumi.Sprintf("/opt/%s/%s", name, filename),
			Triggers:   pulumi.Array{dockerComposeHash},
			Connection: conn,
		},
		opts...)
	if tyErr != nil {
		return nil, nil, tyErr
	}
	return pulumi.DependsOn([]pulumi.Resource{dockerComposeCopy}), &dockerComposeHash, nil
}
