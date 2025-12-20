package mailcow

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

// Secrets holds all generated secrets for mailcow.
type Secrets struct {
	// DBUserPassword is the password for the mailcow database user.
	DBUserPassword pulumi.StringOutput
	// DBRootPassword is the password for the mailcow database root user.
	DBRootPassword pulumi.StringOutput
	// RedisPassword is the password for the mailcow Redis instance.
	RedisPassword pulumi.StringOutput
	// APIKeyReadWrite is the read-write API key for mailcow.
	APIKeyReadWrite pulumi.StringOutput
	// APIKeyRead is the read-only API key for mailcow.
	APIKeyRead pulumi.StringOutput
}
