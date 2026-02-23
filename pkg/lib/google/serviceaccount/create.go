package serviceaccount

import (
	"encoding/json"
	"fmt"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/google/iam/role"
	gcsIam "github.com/muhlba91/pulumi-shared-library/pkg/lib/google/storage/iam"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	gmodel "github.com/muhlba91/pulumi-shared-library/pkg/model/google/iam/serviceaccount"
	slServiceAccount "github.com/muhlba91/pulumi-shared-library/pkg/util/google/iam/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

// Create a Google Cloud Service Account with necessary IAM roles.
// ctx: Pulumi context for resource management.
// dnsConfig: DNS configuration containing project information for IAM role assignment.
func Create(ctx *pulumi.Context, dnsConfig *dns.Config) (*gmodel.User, error) {
	iam, err := slServiceAccount.CreateServiceAccountUser(ctx, &slServiceAccount.CreateOptions{
		Name: fmt.Sprintf("%s-%s", config.GlobalName, config.Environment),
	})
	if err != nil {
		return nil, err
	}

	iam.ServiceAccount.Email.ApplyT(func(email string) error {
		_, _ = gcsIam.CreateIAMMember(ctx, &gcsIam.MemberOptions{
			BucketID: config.BackupBucketID,
			Member:   fmt.Sprintf("serviceAccount:%s", email),
			Role:     "roles/storage.objectAdmin",
		})
		_, _ = gcsIam.CreateIAMMember(ctx, &gcsIam.MemberOptions{
			BucketID: config.BackupBucketID,
			Member:   fmt.Sprintf("serviceAccount:%s", email),
			Role:     "roles/storage.legacyBucketReader",
		})

		_, _ = role.CreateMember(ctx, fmt.Sprintf("%s-dns-admin", config.GlobalNameShort), &role.MemberOptions{
			Member:  pulumi.Sprintf("serviceAccount:%s", email),
			Roles:   []string{"roles/dns.admin"},
			Project: pulumi.String(*dnsConfig.Project),
		})

		return nil
	})

	vaultValue, _ := (iam.Key.PrivateKey.ApplyT(func(creds string) string {
		data, errMarshal := json.Marshal(map[string]string{
			"credentials": creds,
			"bucket":      config.BackupBucketID,
		})
		if errMarshal != nil {
			log.Error().Err(errMarshal).Msg("[google][serviceaccount][vault] failed to marshal credentials")
		}
		return string(data)
	})).(pulumi.StringOutput)

	_, errVault := secret.Create(ctx, &secret.CreateOptions{
		Key:   "google-cloud",
		Value: vaultValue,
		Path:  config.GlobalName,
	})
	if errVault != nil {
		log.Error().Err(errVault).Msg("[google][serviceaccount][vault] failed to create secret")
	}

	return iam, nil
}
