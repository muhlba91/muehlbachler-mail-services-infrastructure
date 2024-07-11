import * as hcloud from '@pulumi/hcloud';
import { Output } from '@pulumi/pulumi';

import { commonLabels, environment, globalName } from '../configuration';

/**
 * Creates a Hetzner SSH key.
 *
 * @param {Output<string>} publicKey the public key of the SSH key
 * @returns {hcloud.SshKey} the registered SSH key
 */
export const registerSSHKey = (publicKey: Output<string>): hcloud.SshKey =>
  new hcloud.SshKey(
    'hcloud-ssh-mail',
    {
      name: `${globalName}-${environment}`,
      publicKey: publicKey,
      labels: commonLabels,
    },
    {},
  );
