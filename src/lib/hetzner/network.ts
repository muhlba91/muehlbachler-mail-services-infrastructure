import * as hcloud from '@pulumi/hcloud';
import { Output } from '@pulumi/pulumi';

import { commonLabels, networkConfig } from '../configuration';
import { hetznerIdentifierToNumber } from '../util/hetzner';

/**
 * Gets or creates a Hetzner network.
 *
 * @returns {Promise<Output<number>>} the network identifier
 */
export const getOrCreateNetwork = async (): Promise<Output<number>> =>
  await hcloud
    .getNetwork({ name: networkConfig.name })
    .then((network) => Output.create(network.id))
    .catch(() => createNetwork().id.apply(hetznerIdentifierToNumber));

/**
 * Creates a Hetzner network.
 *
 * @returns {hcloud.Network} the generated network
 */
export const createNetwork = (): hcloud.Network =>
  new hcloud.Network(
    'hcloud-network-mail',
    {
      name: networkConfig.name,
      ipRange: networkConfig.cidr,
      exposeRoutesToVswitch: true,
      labels: commonLabels,
    },
    {},
  );

/**
 * Creates a Hetzner subnet.
 *
 * @param {Output<number>} network the network identifier to create the subnet in
 * @param {string} cidr the CIDR of the subnet
 * @returns {hcloud.NetworkSubnet} the generated subnet
 */
export const createSubnet = (
  network: Output<number>,
  cidr: string,
): hcloud.NetworkSubnet =>
  new hcloud.NetworkSubnet(
    'hcloud-subnet-mail',
    {
      networkId: network,
      type: 'cloud',
      networkZone: 'eu-central',
      ipRange: cidr,
    },
    {},
  );
