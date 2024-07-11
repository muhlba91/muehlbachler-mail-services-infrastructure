import * as pg from '@pulumi/postgresql';

import { StringMap } from '../../model/map';
import { PostgresqlUserData } from '../../model/postgresql';
import { databaseConfig, globalName } from '../configuration';
import { writeToVault } from '../util/vault/secret';

/**
 * Creates the databases.
 *
 * @param {StringMap<PostgresqlUserData>} users a map containing users and their passwords
 * @param {pg.Provider} provider the database provider
 */
export const createDatabases = (
  users: StringMap<PostgresqlUserData>,
  provider: pg.Provider,
) => {
  Object.entries(databaseConfig.database).map(([db, owner]) => {
    const pgDb = new pg.Database(
      `pg-db-${db}`,
      {
        name: db,
        owner: owner,
      },
      { provider: provider, dependsOn: [users[owner].user] },
    );

    writeToVault(
      `postgresql-database-${db.toLowerCase()}`,
      pgDb.name.apply((db) => JSON.stringify({ name: db })),
      globalName,
    );
  });
};
