import { remote } from '@pulumi/command';
import { all, Output, Resource } from '@pulumi/pulumi';

import { dnsConfig } from '../configuration';
import { getFileHash, readFileContents, writeFileContents } from '../util/file';
import { renderTemplate } from '../util/template';

/**
 * Installs Traefik.
 *
 * @param {Output<string>} ipv4Address the IPv4 address
 * @param {Output<string>} sshKey the SSH key
 * @param {readonly Resource[]} dependsOn the resources this command depends on
 * @returns {Output<remote.Command>} the remote command
 */
export const installTraefik = (
  ipv4Address: Output<string>,
  sshKey: Output<string>,
  dependsOn: readonly Resource[],
): Output<remote.Command> => {
  const connection = {
    host: ipv4Address,
    privateKey: sshKey,
    user: 'root',
  };

  const prepare = new remote.Command(
    'remote-command-prepare-traefik',
    {
      create: readFileContents('./assets/traefik/prepare.sh'),
      connection: connection,
    },
    {
      dependsOn: [...dependsOn],
    },
  );

  const dockerComposeHash = Output.create(
    renderTemplate('./assets/traefik/docker-compose.yml.j2', {
      gcpProject: dnsConfig.project,
    }),
  )
    .apply((content) =>
      writeFileContents('./outputs/traefik_docker-compose.yml', content, {}),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/traefik_docker-compose.yml'));
  const dockerComposeCopy = dockerComposeHash.apply(
    (hash) =>
      new remote.CopyFile(
        'remote-copy-traefik-docker-compose',
        {
          localPath: './outputs/traefik_docker-compose.yml',
          remotePath: '/opt/traefik/docker-compose.yml',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const traefikYmlHash = Output.create(
    renderTemplate('./assets/traefik/traefik.yml.j2', {
      acmeEmail: dnsConfig.email,
    }),
  )
    .apply((content) =>
      writeFileContents('./outputs/traefik_traefik.yml', content, {}),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/traefik_traefik.yml'));
  const traefikYmlCopy = traefikYmlHash.apply(
    (hash) =>
      new remote.CopyFile(
        'remote-copy-traefik-config',
        {
          localPath: './outputs/traefik_traefik.yml',
          remotePath: '/opt/traefik/traefik.yml',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const systemdServiceHash = getFileHash('./assets/traefik/traefik.service');
  const systemdServiceCopy = new remote.CopyFile(
    'remote-copy-traefik-service',
    {
      localPath: './assets/traefik/traefik.service',
      remotePath: '/etc/systemd/system/traefik.service',
      triggers: [Output.create(systemdServiceHash)],
      connection: connection,
    },
    {
      dependsOn: [...dependsOn, prepare],
    },
  );

  return all([dockerComposeCopy, traefikYmlCopy]).apply(
    ([composeCopy, traefikCopy]) =>
      new remote.Command(
        'remote-command-install-traefik',
        {
          create: readFileContents('./assets/traefik/install.sh'),
          update: readFileContents('./assets/traefik/install.sh'),
          triggers: [dockerComposeHash, systemdServiceHash, traefikYmlHash],
          connection: connection,
        },
        {
          dependsOn: [
            ...dependsOn,
            composeCopy,
            traefikCopy,
            systemdServiceCopy,
          ],
        },
      ),
  );
};
