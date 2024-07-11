import { remote } from '@pulumi/command';
import { all, Output, Resource } from '@pulumi/pulumi';

import { b64decode } from '../util/base64';
import { getFileHash, readFileContents, writeFilePulumi } from '../util/file';

import { createServiceAccount } from './service_account';

/**
 * Installs GCloud.
 *
 * @param {Output<string>} ipv4Address the IPv4 address
 * @param {Output<string>} sshKey the SSH key
 * @param {readonly Resource[]} dependsOn the resources this command depends on
 * @returns {Output<remote.Command>} the remote command
 */
export const installGCloud = (
  ipv4Address: Output<string>,
  sshKey: Output<string>,
  dependsOn: readonly Resource[],
): Output<remote.Command> => {
  const serviceAccount = createServiceAccount();

  const connection = {
    host: ipv4Address,
    privateKey: sshKey,
    user: 'root',
  };

  const prepare = new remote.Command(
    'remote-command-prepare-gcloud',
    {
      create: readFileContents('./assets/gcloud/prepare.sh'),
      connection: connection,
    },
    {
      dependsOn: [...dependsOn],
    },
  );

  const gcpCredentialsHash = writeFilePulumi(
    './outputs/google_credentials.json',
    serviceAccount.key.privateKey.apply((key) => b64decode(key)),
    {},
  )
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    .apply((_) => getFileHash('./outputs/google_credentials.json'));
  const gcpCredentialsCopy = gcpCredentialsHash.apply(
    (hash) =>
      new remote.CopyFile(
        'remote-copy-gcloud-service-account',
        {
          localPath: './outputs/google_credentials.json',
          remotePath: '/opt/google/credentials.json',
          triggers: [Output.create(hash)],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, prepare],
        },
      ),
  );

  return all([gcpCredentialsCopy]).apply(
    ([credentialsCopy]) =>
      new remote.Command(
        'remote-command-install-gcloud',
        {
          create: readFileContents('./assets/gcloud/install.sh'),
          update: readFileContents('./assets/gcloud/install.sh'),
          triggers: [gcpCredentialsHash],
          connection: connection,
        },
        {
          dependsOn: [...dependsOn, credentialsCopy],
        },
      ),
  );
};
