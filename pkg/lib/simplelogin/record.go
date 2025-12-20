package simplelogin

import (
	"fmt"
	"strings"

	"github.com/muhlba91/pulumi-shared-library/pkg/lib/google/dns/record"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	dnsConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"
	mailConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	simpleloginConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/simplelogin"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/mail"
)

// createDNSRecords creates DNS records for SimpleLogin based on the provided DNS configuration.
// ctx: The Pulumi context for resource creation.
// dkimPublicKey: The DKIM public key to be used in DNS records.
// mailConfig: The mail configuration containing domain and other settings.
// dnsConfig: The DNS configuration containing zone and domain information.
// simpleloginConfig: The SimpleLogin configuration containing domain and other settings.
func createDNSRecords(
	ctx *pulumi.Context,
	dkimPublicKey pulumi.StringOutput,
	mailConfig *mailConf.Config,
	dnsConfig *dnsConf.Config,
	simpleloginConfig *simpleloginConf.Config,
) error {
	dkimSelectors := []string{"dkim", "dkim02", "dkim03"}

	mainServerDomain, zoneID, project := mail.DNSCoreDetails(
		simpleloginConfig.Mail.ZoneID,
		simpleloginConfig.Mail.Project,
		mailConfig,
		dnsConfig,
	)

	_, cnErr := record.Create(ctx, &record.CreateOptions{
		Domain:     *simpleloginConfig.Domain,
		ZoneID:     zoneID,
		RecordType: "CNAME",
		Records:    pulumi.StringArray([]pulumi.StringInput{mainServerDomain}),
		Project:    &project,
	})
	if cnErr != nil {
		return cnErr
	}

	for _, selector := range dkimSelectors {
		records, _ := pulumi.Sprintf("v=DKIM1; k=rsa; t=s; s=email; p=%s", dkimPublicKey).ApplyT(func(value string) string {
			return splitByLength(value, "TXT")
		}).(pulumi.StringOutput)
		_, dkimErr := record.Create(ctx, &record.CreateOptions{
			Domain:     fmt.Sprintf("%s._domainkey.%s", selector, *simpleloginConfig.Mail.Domain),
			ZoneID:     zoneID,
			RecordType: "TXT",
			Records: pulumi.StringArray{
				records,
			},
			Project: &project,
		})
		if dkimErr != nil {
			return dkimErr
		}
	}

	return nil
}

// splitByLength splits a string into chunks and formats the result according to the DNS record type.
// value: The string to be split.
// typ: The DNS record type (e.g., "TXT").
func splitByLength(value string, typ string) string {
	const maxLen = 200
	if value == "" {
		return ""
	}

	var parts []string
	for i := 0; i < len(value); i += maxLen {
		end := i + maxLen
		if end > len(value) {
			end = len(value)
		}
		parts = append(parts, value[i:end])
	}

	if len(parts) > 1 || typ == "TXT" {
		return fmt.Sprintf("\"%s\"", strings.Join(parts, "\" \""))
	}

	return strings.Join(parts, "")
}
