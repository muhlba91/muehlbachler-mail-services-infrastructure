package network

import (
	"github.com/muhlba91/pulumi-shared-library/pkg/lib/hetzner/network"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/pulumi/convert"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/lib/config"
	networkConf "github.com/muhlba91/muehlbachler-mail-services-infrastructure/pkg/model/config/network"
)

// GetOrCreate retrieves an existing Hetzner network or creates a new one based on the provided configuration.
// ctx: Pulumi context
// networkConfig: Configuration for the Hetzner network.
func GetOrCreate(ctx *pulumi.Context, networkConfig *networkConf.Config) (*pulumi.IntOutput, error) {
	lNet, lErr := network.Get(ctx, *networkConfig.Name)
	if lErr == nil && lNet != nil {
		id := pulumi.Int(lNet.Id).ToIntOutput()
		return &id, nil
	}

	cNet, cErr := network.Create(ctx, &network.CreateOptions{
		Name:   *networkConfig.Name,
		Cidr:   pulumi.String(*networkConfig.CIDR),
		Labels: config.CommonLabels(),
	})
	if cErr != nil {
		return nil, cErr
	}
	id := convert.IDToInt(cNet.ID())
	return &id, nil
}
