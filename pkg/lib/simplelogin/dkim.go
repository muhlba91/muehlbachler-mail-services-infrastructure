package simplelogin

import (
	"encoding/json"
	"strings"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/dns"
	mailConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/mail"
	simpleloginConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/simplelogin"
	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/dkim"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/tls"
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/vault/secret"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dkimKeyLength defines the length of the DKIM RSA key.
const dkimKeyLength = 2048

// createDKIMConfig creates the DKIM configuration for SimpleLogin and necessary DNS records.
// ctx: Pulumi context.
// conn: SSH connection arguments to the remote server.
// simpleloginConfig: Configuration for SimpleLogin installation.
// mailConfig: Configuration for mail services.
// dnsConfig: Configuration for DNS services.
// opts: Additional Pulumi resource options.
func createDKIMConfig(ctx *pulumi.Context,
	conn *remote.ConnectionArgs,
	simpleloginConfig *simpleloginConf.Config,
	mailConfig *mailConf.Config,
	dnsConfig *dns.Config,
	opts ...pulumi.ResourceOption,
) (*dkim.Data, pulumi.Output, error) {
	dkimKey, dkErr := createDKIMKey(ctx)
	if dkErr != nil {
		return nil, nil, dkErr
	}
	dnsErr := createDNSRecords(ctx, dkimKey.PublicKey, mailConfig, dnsConfig, simpleloginConfig)
	if dnsErr != nil {
		return nil, nil, dnsErr
	}

	dkimKeyHash, _ := file.WritePulumi("./outputs/simplelogin_dkim.key", dkimKey.PrivateKey).
		ApplyT(func(_ string) string {
			hash, _ := file.Hash("./outputs/simplelogin_dkim.key")
			return *hash
		}).(pulumi.StringOutput)
	dkimKeyCopy := dkimKeyHash.ApplyT(func(_ string) pulumi.ResourceOption {
		cmd, _ := remote.NewCopyToRemote(
			ctx,
			"remote-copy-simplelogin-dkim-key",
			&remote.CopyToRemoteArgs{
				Source:     pulumi.NewFileAsset("./outputs/simplelogin_dkim.key"),
				RemotePath: pulumi.String("/opt/simplelogin/dkim.key"),
				Triggers:   pulumi.Array{dkimKeyHash},
				Connection: conn,
			},
			opts...)
		return pulumi.DependsOn([]pulumi.Resource{cmd})
	})

	return dkimKey, dkimKeyCopy, nil
}

// createDKIMKey creates a DKIM key pair and stores it in Vault.
// ctx: The Pulumi context for resource creation.
func createDKIMKey(
	ctx *pulumi.Context,
) (*dkim.Data, error) {
	rsaKey, rsaErr := tls.CreateRSAKey(ctx, "dkim-simplelogin-relay", dkimKeyLength)
	if rsaErr != nil {
		return nil, rsaErr
	}

	secretValue, _ := pulumi.All(rsaKey.PrivateKeyPem, rsaKey.PublicKeyPem).ApplyT(func(args []any) string {
		privateKey := args[0].(string)
		publicKey := args[1].(string)

		value, _ := json.Marshal(map[string]string{
			"private_key": privateKey,
			"public_key":  publicKey,
		})
		return string(value)
	}).(pulumi.StringOutput)
	_, sErr := secret.Create(ctx, &secret.CreateOptions{
		Path:  config.GlobalName,
		Key:   "simplelogin-dkim",
		Value: secretValue,
	})
	if sErr != nil {
		return nil, sErr
	}

	publicKey, _ := rsaKey.PublicKeyPem.ApplyT(func(key string) string {
		k := strings.ReplaceAll(key, "-----BEGIN PUBLIC KEY-----\n", "")
		k = strings.ReplaceAll(k, "-----END PUBLIC KEY-----", "")
		k = strings.TrimSpace(k)
		ks := strings.Split(k, "\n")
		return strings.Join(ks, "")
	}).(pulumi.StringOutput)
	return &dkim.Data{
		Resource:   rsaKey,
		PublicKey:  publicKey,
		PrivateKey: rsaKey.PrivateKeyPem,
	}, nil
}
