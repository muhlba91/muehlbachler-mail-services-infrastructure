package simplelogin

import (
	"fmt"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/aws/iam/accesskey"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/aws/iam/policy"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/aws/iam/user"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createAWSUser creates an AWS IAM user with permissions to access the specified S3 bucket.
// ctx: The Pulumi context for resource creation.
// bucketArn: The ARN of the S3 bucket the user should have access to.
func createAWSUser(
	ctx *pulumi.Context,
	bucketArn string,
) (*iam.AccessKeyOutput, error) {
	allow := "Allow"
	policyDoc, _ := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect:  &allow,
				Actions: []string{"s3:*"},
				Resources: []string{
					fmt.Sprintf("%s/*", bucketArn),
				},
			},
		},
	})
	policy, polErr := policy.Create(ctx, "simplelogin", &policy.CreateOptions{
		Policy: pulumi.String(policyDoc.Json),
		Labels: config.CommonLabels(),
	})
	if polErr != nil {
		return nil, polErr
	}

	usr, uErr := user.Create(ctx, fmt.Sprintf("%s-simplelogin", config.GlobalName), &user.CreateOptions{
		Policies: []*iam.Policy{policy},
		Labels:   config.CommonLabels(),
	})
	if uErr != nil {
		return nil, uErr
	}

	accessKey, _ := usr.Name.ApplyT(func(name string) *iam.AccessKey {
		key, _ := accesskey.Create(ctx, &accesskey.CreateOptions{
			UserName: name,
			User:     usr,
		})
		return key
	}).(iam.AccessKeyOutput)

	return &accessKey, nil
}
