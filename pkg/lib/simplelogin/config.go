package simplelogin

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/server"
	simpleloginConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/simplelogin"
	psqlModel "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/postgresql"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/aws/s3/bucket"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/random"
	"github.com/muhlba91/pulumi-shared-library/pkg/model/postgresql"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/aws/region"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/storage"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
)

// flaskSecretLength defines the length of the Flask secret for SimpleLogin.
const flaskSecretLength = 32

// createConfig creates the configuration file for SimpleLogin and necessary resources.
// ctx: Pulumi context.
// conn: SSH connection arguments to the remote server.
// postgresqlUsers: Map of PostgreSQL users needed for SimpleLogin.
// simpleloginConfig: Configuration for SimpleLogin installation.
// serverConfig: Configuration of the server where SimpleLogin is installed.
// opts: Additional Pulumi resource options.
func createConfig(ctx *pulumi.Context,
	conn *remote.ConnectionArgs,
	postgresqlUsers map[string]*pulumi.AnyOutput,
	simpleloginConfig *simpleloginConf.Config,
	serverConfig *server.Config,
	opts ...pulumi.ResourceOption,
) (pulumi.Output, pulumi.StringOutput) {
	flaskSecret, _ := random.CreatePassword(ctx, "password-simplelogin-flask-secret", &random.PasswordOptions{
		Length:  flaskSecretLength,
		Special: false,
	})

	s3Bucket, _ := bucket.Create(ctx, &bucket.CreateOptions{
		Name:   fmt.Sprintf("%s-simplelogin", config.GlobalName),
		Labels: config.CommonLabels(),
	})
	key := s3Bucket.Arn.ApplyT(func(arn string) iam.AccessKeyOutput {
		k, _ := createAWSUser(ctx, arn)
		return *k
	})

	envFile, _ := pulumi.All(postgresqlUsers["simplelogin"], flaskSecret.Password, s3Bucket.Bucket, config.PostgresqlConfig, key).ApplyT(func(args []any) pulumi.StringOutput {
		psqlUsers, _ := args[0].(*postgresql.UserData)
		flaskSecretPassword, _ := args[1].(string)
		bucketName, _ := args[2].(string)
		postgresqlConfig, _ := args[3].(*psqlModel.Config)
		accessKey, _ := args[4].(*iam.AccessKey)

		eFile, _ := psqlUsers.Password.ApplyT(func(postgresqlPasword string) pulumi.StringOutput {
			file, _ := pulumi.All(accessKey.ID(), accessKey.Secret).ApplyT(func(akArgs []any) string {
				accessKeyID, _ := akArgs[0].(pulumi.ID)
				secretAccessKey, _ := akArgs[1].(string)
				env, _ := template.Render("./assets/simplelogin/env.j2", map[string]any{
					"flaskSecret": flaskSecretPassword,
					"db": map[string]any{
						"uri":      fmt.Sprintf("postgresql://simplelogin:%s@%s:%d/simplelogin", postgresqlPasword, postgresqlConfig.Address, postgresqlConfig.Port), //nolint:nosprintfhostport // we need the full address here
						"host":     postgresqlConfig.Address,
						"port":     postgresqlConfig.Port,
						"database": "simplelogin",
						"user":     "simplelogin",
						"password": postgresqlPasword,
					},
					"aws": map[string]any{
						"bucket":          bucketName,
						"region":          region.GetOrDefault(ctx, &config.AWSDefaultRegion),
						"accessKeyId":     accessKeyID,
						"secretAccessKey": secretAccessKey,
					},
					"oidc": map[string]any{
						"wellKnownUrl": simpleloginConfig.OIDC.WellKnownURL,
						"clientId":     simpleloginConfig.OIDC.ClientID,
						"clientSecret": simpleloginConfig.OIDC.ClientSecret,
					},
					"domain": simpleloginConfig.Domain,
					"email": map[string]any{
						"domain": simpleloginConfig.Mail.Domain,
						"mx":     simpleloginConfig.Mail.MX,
						"relay":  serverConfig.IPv4,
					},
				})
				return env
			}).(pulumi.StringOutput)
			return file
		}).(pulumi.StringOutput)
		return eFile
	}).(pulumi.StringOutput)
	envFileHash, _ := storage.WriteFileAndUpload(ctx, &storage.WriteFileAndUploadArgs{
		Name:       "simplelogin_env",
		Content:    envFile,
		OutputPath: "./outputs",
		BucketID:   config.BucketID,
		BucketPath: config.BucketPath,
		Labels:     config.CommonLabels(),
	}).
		ApplyT(func(_ any) string {
			hash, _ := file.Hash("./outputs/simplelogin_env")
			return *hash
		}).(pulumi.StringOutput)
	envFileCopy := envFileHash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			"remote-copy-simplelogin-env",
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset("./outputs/simplelogin_env"),
				RemotePath: pulumi.String("/opt/simplelogin/env"),
				Triggers:   pulumi.Array{envFileHash},
				Connection: conn,
			},
			opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	return envFileCopy, envFileHash
}
