package config

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/database"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/network"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/ntfy"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/roundcube"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/server"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/simplelogin"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/postgresql"
)

//nolint:gochecknoglobals // global configuration is acceptable here
var (
	// Environment holds the current deployment environment (e.g., dev, staging, prod).
	Environment string
	// GlobalName is a constant name used across resources.
	GlobalName = "mail-services"
	// GlobalNameShort is a constant short name used across resources.
	GlobalNameShort = "mail"
	// AWSDefaultRegion is the default AWS region for deployments.
	AWSDefaultRegion = "eu-west-1"
	// BucketPath is the path within the buckets for this project.
	BucketPath string
	// BucketID is the ID of the main storage bucket.
	BucketID string
	// BackupBucketID is the ID of the backup storage bucket.
	BackupBucketID string
	// PostgresqlConfig holds the configuration for PostgreSQL access.
	PostgresqlConfig *pulumi.AnyOutput
)

// LoadConfig loads the configuration for the given Pulumi context.
// ctx: The Pulumi context.
func LoadConfig(
	ctx *pulumi.Context,
) (*dns.Config, *network.Config, *server.Config, *mail.Config, *simplelogin.Config, *roundcube.Config, *ntfy.Config, *database.Config, error) {
	Environment = ctx.Stack()

	cfg := config.New(ctx, "")

	BucketID = cfg.Require("bucketId")
	BackupBucketID = cfg.Require("backupBucketId")
	BucketPath = fmt.Sprintf("%s/%s", GlobalName, Environment)

	var dnsConfig dns.Config
	cfg.RequireObject("dns", &dnsConfig)

	var networkConfig network.Config
	cfg.RequireObject("network", &networkConfig)

	var serverConfig server.Config
	cfg.RequireObject("server", &serverConfig)

	var mailConfig mail.Config
	cfg.RequireObject("mail", &mailConfig)

	var simpleloginConfig simplelogin.Config
	cfg.RequireObject("simplelogin", &simpleloginConfig)

	var roundcubeConfig roundcube.Config
	cfg.RequireObject("roundcube", &roundcubeConfig)

	var ntfyConfig ntfy.Config
	cfg.RequireObject("ntfy", &ntfyConfig)

	var databaseConfig database.Config
	cfg.RequireObject("database", &databaseConfig)

	sharedServicesStack, sErr := pulumi.NewStackReference(
		ctx,
		fmt.Sprintf("%s/%s/%s", ctx.Organization(), "muehlbachler-shared-services", Environment),
		nil,
	)
	if sErr != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, sErr
	}
	sharedServicesStackAws := sharedServicesStack.GetOutput(pulumi.String("aws"))
	psqlConfig, _ := sharedServicesStackAws.ApplyT(func(awsOutput any) *postgresql.Config {
		psqlConn, _ := awsOutput.(map[string]any)["postgresql"].(map[string]any)
		address, _ := psqlConn["address"].(string)
		port, _ := psqlConn["port"].(float64)
		username, _ := psqlConn["username"].(string)
		password, _ := psqlConn["password"].(string)

		return &postgresql.Config{
			Address:  address,
			Port:     int(port),
			Username: username,
			Password: password,
		}
	}).(pulumi.AnyOutput)
	PostgresqlConfig = &psqlConfig

	return &dnsConfig, &networkConfig, &serverConfig, &mailConfig, &simpleloginConfig, &roundcubeConfig, &ntfyConfig, &databaseConfig, nil
}

// CommonLabels returns a map of common labels to be used across resources.
func CommonLabels() map[string]string {
	return map[string]string{
		"environment": Environment,
		"purpose":     GlobalName,
	}
}
