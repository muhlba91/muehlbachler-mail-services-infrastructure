/**
 * Defines server configuration.
 */
export interface ServerConfig {
  readonly location: string;
  readonly type: string;
  readonly ipv4: string;
  readonly publicSsh?: boolean;
}
