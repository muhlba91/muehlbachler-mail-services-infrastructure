/**
 * Defines ntfy configuration.
 */
export interface NtfyConfig {
  readonly domain: NtfyDomainConfig;
}

/**
 * Defines a ntfy domain configuration.
 */
export interface NtfyDomainConfig {
  readonly name: string;
  readonly zoneId: string;
  readonly project?: string;
}
