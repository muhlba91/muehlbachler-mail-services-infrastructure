package scaleway

// Config defines configuration data for Scaleway.
type Config struct {
	// OrganizationID is the Scaleway organization identifier, which may be required for certain API interactions.
	OrganizationID string `yaml:"organizationId,omitempty"`
	// Project is the Scaleway project identifier.
	Project *string `yaml:"project,omitempty"`
	// DNSProject is the Scaleway project identifier for DNS management, which may differ from the main project.
	DNSProject *string `yaml:"dnsProject,omitempty"`
}
