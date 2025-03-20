import { remote } from '@pulumi/command';
import { all, Output, Resource } from '@pulumi/pulumi';
import { FileAsset } from '@pulumi/pulumi/asset';
import { parse } from 'yaml';

import { mailConfig, ntfyConfig } from '../configuration';
import { getFileHash, readFileContents, writeFileContents } from '../util/file';
import { getMailname } from '../util/mail';
import { writeFilePulumiAndUploadToS3 } from '../util/storage';
import { renderTemplate } from '../util/template';

import { createDNSRecords } from './record';

/**
 * Installs ntfy.
 *
 * @param {Output<string>} sshIPv4 the SSH IPv4 address
 * @param {Output<string>} sshKey the SSH key
 * @param {readonly Resource[]} dependsOn the resources this command depends on
 */
export const installNtfy = (
  sshIPv4: Output<string>,
  sshKey: Output<string>,
  dependsOn: readonly Resource[],
) => {
  createDNSRecords();

  const connection = {
    host: sshIPv4,
    privateKey: sshKey,
    user: 'root',
  };

  const prepare = new remote.Command(
    'remote-command-prepare-ntfy',
    {
      create: readFileContents('./assets/ntfy/prepare.sh'),
      connection: connection,
    },
    {
      dependsOn: [...dependsOn],
    },
  );

  const systemdServiceHash = getFileHash('./assets/ntfy/ntfy.service');
  const systemdServiceCopy = new remote.CopyToRemote(
    'remote-copy-ntfy-service',
    {
      source: new FileAsset('./assets/ntfy/ntfy.service'),
      remotePath: '/etc/systemd/system/ntfy.service',
      triggers: [Output.create(systemdServiceHash)],
      connection: connection,
    },
    {
      dependsOn: [...dependsOn, prepare],
    },
  );

  const dockerComposeHash = Output.create(
    renderTemplate('./assets/ntfy/docker-compose.yml.j2', {
      domain: ntfyConfig.domain.name,
    }),
  )
    .apply((content) =>
      writeFileContents('./outputs/ntfy_docker-compose.yml', content, {}),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/ntfy_docker-compose.yml'));
  const dockerComposeCopy = dockerComposeHash.apply(
    (hash) =>
      new remote.CopyToRemote(
        'remote-copy-ntfy-docker-compose',
        {
          source: new FileAsset('./outputs/ntfy_docker-compose.yml'),
          remotePath: '/opt/ntfy/docker-compose.yml',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const configFileHash = Output.create(
    renderTemplate('./assets/ntfy/server.yml.j2', {
      mailname: getMailname(mailConfig.main.name),
      domain: ntfyConfig.domain.name,
    }),
  )
    .apply((content) =>
      writeFilePulumiAndUploadToS3(
        'ntfy_server.yml',
        Output.create(content),
        {},
      ),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/ntfy_server.yml'));
  const configFileCopy = configFileHash.apply(
    (hash) =>
      new remote.CopyToRemote(
        'remote-copy-ntfy-server-yml',
        {
          source: new FileAsset('./outputs/ntfy_server.yml'),
          remotePath: '/opt/ntfy/config/server.yml',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const ntfyVersion = dockerComposeHash.apply(
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    (_) =>
      parse(readFileContents('./outputs/ntfy_docker-compose.yml'))['services'][
        'ntfy'
      ]['image'].split(':')[1],
  );

  const installCommand = ntfyVersion.apply((version) =>
    renderTemplate('./assets/ntfy/install.sh.j2', {
      version: version,
    }),
  );
  const installTask = all([dockerComposeCopy, configFileCopy]).apply(
    ([composeCopy, configCopy]) =>
      new remote.Command(
        'remote-command-install-ntfy',
        {
          create: installCommand,
          update: installCommand,
          triggers: [
            systemdServiceHash,
            dockerComposeHash,
            configFileHash,
            ntfyVersion,
          ],
          connection: connection,
        },
        {
          dependsOn: [
            ...dependsOn,
            prepare,
            systemdServiceCopy,
            composeCopy,
            configCopy,
          ],
        },
      ),
  );

  all([installTask]).apply(
    ([installer]) =>
      new remote.Command(
        'remote-command-postinstall-ntfy',
        {
          create: readFileContents('./assets/ntfy/postinstall.sh'),
          update: readFileContents('./assets/ntfy/postinstall.sh'),
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare, installer],
        },
      ),
  );
};
