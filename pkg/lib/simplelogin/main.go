package simplelogin

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"
	mailConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/server"
	simpleloginConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/simplelogin"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/dkim"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/install"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/random"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/template"
)

// databaseName is the name of the PostgreSQL database used by SimpleLogin.
const databaseName = "simplelogin"

// postgresPasswordLength defines the length of the PostgreSQL password for SimpleLogin.
const postgresPasswordLength = 32

// Install SimpleLogin on the remote server via SSH and create necessary resources.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// postgresqlUsers: Map of PostgreSQL users needed for SimpleLogin.
// simpleloginConfig: Configuration for SimpleLogin installation.
// serverConfig: Configuration of the server where SimpleLogin is installed.
// dependsOn: List of Pulumi resources that this installation depends on.
//
//nolint:funlen // this is a long function, but it's necessary for the installation process
func Install(ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	simpleloginConfig *simpleloginConf.Config,
	serverConfig *server.Config,
	mailConfig *mailConf.Config,
	dnsConfig *dns.Config,
	dependsOn pulumi.ResourceOrInvokeOption,
) (*dkim.Data, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	opts := []pulumi.ResourceOption{dependsOn}

	opts, prepErr := install.Prepare(ctx, "simplelogin", conn, opts...)
	if prepErr != nil {
		return nil, prepErr
	}

	// postgres password
	postgresqlPassword := createPostgresPassword(ctx)

	dockerCompose, _ := postgresqlPassword.ApplyT(func(pgPass string) string {
		tpl, _ := template.Render("./assets/simplelogin/docker-compose.yml.j2", map[string]any{
			//nolint:goconst // intentional duplication of "domain" key for better structure in the template
			"domain": simpleloginConfig.Domain,
			"db": map[string]any{
				"database": databaseName,
				"user":     databaseName,
				//nolint:goconst // intentional duplication of "password" key for better structure
				"password": pgPass,
			},
		})
		return tpl
	}).(pulumi.StringOutput)
	dockerComposeCopy, dockerComposeHash, dcErr := install.DockerCompose(
		ctx,
		"simplelogin",
		dockerCompose,
		false,
		conn,
		opts...)
	if dcErr != nil {
		return nil, dcErr
	}

	dkimKey, dkimKeyCopy, dkErr := createDKIMConfig(ctx, conn, simpleloginConfig, mailConfig, dnsConfig, opts...)
	if dkErr != nil {
		return nil, dkErr
	}
	envFileCopy, envFileHash := createConfig(
		ctx,
		conn,
		postgresqlPassword,
		simpleloginConfig,
		serverConfig,
		opts...)

	_, cronErr := install.Cron(ctx, "simplelogin", conn, opts...)
	if cronErr != nil {
		return nil, cronErr
	}

	opts, systemdServiceHash, shErr := install.SystemDService(ctx, "simplelogin", conn, opts...)
	if shErr != nil {
		return nil, shErr
	}

	simpleloginVersion := install.Version("./outputs/simplelogin_docker-compose.yml", "app", dockerComposeHash)

	initShHash, ishErr := file.Hash("./assets/simplelogin/init.sh")
	if ishErr != nil {
		return nil, ishErr
	}
	initShCopy, ishcErr := remote.NewCopyToRemote(
		ctx,
		"remote-copy-simplelogin-init-sh",
		&remote.CopyToRemoteArgs{
			Source:     pulumi.NewFileAsset("./assets/simplelogin/init.sh"),
			RemotePath: pulumi.String("/opt/simplelogin/init.sh"),
			Triggers:   pulumi.Array{pulumi.String(*initShHash)},
			Connection: conn,
		},
		opts...)
	if ishcErr != nil {
		return nil, ishcErr
	}

	installFn, _ := simpleloginVersion.ApplyT(func(version string) string {
		ic, _ := template.Render("./assets/simplelogin/install.sh.j2", map[string]any{
			"version": version,
		})
		return ic
	}).(pulumi.StringOutput)
	_ = pulumi.All(dkimKeyCopy, envFileCopy, initShCopy, dockerComposeCopy).
		ApplyT(func(args []any) pulumi.ResourceOption {
			dkimCopy, _ := args[0].(pulumi.ResourceOption)
			envCopy, _ := args[1].(pulumi.ResourceOption)
			initCopy, _ := args[2].(pulumi.ResourceOption)
			dockerCopy, _ := args[3].(pulumi.ResourceOption)

			cmd, _ := remote.NewCommand(
				ctx,
				"remote-command-install-simplelogin",
				&remote.CommandArgs{
					Create: installFn,
					Update: installFn,
					Triggers: pulumi.Array{
						pulumi.String(*systemdServiceHash),
						dockerComposeHash,
						envFileHash,
						pulumi.String(*initShHash),
						simpleloginVersion,
					},
					Connection: conn,
				},
				append(opts, dkimCopy, envCopy, initCopy, dockerCopy)...)
			return pulumi.DependsOn([]pulumi.Resource{cmd})
		})

	return dkimKey, nil
}

// createPostgresPassword generates a random password for the PostgreSQL user and stores it in a secret.
// ctx: Pulumi context.
func createPostgresPassword(ctx *pulumi.Context) pulumi.StringOutput {
	postgresqlPassword, _ := random.CreatePassword(ctx, "password-pg-user-simplelogin", &random.PasswordOptions{
		Length:  postgresPasswordLength,
		Special: false,
	})
	secretValue, _ := postgresqlPassword.Password.ApplyT(func(pgPass string) string {
		val, _ := json.Marshal(map[string]any{
			"password": pgPass,
		})
		return string(val)
	}).(pulumi.StringOutput)
	_, _ = secret.Create(ctx, &secret.CreateOptions{
		Path:  config.GlobalName,
		Key:   fmt.Sprintf("postgresql-user-%s", databaseName),
		Value: secretValue,
	})

	return postgresqlPassword.Password
}
