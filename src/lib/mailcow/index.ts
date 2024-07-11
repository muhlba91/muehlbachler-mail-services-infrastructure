import { remote } from '@pulumi/command';
import { all, Output, Resource } from '@pulumi/pulumi';
import { parse } from 'yaml';

import { backupBucketId, dnsConfig, mailConfig } from '../configuration';
import { getFileHash, readFileContents, writeFileContents } from '../util/file';
import { getProject } from '../util/google/project';
import { getMailname } from '../util/mail';
import { BUCKET_PATH, writeFilePulumiAndUploadToS3 } from '../util/storage';
import { renderTemplate } from '../util/template';

/**
 * Installs Mailcow.
 *
 * @param {Output<string>} ipv4Address the IPv4 address
 * @param {Output<string>} ipv6Address the IPv6 address
 * @param {Output<string>} sshIPv4 the SSH IPv4 address
 * @param {Output<string>} sshKey the SSH key
 * @param {Output<string>} dbUserPassword the database user password
 * @param {Output<string>} dbRootPassword the database root password
 * @param {Output<string>} mailcowApiKeyReadWrite the Mailcow read-write API key
 * @param {Output<string>} mailcowApiKeyRead the Mailcow read-only API key
 * @param {readonly Resource[]} dependsOn the resources this command depends on
 */
export const installMailcow = (
  ipv4Address: Output<string>,
  ipv6Address: Output<string>,
  sshIPv4: Output<string>,
  sshKey: Output<string>,
  dbUserPassword: Output<string>,
  dbRootPassword: Output<string>,
  mailcowApiKeyReadWrite: Output<string>,
  mailcowApiKeyRead: Output<string>,
  dependsOn: readonly Resource[],
) => {
  const connection = {
    host: sshIPv4,
    privateKey: sshKey,
    user: 'root',
  };

  const prepare = new remote.Command(
    'remote-command-prepare-mailcow',
    {
      create: readFileContents('./assets/mailcow/prepare.sh'),
      connection: connection,
    },
    {
      dependsOn: [...dependsOn],
    },
  );

  const cronFileHash = getFileHash('./assets/mailcow/cron/cron');
  const cronFileCopy = new remote.CopyFile(
    'remote-copy-mailcow-cron',
    {
      localPath: './assets/mailcow/cron/cron',
      remotePath: '/etc/cron.d/mailcow',
      triggers: [Output.create(cronFileHash)],
      connection: connection,
    },
    {
      dependsOn: [...dependsOn, prepare],
    },
  );

  const backupFileHash = Output.create(
    renderTemplate('./assets/mailcow/cron/mailcow-backup.j2', {
      project: getProject(),
      bucket: {
        id: backupBucketId,
        path: BUCKET_PATH,
      },
    }),
  )
    .apply((content) =>
      writeFileContents('./outputs/mailcow_backup', content, {}),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/mailcow_backup'));
  const backupFileCopy = backupFileHash.apply(
    (hash) =>
      new remote.CopyFile(
        'remote-copy-mailcow-backup',
        {
          localPath: './outputs/mailcow_backup',
          remotePath: '/bin/mailcow-backup',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const cronInstall = all([cronFileCopy, backupFileCopy]).apply(
    ([cronCopy, backupCopy]) =>
      new remote.Command(
        'remote-command-install-mailcow-cron',
        {
          create: readFileContents('./assets/mailcow/cron/install.sh'),
          update: readFileContents('./assets/mailcow/cron/install.sh'),
          triggers: [cronFileHash, backupFileHash],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare, cronCopy, backupCopy],
        },
      ),
  );

  const systemdServiceHash = getFileHash('./assets/mailcow/mailcow.service');
  const systemdServiceCopy = new remote.CopyFile(
    'remote-copy-mailcow-service',
    {
      localPath: './assets/mailcow/mailcow.service',
      remotePath: '/etc/systemd/system/mailcow.service',
      triggers: [Output.create(systemdServiceHash)],
      connection: connection,
    },
    {
      dependsOn: [...dependsOn, prepare],
    },
  );

  const dockerComposeHash = all([mailcowApiKeyRead])
    .apply((apiKey) =>
      renderTemplate('./assets/mailcow/docker-compose.override.yml.j2', {
        mailname: getMailname(mailConfig.main.name),
        apiKey: apiKey,
      }),
    )
    .apply((content) =>
      writeFileContents(
        './outputs/mailcow_docker-compose.override.yml',
        content,
        {},
      ),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/mailcow_docker-compose.override.yml'));
  const dockerComposeCopy = dockerComposeHash.apply(
    (hash) =>
      new remote.CopyFile(
        'remote-copy-mailcow-docker-compose',
        {
          localPath: './outputs/mailcow_docker-compose.override.yml',
          remotePath: '/opt/mailcow/docker-compose.override.yml',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const configFileHash = all([
    dbUserPassword,
    dbRootPassword,
    mailcowApiKeyReadWrite,
    mailcowApiKeyRead,
    ipv4Address,
    ipv6Address,
  ])
    .apply(
      ([userPassword, rootPassword, apiKeyReadWrite, apiKeyRead, ipv4, ipv6]) =>
        renderTemplate('./assets/mailcow/config/mailcow.conf.j2', {
          mailname: getMailname(mailConfig.main.name),
          db: {
            auth: {
              user: userPassword,
              root: rootPassword,
            },
          },
          ip: {
            v4: ipv4,
            v6: ipv6,
          },
          acme: {
            email: dnsConfig.email,
          },
          api: {
            readWrite: apiKeyReadWrite,
            read: apiKeyRead,
          },
        }),
    )
    .apply((content) =>
      writeFilePulumiAndUploadToS3(
        'mailcow_mailcow.conf',
        Output.create(content),
        {},
      ),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/mailcow_mailcow.conf'));
  const configFileCopy = configFileHash.apply(
    (hash) =>
      new remote.CopyFile(
        'remote-copy-mailcow-conf',
        {
          localPath: './outputs/mailcow_mailcow.conf',
          remotePath: '/opt/mailcow/mailcow.conf',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const mailcowVersion = dockerComposeHash.apply(
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    (_) =>
      parse(
        readFileContents(
          './outputs/mailcow_docker-compose.override.yml',
        ).replace('#', ''),
      )['version'],
  );

  // TODO: restore doesn't work automated - https://github.com/mailcow/mailcow-dockerized/pull/5934
  const installCommand = mailcowVersion.apply((version) =>
    renderTemplate('./assets/mailcow/install.sh.j2', {
      version: version,
      project: getProject(),
      bucket: {
        id: backupBucketId,
        path: BUCKET_PATH,
      },
      dkimSignHeaders: mailConfig.dkimSignHeaders.join(':'),
    }),
  );
  const installTask = all([configFileCopy, cronInstall]).apply(
    ([configCopy, cronInstaller]) =>
      new remote.Command(
        'remote-command-install-mailcow',
        {
          create: installCommand,
          update: installCommand,
          triggers: [
            systemdServiceHash,
            dockerComposeHash,
            configFileHash,
            mailcowVersion,
          ],
          connection: connection,
        },
        {
          dependsOn: [
            ...dependsOn,
            prepare,
            systemdServiceCopy,
            dockerComposeCopy,
            configCopy,
            cronInstaller,
          ],
        },
      ),
  );

  const bodyChecksHash = getFileHash(
    './assets/mailcow/config/body_checks.pcre',
  );
  const bodyChecksCopy = installTask.apply(
    (installer) =>
      new remote.CopyFile(
        'remote-copy-mailcow-postfix-body-checks',
        {
          localPath: './assets/mailcow/config/body_checks.pcre',
          remotePath: '/opt/mailcow/data/conf/postfix/body_checks.pcre',
          triggers: [Output.create(bodyChecksHash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare, installer],
        },
      ),
  );

  const clientHeadersHash = getFileHash(
    './assets/mailcow/config/client_headers.pcre',
  );
  const clientHeadersCopy = installTask.apply(
    (installer) =>
      new remote.CopyFile(
        'remote-copy-mailcow-postfix-client-headers',
        {
          localPath: './assets/mailcow/config/client_headers.pcre',
          remotePath: '/opt/mailcow/data/conf/postfix/client_headers.pcre',
          triggers: [Output.create(clientHeadersHash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare, installer],
        },
      ),
  );

  const postfixExtraHash = getFileHash('./assets/mailcow/config/extra.cf');
  const postfixExtraCopy = installTask.apply(
    (installer) =>
      new remote.CopyFile(
        'remote-copy-mailcow-postfix-extra',
        {
          localPath: './assets/mailcow/config/extra.cf',
          remotePath: '/opt/mailcow/data/conf/postfix/extra.cf',
          triggers: [Output.create(postfixExtraHash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare, installer],
        },
      ),
  );

  all([postfixExtraCopy, bodyChecksCopy, clientHeadersCopy, installTask]).apply(
    ([postfixExtra, bodyChecks, clientHeaders, installer]) =>
      new remote.Command(
        'remote-command-postinstall-mailcow',
        {
          create: readFileContents('./assets/mailcow/postinstall.sh'),
          update: readFileContents('./assets/mailcow/postinstall.sh'),
          triggers: [
            systemdServiceHash,
            postfixExtraHash,
            bodyChecksHash,
            clientHeadersHash,
          ],
          connection: connection,
        },
        {
          dependsOn: [
            ...dependsOn,
            prepare,
            postfixExtra,
            bodyChecks,
            clientHeaders,
            installer,
          ],
        },
      ),
  );
};
