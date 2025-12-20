package ntfy

import "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"

// Config defines configuration data for Ntfy.
type Config struct {
	// Domain defines the Ntfy domain configuration.
	Domain *dns.DomainConfig `yaml:"domain,omitempty"`
}
