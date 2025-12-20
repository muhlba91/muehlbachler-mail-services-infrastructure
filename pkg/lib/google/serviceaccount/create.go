package serviceaccount

import (
	"fmt"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/google/iam/role"
	gcsIam "github.com/muhlba91/pulumi-shared-library/pkg/lib/google/storage/iam"
	gmodel "github.com/muhlba91/pulumi-shared-library/pkg/model/google/iam/serviceaccount"
	slServiceAccount "github.com/muhlba91/pulumi-shared-library/pkg/util/google/iam/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Create a Google Cloud Service Account with necessary IAM roles.
// ctx: Pulumi context for resource management.
func Create(ctx *pulumi.Context, dnsConfig *dns.Config) (*gmodel.User, error) {
	iam, err := slServiceAccount.CreateServiceAccountUser(ctx, &slServiceAccount.CreateServiceAccountUserArgs{
		Name: fmt.Sprintf("%s-%s", config.GlobalName, config.Environment),
	})
	if err != nil {
		return nil, err
	}

	iam.ServiceAccount.Email.ApplyT(func(email string) error {
		_, _ = gcsIam.CreateIAMMember(ctx, &gcsIam.MemberArgs{
			BucketID: config.BackupBucketID,
			Member:   fmt.Sprintf("serviceAccount:%s", email),
			Role:     "roles/storage.objectAdmin",
		})
		_, _ = gcsIam.CreateIAMMember(ctx, &gcsIam.MemberArgs{
			BucketID: config.BackupBucketID,
			Member:   fmt.Sprintf("serviceAccount:%s", email),
			Role:     "roles/storage.legacyBucketReader",
		})

		_, _ = role.CreateMember(ctx, fmt.Sprintf("%s-dns-admin", config.GlobalNameShort), &role.MemberArgs{
			Member:  pulumi.Sprintf("serviceAccount:%s", email),
			Roles:   []string{"roles/dns.admin"},
			Project: pulumi.String(*dnsConfig.Project),
		})

		return nil
	})

	return iam, nil
}
