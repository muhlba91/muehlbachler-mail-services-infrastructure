import { all } from '@pulumi/pulumi';

import { DKIMKey } from '../../model/dkim';
import { globalName } from '../configuration';
import { createRSAkey } from '../util/rsa_key';
import { writeToVault } from '../util/vault/secret';

/**
 * Creates the SimpleLogin DKIM key.
 *
 * @returns {DKIMKey} the DKIM key
 */
export const createDKIMKey = (): DKIMKey => {
  const dkimKey = createRSAkey('dkim-simplelogin-relay', { rsaBits: 2048 });

  writeToVault(
    'simplelogin-dkim',
    all([dkimKey.privateKeyPem, dkimKey.publicKeyPem]).apply(
      ([privateKey, publicKey]) =>
        JSON.stringify({ private_key: privateKey, public_key: publicKey }),
    ),
    globalName,
  );

  return {
    resource: dkimKey,
    publicKey: dkimKey.publicKeyPem.apply((key) =>
      key
        .replace('-----BEGIN PUBLIC KEY-----\n', '')
        .replace('-----END PUBLIC KEY-----', '')
        .trim()
        .split('\n')
        .join(''),
    ),
    privateKey: dkimKey.privateKeyPem,
  };
};
