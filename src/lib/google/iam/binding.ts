import * as gcp from '@pulumi/gcp';

import { sanitizeText } from '../../util/string';

/**
 * Defines a new IAM binding.
 *
 * @param {string} accountId the service account id
 * @param {string[]} role the role to assign
 * @param {Output<string>| string} member the name of the member
 */
export const createIAMBinding = (
  accountId: string,
  role: string,
  member: string,
) => {
  new gcp.serviceaccount.IAMBinding(
    `gcp-iam-binding-${sanitizeText(accountId)}-${sanitizeText(
      member,
    )}-${sanitizeText(role)}`,
    {
      serviceAccountId: accountId,
      role: role,
      members: [member],
    },
    {},
  );
};
