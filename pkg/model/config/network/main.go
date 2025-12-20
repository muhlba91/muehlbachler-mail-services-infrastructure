package network

// Config defines network configuration.
type Config struct {
	// Name is the name of the network.
	Name *string `yaml:"name,omitempty"`
	// CIDR is the CIDR block for the network.
	CIDR *string `yaml:"cidr,omitempty"`
	// SubnetCIDR is the CIDR block for the subnet.
	SubnetCIDR *string `yaml:"subnetCidr,omitempty"`
}
