import { StringMap } from '../map';

/**
 * Defines database configuration.
 */
export interface DatabaseConfig {
  readonly users: readonly string[];
  readonly database: StringMap<string>;
}
