/**
 * Defines SimpleLogin configuration.
 */
export interface SimpleLoginConfig {
  readonly domain: string;
  readonly mail: SimpleLoginMailConfig;
  readonly oidc: SimpleLoginOIDCConfig;
}

/**
 * Defines SimpleLogin mail configuration.
 */
export interface SimpleLoginMailConfig {
  readonly domain: string;
  readonly mx: string;
  readonly zoneId?: string;
  readonly project?: string;
}

/**
 * Defines OIDC configuration.
 */
export interface SimpleLoginOIDCConfig {
  readonly wellKnownUrl: string;
  readonly clientId: string;
  readonly clientSecret: string;
}
