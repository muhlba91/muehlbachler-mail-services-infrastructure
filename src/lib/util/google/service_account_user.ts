import { ServiceAccountData } from '../../../model/google/service_account_data';
import { environment, globalName } from '../../configuration';
import { createKey } from '../../google/iam/key';
import { createServiceAccount } from '../../google/iam/service_account';

/**
 * Creates a new service account and key.
 *
 * @param {string} name the name
 * @param {string[]} roles the roles to add (optional)
 * @param {string} project the project (optional)
 * @returns {ServiceAccountData} the user data
 */
export const createGCPServiceAccountAndKey = (
  name: string,
  {
    roles,
    project,
  }: {
    readonly roles?: readonly string[];
    readonly project?: string;
  },
): ServiceAccountData => {
  const accountName = `${globalName}-${name}-${environment}`;
  const serviceAccount = createServiceAccount(name, {
    roles: roles,
    project: project,
  });
  const key = createKey(accountName, serviceAccount.name, [serviceAccount]);
  return {
    serviceAccount: serviceAccount,
    key: key,
  };
};
