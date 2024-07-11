import { PrivateKey } from '@pulumi/tls';

/**
 * Defines a new RSA key.
 *
 * @param {string} name the name of the key
 * @param {number} rsaBits the number of bits for the key (default: 4096)
 * @return {PrivateKey} the generated RSA key
 */
export const createRSAkey = (
  name: string,
  {
    rsaBits = 4096,
  }: {
    readonly rsaBits?: number;
  },
): PrivateKey =>
  new PrivateKey(`rsa-key-${name}`, {
    algorithm: 'RSA',
    rsaBits: rsaBits,
  });
