package mail

import "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"

// Config defines configuration data for the mail server.
type Config struct {
	// Main is the main domain configuration.
	Main *dns.DomainConfig `yaml:"main,omitempty"`
	// Additional is a list of additional domain configurations.
	Additional []*dns.DomainConfig `yaml:"additional,omitempty"`
	// DkimSignHeaders is a list of headers to be signed with DKIM.
	DkimSignHeaders []string `yaml:"dkimSignHeaders,omitempty"`
}
