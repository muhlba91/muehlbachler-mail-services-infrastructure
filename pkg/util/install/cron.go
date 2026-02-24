package install

import (
	"fmt"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/google/project"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Cron executes the cron job setup for the given software on the remote server.
// ctx: Pulumi context.
// name: The name of the software (used to locate the cron job script).
// conn: The remote connection arguments.
// opts: Additional Pulumi resource options.
func Cron(
	ctx *pulumi.Context,
	name string,
	conn *remote.ConnectionArgs,
	opts ...pulumi.ResourceOption,
) ([]pulumi.Output, error) {
	backupFile, dcErr := template.Render(
		fmt.Sprintf("./assets/%s/cron/%s-backup.j2", name, name),
		map[string]any{
			"project": project.GetOrDefault(ctx, nil),
			"bucket": map[string]string{
				"id":   config.BackupBucketID,
				"path": config.BackupBucketPath,
			},
		},
	)
	if dcErr != nil {
		return nil, dcErr
	}
	backupFileHash := file.WritePulumi(fmt.Sprintf("./outputs/%s_backup", name), pulumi.String(backupFile)).
		ApplyT(func(_ string) string {
			hash, _ := file.Hash(fmt.Sprintf("./outputs/%s_backup", name))
			return *hash
		})
	backupFileCopy := backupFileHash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			fmt.Sprintf("remote-copy-%s-backup", name),
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset(fmt.Sprintf("./outputs/%s_backup", name)),
				RemotePath: pulumi.Sprintf("/bin/%s-backup", name),
				Triggers:   pulumi.Array{backupFileHash},
				Connection: conn,
			},
			opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	cronFileHash, shErr := file.Hash(fmt.Sprintf("./assets/%s/cron/cron", name))
	if shErr != nil {
		return nil, shErr
	}
	cronFileCopy, cfErr := remote.NewCopyToRemote(
		ctx,
		fmt.Sprintf("remote-copy-%s-cron", name),
		&remote.CopyToRemoteArgs{
			Source:     pulumi.NewFileAsset(fmt.Sprintf("./assets/%s/cron/cron", name)),
			RemotePath: pulumi.String(fmt.Sprintf("/etc/cron.d/%s", name)),
			Triggers:   pulumi.Array{pulumi.String(*cronFileHash)},
			Connection: conn,
		},
		opts...)
	if cfErr != nil {
		return nil, cfErr
	}
	opts = append(opts, pulumi.DependsOn([]pulumi.Resource{cronFileCopy}))

	cronInstallFn, ciErr := file.ReadContents(fmt.Sprintf("./assets/%s/cron/install.sh", name))
	if ciErr != nil {
		return nil, ciErr
	}
	cronInstall, ciErr := remote.NewCommand(
		ctx,
		fmt.Sprintf("remote-command-install-%s-cron", name),
		&remote.CommandArgs{
			Create:     pulumi.StringPtr(cronInstallFn),
			Update:     pulumi.StringPtr(cronInstallFn),
			Triggers:   pulumi.Array{pulumi.String(*cronFileHash), backupFileHash},
			Connection: conn,
		},
		opts...)
	if ciErr != nil {
		return nil, ciErr
	}

	return []pulumi.Output{
		backupFileCopy,
		pulumi.ToOutput(pulumi.DependsOn([]pulumi.Resource{cronInstall})),
	}, nil
}
