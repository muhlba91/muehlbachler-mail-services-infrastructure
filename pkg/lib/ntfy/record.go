package ntfy

import (
	dnsConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"
	mailConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	ntfyConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/ntfy"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/mail"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/google/dns/record"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createDNSRecords creates DNS records for Ntfy based on the provided DNS configuration.
// ctx: The Pulumi context for resource creation.
// mailConfig: Configuration related to mail services.
// dnsConfig: Configuration related to DNS services.
// roundcubeConfig: Configuration specific to Roundcube.
func createDNSRecords(
	ctx *pulumi.Context,
	mailConfig *mailConf.Config,
	dnsConfig *dnsConf.Config,
	ntfyConfig *ntfyConf.Config,
) error {
	mainServerDomain, zoneID, project := mail.DNSCoreDetails(
		ntfyConfig.Domain.ZoneID,
		ntfyConfig.Domain.Project,
		mailConfig,
		dnsConfig,
	)

	_, cnErr := record.Create(ctx, &record.CreateOptions{
		Domain:     *ntfyConfig.Domain.Name,
		ZoneID:     zoneID,
		RecordType: "CNAME",
		Records:    pulumi.StringArray([]pulumi.StringInput{mainServerDomain}),
		Project:    &project,
	})

	return cnErr
}
