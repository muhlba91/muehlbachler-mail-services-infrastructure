package docker

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Install Docker on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// dependsOn: Pulumi resource option to specify dependencies.
func Install(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*remote.Command, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	daemonJSON, dErr := file.ReadContents("./assets/docker/daemon.json")
	if dErr != nil {
		return nil, dErr
	}
	createFn, cfErr := template.Render("./assets/docker/install.sh", map[string]any{
		"daemonJson": daemonJSON,
	})
	if cfErr != nil {
		return nil, cfErr
	}
	return remote.NewCommand(ctx, "remote-command-install-docker", &remote.CommandArgs{
		Create:     pulumi.StringPtr(createFn),
		Connection: conn,
	}, dependsOn)
}
