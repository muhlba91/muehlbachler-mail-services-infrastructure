/**
 * Defines server configuration.
 */
export interface ServerConfig {
  readonly type: string;
  readonly ipv4: string;
  readonly publicSsh?: boolean;
}
