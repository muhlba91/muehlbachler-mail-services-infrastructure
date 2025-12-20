package dns

// DomainConfig defines configuration data for a DNS domain.
type DomainConfig struct {
	// Name is the domain name.
	Name *string `yaml:"name,omitempty"`
	// ZoneID is the DNS zone ID.
	ZoneID *string `yaml:"zoneId,omitempty"`
	// Project is the GCP project ID.
	Project *string `yaml:"project,omitempty"`
}
