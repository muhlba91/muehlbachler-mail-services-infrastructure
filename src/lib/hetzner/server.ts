import * as hcloud from '@pulumi/hcloud';
import { Output } from '@pulumi/pulumi';

import { commonLabels, environment, globalName } from '../configuration';

/**
 * Creates a Hetzner server.
 *
 * @param {string} serverType the server type
 * @param {Output<string>} sshKey the SSH key
 * @param {Output<number>} firewall the firewall identifier
 * @param {Output<number>} network the network identifier
 * @param {string} ipAddress the IPv4 address of the server
 * @param {Output<number>} primaryIPv4Address the primary IPv4 address
 * @param {Output<number>} primaryIPv6Address the primary IPv6 address
 * @returns {hcloud.Server} the generated server
 */
export const createServer = (
  serverType: string,
  sshKey: Output<string>,
  firewall: Output<number>,
  network: Output<number>,
  ipAddress: string,
  primaryIPv4Address: Output<number>,
  primaryIPv6Address: Output<number>,
): hcloud.Server =>
  new hcloud.Server(
    'hcloud-server-mail',
    {
      name: `${globalName}-${environment}`,
      serverType: serverType,
      image: 'ubuntu-24.04',
      sshKeys: [sshKey],
      // datacenter: 'fsn1-dc14',
      location: 'fsn1',
      networks: [
        {
          networkId: network,
          ip: ipAddress,
        },
      ],
      publicNets: [
        {
          ipv4Enabled: true,
          ipv4: primaryIPv4Address,
          ipv6Enabled: false,
          ipv6: primaryIPv6Address,
        },
      ],
      firewallIds: [firewall],
      backups: true,
      labels: commonLabels,
    },
    {},
  );
