package dkim

import (
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Data holds the DKIM resources.
type Data struct {
	// Resource is the TLS private key resource.
	Resource *tls.PrivateKey
	// PublicKey is the DKIM public key.
	PublicKey pulumi.StringOutput
	// PrivateKey is the DKIM private key.
	PrivateKey pulumi.StringOutput
}
