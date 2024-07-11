import * as gcp from '@pulumi/gcp';
import { CustomResourceOptions, Output } from '@pulumi/pulumi';

import { getProject } from '../../util/google/project';
import { sanitizeText } from '../../util/string';

/**
 * Defines a new IAM member.
 *
 * @param {string} name the name of the account
 * @param {Output<string>| string} member the name of the member
 * @param {string[]} roles the roles
 * @param {string} project the project to create the IAM binding in (default: current project)
 * @param {CustomResourceOptions} pulumiOptions pulumi options (optional)
 */
export const createIAMMember = (
  name: string,
  member: Output<string> | string,
  roles: readonly string[],
  {
    project,
    pulumiOptions,
  }: {
    readonly project?: string;
    readonly pulumiOptions?: CustomResourceOptions;
  },
) => {
  roles.forEach(
    (role) =>
      new gcp.projects.IAMMember(
        `gcp-iam-member-${name}-${getProject(project)}-${sanitizeText(role)}`,
        {
          member: member,
          role: role,
          project: getProject(project),
        },
        pulumiOptions,
      ),
  );
};
