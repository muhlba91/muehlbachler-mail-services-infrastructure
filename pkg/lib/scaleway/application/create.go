package application

import (
	"encoding/json"
	"fmt"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/scaleway"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/scaleway/iam/policy"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	smodel "github.com/muhlba91/pulumi-shared-library/pkg/model/scaleway/iam/application"
	slApplication "github.com/muhlba91/pulumi-shared-library/pkg/util/scaleway/iam/application"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/iam"
	"github.com/rs/zerolog/log"
)

const (
	domainPermission = "DomainsDNSFullAccess"
	bucketPermission = "ObjectStorageFullAccess"
)

// Create a Scaleway application with necessary IAM roles.
// ctx: Pulumi context for resource management.
// scalewayConfig: Configuration for Scaleway settings, including project information for IAM role assignment.
func Create(ctx *pulumi.Context, scalewayConfig *scaleway.Config) (*smodel.Application, error) {
	resourceName := fmt.Sprintf("%s-%s", config.GlobalName, config.Environment)

	app, err := slApplication.CreateApplication(ctx, &slApplication.CreateOptions{
		Name:             resourceName,
		DefaultProjectID: pulumi.StringPtrFromPtr(scalewayConfig.Project),
	})
	if err != nil {
		return nil, err
	}

	rules := []iam.PolicyRuleInput{
		&iam.PolicyRuleArgs{
			ProjectIds:         pulumi.ToStringArray([]string{*scalewayConfig.Project}),
			PermissionSetNames: pulumi.ToStringArray([]string{bucketPermission}),
		},
		&iam.PolicyRuleArgs{
			ProjectIds:         pulumi.ToStringArray([]string{*scalewayConfig.DNSProject}),
			PermissionSetNames: pulumi.ToStringArray([]string{domainPermission}),
		},
	}

	_, errPolicy := policy.Create(ctx, resourceName, &policy.CreateOptions{
		Name: pulumi.Sprintf("scw-iam-policy-%s", resourceName),
		Description: pulumi.Sprintf(
			"Policy for the %s: %s",
			config.GlobalName,
			config.Environment,
		),
		Rules:         rules,
		ApplicationID: app.Application.ID(),
	})
	if errPolicy != nil {
		log.Error().
			Err(errPolicy).
			Msgf("[buckets][scaleway][application] failed to create IAM policy for %s", resourceName)
	}

	vaultValue, _ := (pulumi.All(app.Key.AccessKey, app.Key.SecretKey).ApplyT(func(args []any) string {
		accessKey, ok := args[0].(string)
		if !ok {
			log.Error().Msgf("[buckets][scaleway][application] failed to cast access key for %s", resourceName)
		}
		secretKey, ok := args[1].(string)
		if !ok {
			log.Error().Msgf("[buckets][scaleway][application] failed to cast secret key for %s", resourceName)
		}
		data, errMarshal := json.Marshal(map[string]string{
			"access_key":      accessKey,
			"secret_key":      secretKey,
			"organization_id": scalewayConfig.OrganizationID,
			"project_id":      *scalewayConfig.Project,
			"region":          config.ScalewayDefaultRegion,
			"bucket":          config.BackupBucketID,
		})
		if errMarshal != nil {
			log.Error().Err(errMarshal).Msgf("[buckets][scaleway][application][vault] failed to marshal credentials for %s", resourceName)
		}
		return string(data)
	})).(pulumi.StringOutput)

	_, errVault := secret.Create(ctx, &secret.CreateOptions{
		Key:   "scaleway",
		Value: vaultValue,
		Path:  config.GlobalName,
	})
	if errVault != nil {
		log.Error().
			Err(errVault).
			Msgf("[buckets][scaleway][application][vault] failed to create secret for %s", resourceName)
	}

	return app, nil
}
