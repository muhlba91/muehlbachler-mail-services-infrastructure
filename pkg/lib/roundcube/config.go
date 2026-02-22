package roundcube

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	mailConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	rcConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/roundcube"
	psqlModel "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/postgresql"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/file"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/mail"
	"github.com/muhlba91/pulumi-shared-library/pkg/model/postgresql"
	fileUtil "github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
)

// createConfig creates and uploads the necessary configuration files for Roundcube.
// ctx: Pulumi context.
// conn: SSH connection arguments.
// postgresqlUsers: Map of PostgreSQL users needed for Roundcube.
// roundcubeConfig: Configuration for Roundcube installation.
// mailConfig: Configuration for mail services.
// opts: Additional Pulumi resource options.
func createConfig(ctx *pulumi.Context,
	conn *remote.ConnectionArgs,
	postgresqlUsers map[string]*pulumi.AnyOutput,
	roundcubeConfig *rcConf.Config,
	mailConfig *mailConf.Config,
	opts ...pulumi.ResourceOption,
) (*remote.CopyToRemote, pulumi.Output, pulumi.StringOutput) {
	nginxConfHash, _ := fileUtil.Hash("./assets/roundcube/nginx.conf")
	nginxConfCopy, _ := remote.NewCopyToRemote(
		ctx,
		"remote-copy-roundcube-nginx",
		&remote.CopyToRemoteArgs{
			Source:     pulumi.NewFileAsset("./assets/roundcube/nginx.conf"),
			RemotePath: pulumi.String("/opt/roundcube/nginx.conf"),
			Triggers:   pulumi.Array{pulumi.String(*nginxConfHash)},
			Connection: conn,
		},
		opts...)
	opts = append(opts, pulumi.DependsOn([]pulumi.Resource{nginxConfCopy}))

	configFile, _ := pulumi.All(postgresqlUsers["roundcube"], config.PostgresqlConfig).ApplyT(func(args []any) pulumi.StringOutput {
		psqlUser, _ := args[0].(*postgresql.UserData)
		postgresqlConfig, _ := args[1].(*psqlModel.Config)

		dc, _ := psqlUser.Password.ApplyT(func(postgresqlPassword string) string {
			tpl, _ := template.Render("./assets/roundcube/custom.inc.php.j2", map[string]any{
				"mailname": mail.Mailname(*mailConfig.Main.Name),
				"domain":   roundcubeConfig.Domain.Name,
				"db": map[string]string{
					"host":     postgresqlConfig.Address,
					"database": "roundcube",
					"user":     "roundcube",
					"password": postgresqlPassword,
				},
			})
			return tpl
		}).(pulumi.StringOutput)
		return dc
	}).(pulumi.StringOutput)
	configFileHash, _ := file.WriteAndUpload(ctx, "roundcube_custom.inc.php", configFile).
		ApplyT(func(_ any) string {
			hash, _ := fileUtil.Hash("./outputs/roundcube_custom.inc.php")
			return *hash
		}).(pulumi.StringOutput)
	configFileCopy := configFileHash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			"remote-copy-roundcube-custom-conf",
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset("./outputs/roundcube_custom.inc.php"),
				RemotePath: pulumi.String("/opt/roundcube/config/custom.inc.php"),
				Triggers:   pulumi.Array{configFileHash},
				Connection: conn,
			},
			opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	return nginxConfCopy, configFileCopy, configFileHash
}
