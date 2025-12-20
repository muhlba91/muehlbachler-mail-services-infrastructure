package mailcow

import (
	"fmt"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"
	mailConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/mail"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/google/dns/record"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateDNSRecords creates DNS records for Mailcow based on the provided DNS configuration.
// ctx: The Pulumi context for resource creation.
// dnsConfig: The DNS configuration containing domain and record details.
// ipv4: The public IPv4 address to create DNS records for.
// ipv6: The public IPv6 address to create DNS records for.
func CreateDNSRecords(
	ctx *pulumi.Context,
	mailConfig *mailConf.Config,
	ipv4 pulumi.StringOutput,
	ipv6 pulumi.StringOutput,
) error {
	// main server A/AAAA records
	mainServer := mail.Mailname(*mailConfig.Main.Name)
	mainServerDomain := pulumi.Sprintf("%s.", mainServer)

	_, v4Err := record.Create(ctx, &record.CreateOptions{
		Domain:     mainServer,
		ZoneID:     pulumi.String(*mailConfig.Main.ZoneID),
		RecordType: "A",
		Records:    pulumi.StringArray([]pulumi.StringInput{ipv4}),
		Project:    mailConfig.Main.Project,
	})
	if v4Err != nil {
		return v4Err
	}

	_, v6Err := record.Create(ctx, &record.CreateOptions{
		Domain:     mainServer,
		ZoneID:     pulumi.String(*mailConfig.Main.ZoneID),
		RecordType: "AAAA",
		Records:    pulumi.StringArray([]pulumi.StringInput{ipv6}),
		Project:    mailConfig.Main.Project,
	})
	if v6Err != nil {
		return v6Err
	}

	// main domain
	dnsErr := createDomainRecords(ctx, mailConfig.Main, &mainServerDomain, true)
	if dnsErr != nil {
		return dnsErr
	}

	// additional domains have a CNAME pointing to the main domain
	for _, domain := range mailConfig.Additional {
		drErr := createDomainRecords(ctx, domain, &mainServerDomain, false)
		if drErr != nil {
			return drErr
		}
	}

	return nil
}

// createDomainRecords creates necessary DNS records for a given mail domain.
// ctx: The Pulumi context for resource creation.
// domain: The mail domain configuration.
// primaryDomain: The primary domain to point records to.
// main: A boolean indicating if this is the main domain.
func createDomainRecords(
	ctx *pulumi.Context,
	domain *dns.DomainConfig,
	primaryDomain *pulumi.StringOutput,
	main bool,
) error {
	records := pulumi.StringArray([]pulumi.StringInput{*primaryDomain})

	// if this is not the main domain, create the 'mail' record
	if !main {
		_, mErr := record.Create(ctx, &record.CreateOptions{
			Domain:     fmt.Sprintf("mail.%s", *domain.Name),
			ZoneID:     pulumi.String(*domain.ZoneID),
			RecordType: "CNAME",
			Records:    records,
			Project:    domain.Project,
		})
		if mErr != nil {
			return mErr
		}
	}

	// create the necessary autodiscover, autoconfig, and mta-sts records
	_, aErr := record.Create(ctx, &record.CreateOptions{
		Domain:     fmt.Sprintf("autodiscover.%s", *domain.Name),
		ZoneID:     pulumi.String(*domain.ZoneID),
		RecordType: "CNAME",
		Records:    records,
		Project:    domain.Project,
	})
	if aErr != nil {
		return aErr
	}
	_, acErr := record.Create(ctx, &record.CreateOptions{
		Domain:     fmt.Sprintf("autoconfig.%s", *domain.Name),
		ZoneID:     pulumi.String(*domain.ZoneID),
		RecordType: "CNAME",
		Records:    records,
		Project:    domain.Project,
	})
	if acErr != nil {
		return acErr
	}
	_, mtaErr := record.Create(ctx, &record.CreateOptions{
		Domain:     fmt.Sprintf("mta-sts.%s", *domain.Name),
		ZoneID:     pulumi.String(*domain.ZoneID),
		RecordType: "CNAME",
		Records:    records,
		Project:    domain.Project,
	})
	return mtaErr
}
