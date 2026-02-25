package server

import (
	"fmt"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/hetzner/dns"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/hetzner/firewall"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/hetzner/network"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/hetzner/network/subnet"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/hetzner/primaryip"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/hetzner/server"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/hetzner/sshkey"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/hetzner/location"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/pulumi/convert"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	networkConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/network"
	serverConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/server"
	serverModel "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/server"
)

// Create creates a new Hetzner server.
// ctx: Pulumi context
// publicSSHKey: Public SSH key to be added to the server for access.
// serverConfig: Configuration for the server.
// networkConfig: Configuration for the network the server will be part of.
// mailConfig: Configuration for mail services, used for DNS setup.
func Create(
	ctx *pulumi.Context,
	publicSSHKey pulumi.StringOutput,
	serverConfig *serverConf.Config,
	networkConfig *networkConf.Config,
	mailConfig *mail.Config,
) (*serverModel.Data, error) {
	// location & datacenter
	dc := location.ToDatacenter(serverConfig.Location)

	// SSH Key
	hetznerSSHKey, hErr := sshkey.Create(ctx, config.GlobalNameShort, &sshkey.CreateOptions{
		Name:      fmt.Sprintf("%s-%s", config.GlobalName, config.Environment),
		PublicKey: publicSSHKey,
		Labels:    config.CommonLabels(),
	})
	if hErr != nil {
		return nil, hErr
	}

	// network
	network, nErr := network.GetOrCreate(ctx, networkConfig)
	if nErr != nil {
		return nil, nErr
	}
	_, _ = subnet.Create(ctx, config.GlobalNameShort, &subnet.CreateOptions{
		NetworkID: network,
		Cidr:      *networkConfig.SubnetCIDR,
	})

	firewall, fErr := firewall.Create(ctx, networkConfig, serverConfig)
	if fErr != nil {
		return nil, fErr
	}

	// primary IPs
	primaryIPv4, primaryIPv6, publicIPv6, pipErr := createIPAddresses(ctx, dc, *serverConfig.Location, mailConfig)
	if pipErr != nil {
		return nil, pipErr
	}

	// server
	enableIPv6 := false
	server, sErr := server.Create(
		ctx,
		fmt.Sprintf("%s-%s", config.GlobalNameShort, *serverConfig.Location),
		&server.CreateOptions{
			Hostname: pulumi.Sprintf(
				"%s-%s-%s",
				config.GlobalName,
				config.Environment,
				*serverConfig.Location,
			),
			ServerType:         pulumi.String(*serverConfig.Type),
			Image:              pulumi.String("ubuntu-24.04"),
			SSHKeys:            []pulumi.StringInput{hetznerSSHKey.ID().ToStringOutput()},
			Location:           pulumi.String(*serverConfig.Location),
			NetworkID:          network,
			IPAddress:          pulumi.String(*serverConfig.IPv4),
			PrimaryIPv4Address: primaryIPv4,
			PrimaryIPv6Address: primaryIPv6,
			EnableIPv6:         &enableIPv6,
			Firewalls:          []pulumi.IntInput{convert.IDToInt(firewall.ID())},
			Backups:            pulumi.Bool(true),
			Protection:         true,
			Labels:             config.CommonLabels(),
			PublicSSH:          *serverConfig.PublicSSH,
		},
	)
	if sErr != nil {
		return nil, sErr
	}

	sshIP := pulumi.String(*serverConfig.IPv4).ToStringOutput()
	if *serverConfig.PublicSSH {
		sshIP = primaryIPv4.IpAddress
	}
	return &serverModel.Data{
		Resource:    server.Resource,
		Hostname:    server.Hostname,
		PrivateIPv4: pulumi.String(*serverConfig.IPv4).ToStringOutput(),
		PublicIPv4:  primaryIPv4.IpAddress,
		PublicIPv6:  *publicIPv6,
		SSHIPv4:     sshIP,
		Network:     pulumi.String(*networkConfig.Name).ToStringOutput(),
	}, nil
}

// createIPAddresses creates primary IPv4 and IPv6 addresses, sets up reverse DNS records, and returns the created IPs.
// ctx: Pulumi context.
// dc: Datacenter where the IPs will be created.
// location: Location for the IPs, used for DNS setup.
// mailConfig: Configuration for mail services, used for DNS setup.
func createIPAddresses(
	ctx *pulumi.Context,
	dc string,
	location string,
	mailConfig *mail.Config,
) (*hcloud.PrimaryIp, *hcloud.PrimaryIp, *pulumi.StringOutput, error) {
	// primary IPs
	primaryIPv4, pv4Err := primaryip.Create(ctx, config.GlobalNameShort, &primaryip.CreateOptions{
		Name:       fmt.Sprintf("%s-%s", config.GlobalName, config.Environment),
		IPType:     "ipv4",
		Datacenter: &dc,
		Location:   location,
		AutoDelete: pulumi.Bool(false),
		Labels:     config.CommonLabels(),
	})
	if pv4Err != nil {
		return nil, nil, nil, pv4Err
	}
	primaryIPv6, pv6Err := primaryip.Create(ctx, config.GlobalNameShort, &primaryip.CreateOptions{
		Name:       fmt.Sprintf("%s-%s", config.GlobalName, config.Environment),
		IPType:     "ipv6",
		Datacenter: &dc,
		Location:   location,
		AutoDelete: pulumi.Bool(false),
		Labels:     config.CommonLabels(),
	})
	if pv6Err != nil {
		return nil, nil, nil, pv6Err
	}
	publicIPv6 := pulumi.Sprintf("%s1", primaryIPv6.IpAddress)

	// dns
	dErr := dns.CreateReverseDNSRecords(ctx, primaryIPv4, primaryIPv6, publicIPv6, dc, mailConfig)
	if dErr != nil {
		return nil, nil, nil, dErr
	}

	return primaryIPv4, primaryIPv6, &publicIPv6, nil
}
