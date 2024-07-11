import * as aws from '@pulumi/aws';
import { remote } from '@pulumi/command';
import { all, Output, Resource } from '@pulumi/pulumi';
import { parse } from 'yaml';

import { DKIMKey } from '../../model/dkim';
import { StringMap } from '../../model/map';
import { PostgresqlUserData } from '../../model/postgresql';
import { createS3Bucket } from '../aws/storage/bucket';
import {
  awsDefaultRegion,
  postgresqlConfig,
  serverConfig,
  simpleloginConfig,
} from '../configuration';
import { getFileHash, readFileContents, writeFileContents } from '../util/file';
import { createRandomPassword } from '../util/random';
import { writeFilePulumiAndUploadToS3 } from '../util/storage';
import { renderTemplate } from '../util/template';

import { createUser } from './aws';
import { createDKIMKey } from './dkim';
import { createDNSRecords } from './record';

/**
 * Installs SimpleLogin.
 *
 * @param {Output<string>} sshIPv4 the SSH IPv4 address
 * @param {Output<string>} sshKey the SSH key
 * @param {StringMap<PostgresqlUserData>} postgresqlUsers the Postgresql users
 * @param {readonly Resource[]} dependsOn the resources this command depends on
 * @returns {DKIMKey} the DKIM key
 */
export const installSimpleLogin = (
  sshIPv4: Output<string>,
  sshKey: Output<string>,
  postgresqlUsers: StringMap<PostgresqlUserData>,
  dependsOn: readonly Resource[],
): DKIMKey => {
  const connection = {
    host: sshIPv4,
    privateKey: sshKey,
    user: 'root',
  };

  const prepare = new remote.Command(
    'remote-command-prepare-simplelogin',
    {
      create: readFileContents('./assets/simplelogin/prepare.sh'),
      connection: connection,
    },
    {
      dependsOn: [...dependsOn],
    },
  );

  const systemdServiceHash = getFileHash(
    './assets/simplelogin/simplelogin.service',
  );
  const systemdServiceCopy = new remote.CopyFile(
    'remote-copy-simplelogin-service',
    {
      localPath: './assets/simplelogin/simplelogin.service',
      remotePath: '/etc/systemd/system/simplelogin.service',
      triggers: [Output.create(systemdServiceHash)],
      connection: connection,
    },
    {
      dependsOn: [...dependsOn, prepare],
    },
  );

  const dockerComposeHash = Output.create(
    renderTemplate('./assets/simplelogin/docker-compose.yml.j2', {
      domain: simpleloginConfig.domain,
    }),
  )
    .apply((content) =>
      writeFileContents(
        './outputs/simplelogin_docker-compose.yml',
        content,
        {},
      ),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/simplelogin_docker-compose.yml'));
  const dockerComposeCopy = dockerComposeHash.apply(
    (hash) =>
      new remote.CopyFile(
        'remote-copy-simplelogin-docker-compose',
        {
          localPath: './outputs/simplelogin_docker-compose.yml',
          remotePath: '/opt/simplelogin/docker-compose.yml',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const dkimKey = createDKIMKey();
  createDNSRecords(dkimKey.publicKey);

  const dkimKeyHash = dkimKey.privateKey
    .apply((content) =>
      writeFileContents('./outputs/simplelogin_dkim.key', content, {}),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/simplelogin_dkim.key'));
  const dkimKeyCopy = dkimKeyHash.apply(
    (hash) =>
      new remote.CopyFile(
        'remote-copy-simplelogin-dkim-key',
        {
          localPath: './outputs/simplelogin_dkim.key',
          remotePath: '/opt/simplelogin/dkim.key',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const flaskSecret = createRandomPassword('simplelogin-flask-secret', {
    length: 32,
    special: false,
  });

  const bucket = createS3Bucket('simplelogin');
  const key = bucket.arn.apply((arn) => createUser(arn));

  const envFileHash = all([
    postgresqlConfig.address,
    postgresqlConfig.port,
    postgresqlUsers['simplelogin'].password,
    flaskSecret.password,
    bucket.bucket,
    key.id,
    key.secret,
  ])
    .apply(
      ([
        postgresqlHost,
        postgresqlPort,
        postgresqlPassword,
        flaskSecretPassword,
        bucketName,
        accessKeyId,
        secretAccessKey,
      ]) =>
        renderTemplate('./assets/simplelogin/env.j2', {
          flaskSecret: flaskSecretPassword,
          db: {
            uri: `postgresql://simplelogin:${postgresqlPassword}@${postgresqlHost}:${postgresqlPort}/simplelogin`,
            host: postgresqlHost,
            port: postgresqlPort,
            database: 'simplelogin',
            user: 'simplelogin',
            password: postgresqlPassword,
          },
          aws: {
            bucket: bucketName,
            region: aws.config.region ?? awsDefaultRegion,
            accessKeyId: accessKeyId,
            secretAccessKey: secretAccessKey,
          },
          oidc: {
            wellKnownUrl: simpleloginConfig.oidc.wellKnownUrl,
            clientId: simpleloginConfig.oidc.clientId,
            clientSecret: simpleloginConfig.oidc.clientSecret,
          },
          domain: simpleloginConfig.domain,
          email: {
            domain: simpleloginConfig.mail.domain,
            mx: simpleloginConfig.mail.mx,
            relay: serverConfig.ipv4,
          },
        }),
    )
    .apply((content) =>
      writeFilePulumiAndUploadToS3(
        'simplelogin_env',
        Output.create(content),
        {},
      ),
    )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/simplelogin_env'));
  const envFileCopy = envFileHash.apply(
    (hash) =>
      new remote.CopyFile(
        'remote-copy-simplelogin-env',
        {
          localPath: './outputs/simplelogin_env',
          remotePath: '/opt/simplelogin/env',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  const simpleloginVersion = dockerComposeHash.apply(
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    (_) =>
      parse(readFileContents('./outputs/simplelogin_docker-compose.yml'))[
        'services'
      ]['app']['image'].split(':')[1],
  );

  const initShHash = getFileHash('./assets/simplelogin/init.sh');
  const initShCopy = new remote.CopyFile(
    'remote-copy-simplelogin-init-sh',
    {
      localPath: './assets/simplelogin/init.sh',
      remotePath: '/opt/simplelogin/init.sh',
      triggers: [Output.create(initShHash)],
      connection: connection,
    },
    {
      dependsOn: [...dependsOn, prepare],
    },
  );

  simpleloginVersion.apply((version) =>
    renderTemplate('./assets/simplelogin/install.sh.j2', {
      version: version,
    }),
  );

  const installCommand = simpleloginVersion.apply((version) =>
    renderTemplate('./assets/simplelogin/install.sh.j2', {
      version: version,
    }),
  );
  all([dockerComposeCopy, dkimKeyCopy, envFileCopy, initShCopy]).apply(
    ([composeCopy, dkimCopy, envCopy, initCopy]) =>
      new remote.Command(
        'remote-command-install-simplelogin',
        {
          create: installCommand,
          update: installCommand,
          triggers: [
            systemdServiceHash,
            dockerComposeHash,
            dkimKeyHash,
            envFileHash,
            initShHash,
            simpleloginVersion,
          ],
          connection: connection,
        },
        {
          dependsOn: [
            ...dependsOn,
            prepare,
            systemdServiceCopy,
            composeCopy,
            dkimCopy,
            envCopy,
            initCopy,
          ],
        },
      ),
  );

  return dkimKey;
};
