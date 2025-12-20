package postgresql

import (
	"encoding/json"
	"fmt"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/database"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/postgresql/user"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	psqlModel "github.com/muhlba91/pulumi-shared-library/pkg/model/postgresql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createUsers creates PostgreSQL users based on the provided database configuration.
// ctx: The Pulumi context for resource management.
// databaseConfig: The configuration containing database user details.
// provider: The Pulumi PostgreSQL provider to use for resource creation.
func createUsers(
	ctx *pulumi.Context,
	databaseConfig *database.Config,
	provider pulumi.AnyOutput,
) map[string]*pulumi.AnyOutput {
	users := make(map[string]*pulumi.AnyOutput)

	for _, username := range databaseConfig.Users {
		pgUser, _ := provider.ApplyT(func(prov any) *psqlModel.UserData {
			pgProv, _ := prov.(pulumi.ResourceOrInvokeOption)
			pgu, _ := user.Create(ctx, &user.CreateOptions{
				Username: username,
				PulumiOptions: []pulumi.ResourceOption{
					pgProv,
				},
			})

			secretValue, _ := pulumi.All(pgu.User.Name, pgu.Password).ApplyT(func(args []any) string {
				username := args[0].(string)
				userPassword := args[1].(string)
				val, _ := json.Marshal(map[string]any{
					"user":     username,
					"password": userPassword,
				})
				return string(val)
			}).(pulumi.StringOutput)
			_, _ = secret.Write(ctx, &secret.WriteArgs{
				Path:  config.GlobalName,
				Key:   fmt.Sprintf("postgresql-user-%s", username),
				Value: secretValue,
			})

			return pgu
		}).(pulumi.AnyOutput)

		users[username] = &pgUser
	}

	return users
}
