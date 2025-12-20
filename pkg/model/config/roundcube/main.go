package roundcube

import "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"

// Config defines configuration data for Roundcube.
type Config struct {
	// Domain defines the Roundcube domain configuration.
	Domain *dns.DomainConfig `yaml:"domain,omitempty"`
}
