import * as hcloud from '@pulumi/hcloud';

import { commonLabels, environment, globalName } from '../configuration';

/**
 * Creates a Hetzner primary IP address.
 *
 * @param {string} type the IP type of the primary IP address
 * @returns {hcloud.PrimaryIp} the generated primary IP address
 */
export const createPrimaryIP = (type: string): hcloud.PrimaryIp =>
  new hcloud.PrimaryIp(
    `hcloud-primary-ip-mail-${type}`,
    {
      name: `${globalName}-${environment}-${type}`,
      assigneeType: 'server',
      type: type,
      datacenter: 'fsn1-dc14',
      autoDelete: false,
      labels: commonLabels,
    },
    {},
  );
