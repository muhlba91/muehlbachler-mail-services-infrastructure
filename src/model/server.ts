import { Output, Resource } from '@pulumi/pulumi';

/**
 * Defines a server.
 */
export interface ServerData {
  readonly resource: Resource;
  readonly privateIPv4: Output<string>;
  readonly publicIPv4: Output<string>;
  readonly publicIPv6: Output<string>;
  readonly sshIPv4: Output<string>;
  readonly network?: Output<string>;
}
