package simplelogin

// Config defines configuration data for SimpleLogin.
type Config struct {
	// Domain defines the SimpleLogin domain configuration.
	Domain *string `yaml:"domain,omitempty"`
	// Mail defines the mail configuration for SimpleLogin.
	Mail *MailConfig `yaml:"mail,omitempty"`
	// OIDC defines the OIDC configuration for SimpleLogin.
	OIDC *OIDCConfig `yaml:"oidc,omitempty"`
}

// MailConfig defines mail-related configuration for SimpleLogin.
type MailConfig struct {
	// Domain defines the mail domain configuration.
	Domain *string `yaml:"domain,omitempty"`
	// MX defines the mail exchange server configuration.
	MX *string `yaml:"mx,omitempty"`
	// ZoneID defines the zone ID configuration.
	ZoneID *string `yaml:"zoneId,omitempty"`
	// Project defines the project configuration.
	Project *string `yaml:"project,omitempty"`
}

// OIDCConfig defines OIDC-related configuration for SimpleLogin.
type OIDCConfig struct {
	// WellKnownURL defines the well-known URL configuration.
	WellKnownURL *string `yaml:"wellKnownUrl,omitempty"`
	// ClientID defines the client ID configuration.
	ClientID *string `yaml:"clientId,omitempty"`
	// ClientSecret defines the client secret configuration.
	//nolint:gosec // This is not a hardcoded secret, it's a configuration field.
	ClientSecret *string `yaml:"clientSecret,omitempty"`
}
