package journal

import (
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/install"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Install journal configuration on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// dependsOn: Pulumi resource option to specify dependencies.
func Install(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	dependsOn pulumi.ResourceOrInvokeOption,
) error {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	confHash, _ := file.Hash("./assets/journal/journald.conf")
	rCopy, cErr := remote.NewCopyToRemote(ctx, "remote-copy-journald-conf",
		&remote.CopyToRemoteArgs{
			Source:     pulumi.NewFileAsset("./assets/journal/journald.conf"),
			RemotePath: pulumi.String("/etc/systemd/journald.conf"),
			Triggers:   pulumi.Array{pulumi.String(*confHash)},
			Connection: conn,
		},
		opts...)
	if cErr != nil {
		return cErr
	}

	install.Postinstall(
		ctx,
		"journal",
		pulumi.Array{},
		conn,
		append(opts, pulumi.DependsOn([]pulumi.Resource{rCopy}))...)
	return nil
}
