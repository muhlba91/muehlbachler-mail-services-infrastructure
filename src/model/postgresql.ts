import * as pg from '@pulumi/postgresql';
import { Output } from '@pulumi/pulumi';

/**
 * Defines a PostgreSQL user.
 */
export interface PostgresqlUserData {
  readonly user: pg.Role;
  readonly password: Output<string>;
}
