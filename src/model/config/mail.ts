/**
 * Defines mail configuration.
 */
export interface MailConfig {
  readonly main: MailDomainConfig;
  readonly additional?: MailDomainConfig[];
  readonly dkimSignHeaders: string[];
}

/**
 * Defines a mail domain configuration.
 */
export interface MailDomainConfig {
  readonly name: string;
  readonly zoneId: string;
  readonly project?: string;
}
