package postgresql

import (
	"encoding/json"
	"fmt"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	dbModel "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/database"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/postgresql/database"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	"github.com/muhlba91/pulumi-shared-library/pkg/model/postgresql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createDatabases creates PostgreSQL databases based on the provided database configuration.
// ctx: The Pulumi context for resource management.
// databaseConfig: The configuration containing database details.
// users: A map of database users to be assigned as owners of the databases.
// provider: The Pulumi PostgreSQL provider resource.
func createDatabases(
	ctx *pulumi.Context,
	databaseConfig *dbModel.Config,
	users map[string]*pulumi.AnyOutput,
	provider pulumi.AnyOutput,
) {
	provider.ApplyT(func(prov any) error {
		pgProv, _ := prov.(pulumi.ResourceOrInvokeOption)
		for db, owner := range databaseConfig.Database {
			_ = users[owner].ApplyT(func(user any) error {
				usr, _ := user.(*postgresql.UserData)
				_, _ = database.Create(ctx, &database.CreateOptions{
					Name:  db,
					Owner: usr,
					PulumiOptions: []pulumi.ResourceOption{
						pgProv,
					},
				})

				secretValue, _ := json.Marshal(map[string]any{
					"name": db,
				})
				_, _ = secret.Write(ctx, &secret.WriteArgs{
					Path:  config.GlobalName,
					Key:   fmt.Sprintf("postgresql-database-%s", db),
					Value: pulumi.String(secretValue),
				})

				return nil
			})
		}
		return nil
	})
}
