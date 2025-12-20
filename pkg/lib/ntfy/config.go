package ntfy

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	ntfyConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/ntfy"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/storage"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
)

// createConfig creates the Ntfy configuration file on the remote server.
// ctx: Pulumi context.
// conn: The remote connection arguments.
// ntfyConfig: Ntfy configuration.
// opts: Additional Pulumi resource options.
func createConfig(ctx *pulumi.Context,
	conn *remote.ConnectionArgs,
	ntfyConfig *ntfyConf.Config,
	opts ...pulumi.ResourceOption,
) (pulumi.Output, pulumi.StringOutput) {
	configFile, _ := template.Render("./assets/ntfy/server.yml.j2", map[string]any{
		"domain": ntfyConfig.Domain.Name,
	})
	configFileHash, _ := storage.WriteFileAndUpload(ctx, &storage.WriteFileAndUploadArgs{
		Name:       "ntfy_server.yml",
		Content:    pulumi.String(configFile),
		OutputPath: "./outputs",
		BucketID:   config.BucketID,
		BucketPath: config.BucketPath,
		Labels:     config.CommonLabels(),
	}).
		ApplyT(func(_ any) string {
			hash, _ := file.Hash("./outputs/ntfy_server.yml")
			return *hash
		}).(pulumi.StringOutput)
	configFileCopy := configFileHash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			"remote-copy-ntfy-server-yml",
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset("./outputs/ntfy_server.yml"),
				RemotePath: pulumi.String("/opt/ntfy/config/server.yml"),
				Triggers:   pulumi.Array{configFileHash},
				Connection: conn,
			},
			opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	return configFileCopy, configFileHash
}
