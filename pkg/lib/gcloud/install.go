package gcloud

import (
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/install"
	"github.com/muhlba91/pulumi-shared-library/pkg/model/google/iam/serviceaccount"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/encoding"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Install gcloud on the remote server via SSH.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// serviceAccount: The Google service account to use for authentication.
// dependsOn: Pulumi resource option to specify dependencies.
func Install(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	serviceAccount *serviceaccount.User,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*remote.Command, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "gcloud", conn, opts...)
	if prepErr != nil {
		return nil, prepErr
	}

	privateKey, _ := serviceAccount.Key.PrivateKey.ApplyT(func(key string) string {
		decKey, _ := encoding.B64Decode(key)
		return decKey
	}).(pulumi.StringOutput)
	gcpCredentialsHash := file.WritePulumi("./outputs/google_credentials.json", privateKey).
		ApplyT(func(_ string) string {
			hash, _ := file.Hash("./outputs/google_credentials.json")
			return *hash
		})
	gcpCredentialsCopy := gcpCredentialsHash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			"remote-copy-gcloud-service-account",
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset("./outputs/google_credentials.json"),
				RemotePath: pulumi.String("/opt/google/credentials.json"),
				Triggers:   pulumi.Array{gcpCredentialsHash},
				Connection: conn,
			},
			opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	installFn, iErr := file.ReadContents("./assets/gcloud/install.sh")
	if iErr != nil {
		return nil, iErr
	}
	return remote.NewCommand(ctx, "remote-command-install-gcloud", &remote.CommandArgs{
		Create:     pulumi.StringPtr(installFn),
		Update:     pulumi.StringPtr(installFn),
		Triggers:   pulumi.Array{gcpCredentialsHash},
		Connection: conn,
	}, append(opts, install.CollectResourceOptions([]pulumi.Output{gcpCredentialsCopy})...)...)
}
