package mail

import (
	dnsConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"
	mailConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/defaults"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// DNSCoreDetails returns common DNS details needed for creating DNS records.
// zoneID: Optional zone ID to use for DNS records.
// project: Optional GCP project to use for DNS records.
// mailConfig: Configuration related to mail services.
// dnsConfig: Configuration related to DNS services.
func DNSCoreDetails(
	zoneID *string,
	project *string,
	mailConfig *mailConf.Config,
	dnsConfig *dnsConf.Config,
) (pulumi.StringInput, pulumi.StringInput, string) {
	mainServerDomain := pulumi.Sprintf("%s.", Mailname(*mailConfig.Main.Name))
	zone := pulumi.String(defaults.GetOrDefault(zoneID, *mailConfig.Main.ZoneID))
	proj := defaults.GetOrDefault(
		project,
		defaults.GetOrDefault(mailConfig.Main.Project, *dnsConfig.Project),
	)

	return mainServerDomain, zone, proj
}
