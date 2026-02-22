package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/mailcow"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/ntfy"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/postgresql"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/roundcube"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/scaleway"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/scaleway/application"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/simplelogin"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/util/file"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/tls"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/dir"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/docker"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/gcloud"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/google/serviceaccount"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/hetzner/server"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/traefik"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/dkim"
	serverModel "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/server"
)

//nolint:gocognit,funlen // main is the entry point of the Pulumi program.
func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		dErr := dir.Create("outputs")
		if dErr != nil {
			return dErr
		}

		// configuration
		dnsConfig, scalewayConfig, networkConfig, serverConfig, mailConfig, simpleloginConfig, roundcubeConfig, ntfyConfig, databaseConfig, err := config.LoadConfig(
			ctx,
		)
		if err != nil {
			return err
		}

		// mailcow secrets
		mailcowSecrets, mcsErr := mailcow.CreateSecrets(ctx)
		if mcsErr != nil {
			return mcsErr
		}

		// database
		postgresqlUsers, pgErr := postgresql.Create(ctx, databaseConfig)
		if pgErr != nil {
			return pgErr
		}

		// instance
		sshKey, sErr := tls.CreateSSHKey(ctx, fmt.Sprintf("%s-%s", config.GlobalNameShort, config.Environment), 0)
		if sErr != nil {
			return sErr
		}
		instance, iErr := server.Create(ctx, sshKey.PublicKeyOpenssh, serverConfig, networkConfig, mailConfig)
		if iErr != nil {
			return iErr
		}
		dependsOn := []pulumi.Resource{instance.Resource}

		// docker
		dockerInstall, doErr := docker.Install(ctx, instance.SSHIPv4, sshKey.PrivateKeyPem, pulumi.DependsOn(dependsOn))
		if doErr != nil {
			return doErr
		}
		dependsOn = append(dependsOn, dockerInstall)

		// google cloud
		serviceAccount, saErr := serviceaccount.Create(ctx, dnsConfig)
		if saErr != nil {
			return saErr
		}
		gcloudInstall, gcErr := gcloud.Install(
			ctx,
			instance.SSHIPv4,
			sshKey.PrivateKeyPem,
			serviceAccount,
			pulumi.DependsOn(dependsOn),
		)
		if gcErr != nil {
			return gcErr
		}
		dependsOn = append(dependsOn, gcloudInstall)

		// scaleway
		scwApplication, scwaErr := application.Create(ctx, scalewayConfig)
		if scwaErr != nil {
			return scwaErr
		}
		scalewayInstall, scwErr := scaleway.Install(
			ctx,
			instance.SSHIPv4,
			sshKey.PrivateKeyPem,
			scwApplication,
			scalewayConfig,
			pulumi.DependsOn(dependsOn),
		)
		if scwErr != nil {
			return scwErr
		}
		dependsOn = append(dependsOn, scalewayInstall)

		// traefik
		traefikInstall, tErr := traefik.Install(
			ctx,
			instance.SSHIPv4,
			sshKey.PrivateKeyPem,
			dnsConfig,
			pulumi.DependsOn(dependsOn),
		)
		if tErr != nil {
			return tErr
		}
		dependsOn = append(dependsOn, traefikInstall)

		// mailcow
		mcErr := mailcow.Install(
			ctx,
			instance.PublicIPv4,
			instance.PublicIPv6,
			instance.SSHIPv4,
			sshKey.PrivateKeyPem,
			mailcowSecrets,
			mailConfig,
			dnsConfig,
			pulumi.DependsOn(dependsOn),
		)
		if mcErr != nil {
			return mcErr
		}
		mcdErr := mailcow.CreateDNSRecords(ctx, mailConfig, instance.PublicIPv4, instance.PublicIPv6)
		if mcdErr != nil {
			return mcdErr
		}

		// simplelogin
		dkim, slErr := simplelogin.Install(
			ctx,
			instance.SSHIPv4,
			sshKey.PrivateKeyPem,
			postgresqlUsers,
			simpleloginConfig,
			serverConfig,
			mailConfig,
			dnsConfig,
			pulumi.DependsOn(dependsOn),
		)
		if slErr != nil {
			return slErr
		}

		// roundcube
		rcErr := roundcube.Install(
			ctx,
			instance.SSHIPv4,
			sshKey.PrivateKeyPem,
			postgresqlUsers,
			mailcowSecrets.APIKeyReadWrite,
			roundcubeConfig,
			mailConfig,
			dnsConfig,
			pulumi.DependsOn(dependsOn),
		)
		if rcErr != nil {
			return rcErr
		}

		// ntfy
		ntfyErr := ntfy.Install(
			ctx,
			instance.SSHIPv4,
			sshKey.PrivateKeyPem,
			ntfyConfig,
			mailConfig,
			dnsConfig,
			pulumi.DependsOn(dependsOn),
		)
		if ntfyErr != nil {
			return ntfyErr
		}

		// write output files
		//nolint:mnd // 0o600 is the correct permission for private keys
		file.WriteAndUpload(ctx, "ssh.key", sshKey.PrivateKeyPem, 0o600)

		// outputs
		exportPulumiOutputs(ctx, instance, dkim)

		return nil
	})
}

// exportPulumiOutputs exports the necessary Pulumi outputs.
// ctx: The Pulumi context.
// instance: The Hetzner server instance data.
// dkim: The DKIM data.
func exportPulumiOutputs(
	ctx *pulumi.Context,
	instance *serverModel.Data,
	dkim *dkim.Data,
) {
	ctx.Export("server", pulumi.ToMap(map[string]any{
		"network": map[string]any{
			"public": map[string]any{
				"ipv4": instance.PublicIPv4,
				"ipv6": instance.PublicIPv6,
				"ssh":  instance.SSHIPv4,
			},
			"private": map[string]any{
				"ipv4": instance.PrivateIPv4,
			},
		},
	}))

	ctx.Export("simplelogin", pulumi.ToMap(map[string]any{
		"dkim": map[string]any{
			"publicKey":  dkim.PublicKey,
			"privateKey": dkim.PrivateKey,
		},
	}))
}
