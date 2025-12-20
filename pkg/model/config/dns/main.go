package dns

// Config defines configuration data for DNS.
type Config struct {
	// Project is the DNS project identifier.
	Project *string `yaml:"project,omitempty"`
	// Email is the DNS contact email.
	Email *string `yaml:"email,omitempty"`
}
