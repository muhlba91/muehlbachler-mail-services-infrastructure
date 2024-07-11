/**
 * Defines network configuration.
 */
export interface NetworkConfig {
  readonly name: string;
  readonly cidr: string;
  readonly subnetCidr: string;
}
