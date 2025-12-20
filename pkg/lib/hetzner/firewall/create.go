package firewall

import (
	"fmt"

	slFirewall "github.com/muhlba91/pulumi-shared-library/pkg/lib/hetzner/firewall"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	networkConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/network"
	serverConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/server"
)

// networkAllCIDR defines the CIDR blocks that represent all IP addresses.
//
//nolint:gochecknoglobals // global is acceptable here
var networkAllCIDR = []pulumi.StringInput{pulumi.String("0.0.0.0/0"), pulumi.String("::/0")}

// Create gets or creates a Hetzner firewall based on the provided configuration.
// ctx: Pulumi context
// networkConfig: Configuration for the Hetzner network.
// serverConfig: Configuration for the Hetzner server.
func Create(
	ctx *pulumi.Context,
	networkConfig *networkConf.Config,
	serverConfig *serverConf.Config,
) (*hcloud.Firewall, error) {
	sshSourceIps := networkAllCIDR
	if !*serverConfig.PublicSSH {
		sshSourceIps = []pulumi.StringInput{pulumi.String(*networkConfig.SubnetCIDR)}
	}
	sshRule := slFirewall.Rule{
		Description: pulumi.String("Allow incoming SSH traffic"),
		Direction:   "in",
		Port:        "22",
		Protocol:    "tcp",
		SourceIPs:   sshSourceIps,
	}

	rules := []slFirewall.Rule{
		sshRule,
		{
			Description: pulumi.String("Allow incoming Prometheus traffic (Mailcow)"),
			Direction:   "in",
			Port:        "9099",
			Protocol:    "tcp",
			SourceIPs:   []pulumi.StringInput{pulumi.String(*networkConfig.SubnetCIDR)},
		},
		{
			Description: pulumi.String("Allow incoming mail traffic (SMTP)"),
			Direction:   "in",
			Port:        "25",
			Protocol:    "tcp",
			SourceIPs:   networkAllCIDR,
		},
		{
			Description: pulumi.String("Allow incoming mail traffic (SMTPS)"),
			Direction:   "in",
			Port:        "465",
			Protocol:    "tcp",
			SourceIPs:   networkAllCIDR,
		},
		{
			Description: pulumi.String("Allow incoming mail traffic (IMAPS)"),
			Direction:   "in",
			Port:        "993",
			Protocol:    "tcp",
			SourceIPs:   networkAllCIDR,
		},
		{
			Description: pulumi.String("Allow incoming mail traffic (Sieve)"),
			Direction:   "in",
			Port:        "4190",
			Protocol:    "tcp",
			SourceIPs:   networkAllCIDR,
		},
		{
			Description: pulumi.String("Allow incoming web traffic (HTTP)"),
			Direction:   "in",
			Port:        "80",
			Protocol:    "tcp",
			SourceIPs:   networkAllCIDR,
		},
		{
			Description: pulumi.String("Allow incoming web traffic (HTTPS)"),
			Direction:   "in",
			Port:        "443",
			Protocol:    "tcp",
			SourceIPs:   networkAllCIDR,
		},
	}

	return slFirewall.Create(ctx, config.GlobalNameShort, &slFirewall.CreateOptions{
		Name:   fmt.Sprintf("%s-%s", config.GlobalName, config.Environment),
		Labels: config.CommonLabels(),
		Rules:  rules,
	})
}
