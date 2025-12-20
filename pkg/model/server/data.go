package server

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

// Data represents the data of a Hetzner server.
type Data struct {
	// Resource is the Pulumi resource representing the server.
	Resource pulumi.Resource
	// Hostname is the hostname of the server.
	Hostname pulumi.StringOutput
	// PrivateIPv4 is the private IPv4 address of the server.
	PrivateIPv4 pulumi.StringOutput
	// PublicIPv4 is the public IPv4 address of the server.
	PublicIPv4 pulumi.StringOutput
	// PublicIPv6 is the public IPv6 address of the server.
	PublicIPv6 pulumi.StringOutput
	// SSHIPv4 is the SSH IPv4 address of the server.
	SSHIPv4 pulumi.StringOutput
	// Network is the network of the server.
	Network pulumi.StringOutput
}
