import { all } from '@pulumi/pulumi';

import { globalName } from './lib/configuration';
import { installDocker } from './lib/docker';
import { installGCloud } from './lib/gcloud';
import { createHetznerSetup } from './lib/hetzner';
import { installMailcow } from './lib/mailcow';
import { createDNSRecords } from './lib/mailcow/record';
import { createPostgresql } from './lib/postgresql';
import { installSimpleLogin } from './lib/simplelogin';
import { installTraefik } from './lib/traefik';
import { createDir } from './lib/util/create_dir';
import { createRandomPassword } from './lib/util/random';
import { createSSHKey } from './lib/util/ssh_key';
import { writeFilePulumiAndUploadToS3 } from './lib/util/storage';
import { writeToVault } from './lib/util/vault/secret';

export = async () => {
  createDir('outputs');

  // server authentication
  const sshKey = createSSHKey('mail', {});

  // mailcow passwords
  const dbUserPassword = createRandomPassword('db-user', {
    special: false,
  });
  const dbRootPassword = createRandomPassword('db-root', {
    special: false,
  });
  const mailcowApiKeyReadWrite = createRandomPassword(
    'mailcow-api-read-write',
    {
      special: false,
    },
  );
  const mailcowApiKeyRead = createRandomPassword('mailcow-api-read-only', {
    special: false,
  });
  writeToVault(
    'mailcow-api',
    all([mailcowApiKeyReadWrite.password, mailcowApiKeyRead.password]).apply(
      ([readWrite, read]) =>
        JSON.stringify({ read_write: readWrite, read: read }),
    ),
    globalName,
  );

  // database
  const postgresqlUsers = createPostgresql();

  // server
  const server = await createHetznerSetup(sshKey.publicKeyOpenssh);
  createDNSRecords(server.publicIPv4, server.publicIPv6);

  // install docker
  const docker = installDocker(server.sshIPv4, sshKey.privateKeyPem, [
    server.resource,
  ]);

  // install gcloud
  const gcloud = installGCloud(server.sshIPv4, sshKey.privateKeyPem, [
    docker,
    server.resource,
  ]);

  // install traefik
  const traefik = gcloud.apply((gcloudInstall) =>
    installTraefik(server.sshIPv4, sshKey.privateKeyPem, [
      docker,
      gcloudInstall,
      server.resource,
    ]),
  );

  // install mailcow
  all([gcloud, traefik]).apply(([gcloudInstall, traefikInstall]) =>
    installMailcow(
      server.publicIPv4,
      server.publicIPv6,
      server.sshIPv4,
      sshKey.privateKeyPem,
      dbUserPassword.password,
      dbRootPassword.password,
      mailcowApiKeyReadWrite.password,
      mailcowApiKeyRead.password,
      [docker, gcloudInstall, traefikInstall, server.resource],
    ),
  );

  // install simplelogin
  const dkim = all([gcloud, traefik]).apply(([gcloudInstall, traefikInstall]) =>
    installSimpleLogin(server.sshIPv4, sshKey.privateKeyPem, postgresqlUsers, [
      docker,
      gcloudInstall,
      traefikInstall,
      server.resource,
    ]),
  );

  // write output files for the server
  writeFilePulumiAndUploadToS3('ssh.key', sshKey.privateKeyPem, {
    permissions: '0600',
  });

  return {
    network: {
      name: server.network,
    },
    server: {
      network: {
        public: {
          ipv4: server.publicIPv4,
          ipv6: server.publicIPv6,
          ssh: server.sshIPv4,
        },
        private: {
          ipv4: server.privateIPv4,
        },
      },
    },
    simplelogin: {
      dkim: {
        publicKey: dkim.publicKey,
        privateKey: dkim.privateKey,
      },
    },
  };
};
