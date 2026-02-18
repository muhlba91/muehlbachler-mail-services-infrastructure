package postgresql

import (
	"encoding/json"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/database"
	psqlModel "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/postgresql"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	"github.com/pulumi/pulumi-postgresql/sdk/v3/go/postgresql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Create creates PostgreSQL users and databases based on the provided database configuration.
// ctx: The Pulumi context for resource management.
// databaseConfig: The configuration containing database user details.
func Create(ctx *pulumi.Context, databaseConfig *database.Config) (map[string]*pulumi.AnyOutput, error) {
	pgProvider, _ := config.PostgresqlConfig.ApplyT(func(conf any) pulumi.ResourceOrInvokeOption {
		psqlConf := conf.(*psqlModel.Config)

		provider, provErr := postgresql.NewProvider(ctx, "postgresql", &postgresql.ProviderArgs{
			Host:      pulumi.String(psqlConf.Address),
			Port:      pulumi.Int(psqlConf.Port),
			Username:  pulumi.String(psqlConf.Username),
			Password:  pulumi.String(psqlConf.Password),
			Superuser: pulumi.Bool(false),
		})
		if provErr != nil {
			return nil
		}

		connectionSecret, _ := json.Marshal(map[string]any{
			"port": psqlConf.Port,
			"host": psqlConf.Address,
		})
		_, _ = secret.Create(ctx, &secret.CreateOptions{
			Path:  config.GlobalName,
			Key:   "postgresql-connection",
			Value: pulumi.String(string(connectionSecret)),
		})

		return pulumi.Provider(provider)
	}).(pulumi.AnyOutput)

	users := createUsers(ctx, databaseConfig, pgProvider)
	createDatabases(ctx, databaseConfig, users, pgProvider)

	return users, nil
}
