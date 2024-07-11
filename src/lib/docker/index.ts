import { remote } from '@pulumi/command';
import { Output, Resource } from '@pulumi/pulumi';

import { readFileContents } from '../util/file';
import { renderTemplate } from '../util/template';

/**
 * Installs Docker.
 *
 * @param {Output<string>} ipv4Address the IPv4 address
 * @param {Output<string>} sshKey the SSH key
 * @param {readonly Resource[]} dependsOn the dependencies
 * @returns {remote.Command} the remote command
 */
export const installDocker = (
  ipv4Address: Output<string>,
  sshKey: Output<string>,
  dependsOn: readonly Resource[],
): remote.Command => {
  const connection = {
    host: ipv4Address,
    privateKey: sshKey,
    user: 'root',
  };

  return new remote.Command(
    'remote-command-install-docker',
    {
      create: renderTemplate('./assets/docker/install.sh', {
        daemonJson: readFileContents('./assets/docker/daemon.json'),
      }),
      connection: connection,
    },
    {
      dependsOn: [...dependsOn],
    },
  );
};
