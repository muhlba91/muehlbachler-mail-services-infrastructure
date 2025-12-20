package dns

import (
	"fmt"

	mailConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/mail"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/pulumi/convert"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateReverseDNSRecords creates reverse DNS records for the given IPv4 and IPv6 primary IPs.
// ctx: Pulumi context.
// ipv4: Primary IPv4 address.
// ipv6: Primary IPv6 address.
// ipv6Address: IPv6 address.
// datacenter: Datacenter name.
// mailConfig: Mail configuration for determining the mail server name.
func CreateReverseDNSRecords(ctx *pulumi.Context,
	ipv4 *hcloud.PrimaryIp,
	ipv6 *hcloud.PrimaryIp,
	ipv6Address pulumi.StringOutput,
	datacenter string,
	mailConfig *mailConf.Config,
) error {
	mainServer := mail.Mailname(*mailConfig.Main.Name)

	_, rdns4Err := hcloud.NewRdns(ctx, fmt.Sprintf("hcloud-rdns-ipv4-%s", datacenter), &hcloud.RdnsArgs{
		PrimaryIpId: convert.IDToInt(ipv4.ID()),
		IpAddress:   ipv4.IpAddress,
		DnsPtr:      pulumi.String(mainServer),
	})
	if rdns4Err != nil {
		return rdns4Err
	}

	_, rdns6Err := hcloud.NewRdns(ctx, fmt.Sprintf("hcloud-rdns-ipv6-%s", datacenter), &hcloud.RdnsArgs{
		PrimaryIpId: convert.IDToInt(ipv6.ID()),
		IpAddress:   ipv6Address,
		DnsPtr:      pulumi.String(mainServer),
	})
	if rdns6Err != nil {
		return rdns6Err
	}

	return nil
}
