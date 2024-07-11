import * as pg from '@pulumi/postgresql';
import { all } from '@pulumi/pulumi';

import { StringMap } from '../../model/map';
import { PostgresqlUserData } from '../../model/postgresql';
import { databaseConfig, globalName } from '../configuration';
import { createRandomPassword } from '../util/random';
import { writeToVault } from '../util/vault/secret';

/**
 * Creates the users.
 *
 * @param {pg.Provider} provider the database provider
 * @returns {StringMap<PostgresqlUserData>} a map containing users and their passwords
 */
export const createUsers = (
  provider: pg.Provider,
): StringMap<PostgresqlUserData> =>
  Object.fromEntries(
    databaseConfig.users.map((user) => {
      const password = createRandomPassword(`pg-user-${user}`, {
        length: 32,
        special: false,
      });
      const pgUser = new pg.Role(
        `pg-db-user-${user}`,
        {
          name: user,
          password: password.password,
          createDatabase: false,
          createRole: false,
          login: true,
        },
        { provider: provider },
      );

      writeToVault(
        `postgresql-user-${user.toLowerCase()}`,
        all([pgUser.name, password.password]).apply(([user, userPassword]) =>
          JSON.stringify({ user, password: userPassword }),
        ),
        globalName,
      );

      return [user, { password: password.password, user: pgUser }];
    }),
  );
