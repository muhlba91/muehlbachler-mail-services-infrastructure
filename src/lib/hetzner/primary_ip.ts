import * as hcloud from '@pulumi/hcloud';

import { commonLabels, environment, globalName } from '../configuration';

/**
 * Creates a Hetzner primary IP address.
 *
 * @param {string} type the IP type of the primary IP address
 * @param {string} datacenter the datacenter of the primary IP address
 * @returns {hcloud.PrimaryIp} the generated primary IP address
 */
export const createPrimaryIP = (
  type: string,
  datacenter: string,
): hcloud.PrimaryIp =>
  new hcloud.PrimaryIp(
    `hcloud-primary-ip-mail-${type}${datacenter == 'fsn1-dc14' ? '' : '-' + datacenter}`, // FIXME: nbg1
    {
      name: `${globalName}-${environment}-${type}-${datacenter}`,
      assigneeType: 'server',
      type: type,
      datacenter: datacenter,
      autoDelete: false,
      labels: commonLabels,
    },
    {},
  );
