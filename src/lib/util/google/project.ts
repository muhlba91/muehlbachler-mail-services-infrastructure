import * as gcp from '@pulumi/gcp';

/**
 * Gets the current project.
 *
 * @param {string} project the project input
 * @returns {string} the project or a default project
 */
export const getProject = (project?: string): string =>
  project ?? gcp.config.project ?? '';
