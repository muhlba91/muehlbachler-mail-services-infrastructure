package mailcow

import (
	"encoding/json"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	mcModel "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/mailcow"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/random"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateSecrets generates all required secrets for mailcow.
// ctx: The Pulumi context.
func CreateSecrets(ctx *pulumi.Context) (*mcModel.Secrets, error) {
	dbUserPassword, mcdbErr := random.CreatePassword(
		ctx,
		"password-db-user",
		&random.PasswordOptions{
			Special: false,
		},
	)
	if mcdbErr != nil {
		return nil, mcdbErr
	}
	dbRootPassword, mcdrErr := random.CreatePassword(
		ctx,
		"password-db-root",
		&random.PasswordOptions{
			Special: false,
		},
	)
	if mcdrErr != nil {
		return nil, mcdrErr
	}
	redisPassword, mcrErr := random.CreatePassword(
		ctx,
		"password-redis",
		&random.PasswordOptions{
			Special: false,
		},
	)
	if mcrErr != nil {
		return nil, mcrErr
	}

	apiKeyReadWrite, mcaErr := random.CreatePassword(
		ctx,
		"password-mailcow-api-read-write",
		&random.PasswordOptions{
			Special: false,
		},
	)
	if mcaErr != nil {
		return nil, mcaErr
	}
	apiKeyRead, mcarErr := random.CreatePassword(
		ctx,
		"password-mailcow-api-read-only",
		&random.PasswordOptions{
			Special: false,
		},
	)
	if mcarErr != nil {
		return nil, mcarErr
	}

	mailcowAPIKeysSecret, _ := pulumi.All(apiKeyReadWrite.Password, apiKeyRead.Password).ApplyT(func(args []any) string {
		readWrite, _ := args[0].(string)
		read, _ := args[1].(string)
		secret, _ := json.Marshal(map[string]string{
			"read_write": readWrite,
			"read":       read,
		})
		return string(secret)
	}).(pulumi.StringOutput)
	_, _sErr := secret.Create(ctx, &secret.CreateOptions{
		Path:  config.GlobalName,
		Key:   "mailcow-api",
		Value: mailcowAPIKeysSecret,
	})
	if _sErr != nil {
		return nil, _sErr
	}

	return &mcModel.Secrets{
		DBUserPassword:  dbUserPassword.Password,
		DBRootPassword:  dbRootPassword.Password,
		RedisPassword:   redisPassword.Password,
		APIKeyReadWrite: apiKeyReadWrite.Password,
		APIKeyRead:      apiKeyRead.Password,
	}, nil
}
