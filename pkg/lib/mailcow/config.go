package mailcow

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"
	mailConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	mcModel "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/mailcow"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/mail"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/storage"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
)

// createConfig creates the mailcow.conf file and uploads it to the remote server.
// It returns a Pulumi Output representing the copy operation and the hash of the configuration file.
// ctx: Pulumi context.
// conn: Remote connection arguments for SSH.
// ipv4Address: The public IPv4 address of the server.
// ipv6Address: The public IPv6 address of the server.
// secrets: Mailcow secrets needed for configuration.
// mailConfig: Mail configuration.
// dnsConfig: DNS configuration.
// opts: Additional Pulumi resource options.
func createConfig(ctx *pulumi.Context,
	conn *remote.ConnectionArgs,
	ipv4Address pulumi.StringOutput,
	ipv6Address pulumi.StringOutput,
	secrets *mcModel.Secrets,
	mailConfig *mailConf.Config,
	dnsConfig *dns.Config,
	opts ...pulumi.ResourceOption,
) (pulumi.Output, pulumi.StringOutput) {
	configFile, _ := pulumi.All(secrets.DBUserPassword, secrets.DBRootPassword, secrets.RedisPassword, secrets.APIKeyReadWrite, secrets.APIKeyRead, ipv4Address, ipv6Address).ApplyT(func(args []any) string {
		userPassword, _ := args[0].(string)
		rootPassword, _ := args[1].(string)
		redisPassword, _ := args[2].(string)
		apiKeyReadWrite, _ := args[3].(string)
		apiKeyRead, _ := args[4].(string)
		ipv4, _ := args[5].(string)
		ipv6, _ := args[6].(string)

		dc, _ := template.Render("./assets/mailcow/config/mailcow.conf.j2", map[string]any{
			"mailname": mail.Mailname(*mailConfig.Main.Name),
			"db": map[string]any{
				"auth": map[string]string{
					"user": userPassword,
					"root": rootPassword,
				},
			},
			"redis": map[string]string{
				"password": redisPassword,
			},
			"api": map[string]string{
				"readWrite": apiKeyReadWrite,
				"read":      apiKeyRead,
			},
			"ip": map[string]string{
				"v4": ipv4,
				"v6": ipv6,
			},
			"acme": map[string]string{
				"email": *dnsConfig.Email,
			},
		})
		return dc
	}).(pulumi.StringOutput)
	configFileHash, _ := storage.WriteFileAndUpload(ctx, &storage.WriteFileAndUploadArgs{
		Name:       "mailcow_mailcow.conf",
		Content:    configFile,
		OutputPath: "./outputs",
		BucketID:   config.BucketID,
		BucketPath: config.BucketPath,
		Labels:     config.CommonLabels(),
	}).
		ApplyT(func(_ any) string {
			hash, _ := file.Hash("./outputs/mailcow_mailcow.conf")
			return *hash
		}).(pulumi.StringOutput)
	configFileCopy := configFileHash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			"remote-copy-mailcow-conf",
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset("./outputs/mailcow_mailcow.conf"),
				RemotePath: pulumi.String("/opt/mailcow/mailcow.conf"),
				Triggers:   pulumi.Array{configFileHash},
				Connection: conn,
			},
			opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	return configFileCopy, configFileHash
}
