import { Output } from '@pulumi/pulumi';
import { PrivateKey } from '@pulumi/tls';

/**
 * Defines a DKIM key.
 */
export interface DKIMKey {
  readonly resource: PrivateKey;
  readonly publicKey: Output<string>;
  readonly privateKey: Output<string>;
}
