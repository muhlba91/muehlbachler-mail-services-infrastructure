import * as pg from '@pulumi/postgresql';
import { all } from '@pulumi/pulumi';

import { StringMap } from '../../model/map';
import { PostgresqlUserData } from '../../model/postgresql';
import { globalName, postgresqlConfig } from '../configuration';
import { writeToVault } from '../util/vault/secret';

import { createDatabases } from './database';
import { createUsers } from './user';

/**
 * Creates the Postgresql databases and users.
 *
 * @returns {StringMap<PostgresqlUserData>} a map containing users and their passwords
 */
export const createPostgresql = (): StringMap<PostgresqlUserData> => {
  const pgProvider = new pg.Provider(
    'postgresql',
    {
      host: postgresqlConfig.address,
      port: postgresqlConfig.port,
      username: postgresqlConfig.username,
      password: postgresqlConfig.password,
      superuser: false,
    },
    {},
  );

  writeToVault(
    'postgresql-connection',
    all([postgresqlConfig.address, postgresqlConfig.port]).apply(
      ([host, port]) => JSON.stringify({ port: port.toString(), host: host }),
    ),
    globalName,
  );

  const users = createUsers(pgProvider);
  createDatabases(users, pgProvider);
  return users;
};
