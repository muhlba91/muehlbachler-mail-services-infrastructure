import { Output } from '@pulumi/pulumi';
import * as vault from '@pulumi/vault';

/**
 * Stores a value in Vault.
 *
 * @param {string} key the key
 * @param {Output<string>} value the value
 * @param {string} path the path
 */
export const writeToVault = (
  key: string,
  value: Output<string>,
  path: string,
) => {
  new vault.kv.SecretV2(
    `vault-secret-${path}-${key}`,
    {
      mount: path,
      name: key,
      dataJson: value,
    },
    {},
  );
};
