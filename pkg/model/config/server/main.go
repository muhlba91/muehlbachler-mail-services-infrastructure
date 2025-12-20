package server

// Config defines configuration data for the server.
type Config struct {
	// Location is the server location.
	Location *string `yaml:"location,omitempty"`
	// Type is the server type.
	Type *string `yaml:"type,omitempty"`
	// IPv4 is the server IPv4 address.
	IPv4 *string `yaml:"ipv4,omitempty"`
	// PublicSSH indicates if public SSH access is enabled.
	PublicSSH *bool `yaml:"publicSsh,omitempty"`
}
