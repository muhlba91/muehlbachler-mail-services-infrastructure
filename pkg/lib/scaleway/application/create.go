package application

import (
	"fmt"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/scaleway"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/scaleway/iam/policy"
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
		Name: resourceName,
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

	return app, nil
}
