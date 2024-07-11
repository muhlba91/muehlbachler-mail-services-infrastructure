import * as crypto from 'crypto';
import * as fs from 'fs';

import { Output } from '@pulumi/pulumi';

/**
 * Writes the contents to a file.
 *
 * @param {string} path the path to the file
 * @param {string} content the content
 * @param {string} permissions the permissions (default: 0644)
 */
export const writeFileContents = (
  path: string,
  content: string,
  { permissions = '0644' }: { readonly permissions?: string },
) => fs.writeFileSync(path, content, { mode: permissions });

/**
 * Writes the pulumi Output to a file.
 *
 * @param {string} path the path to the file
 * @param {Output<string>} content the content
 * @param {string} permissions the permissions (default: 0644)
 * @returns {Output<unknown>} to track state
 */
export const writeFilePulumi = (
  path: string,
  content: Output<string>,
  { permissions = '0644' }: { readonly permissions?: string },
): Output<string> =>
  content
    .apply((value) =>
      writeFileContents(path, value, { permissions: permissions }),
    )
    .apply(() => content);

/**
 * Reads the contents of a given file.
 *
 * @param {string} path the path to the file
 * @returns {string} the contents
 */
export const readFileContents = (path: string): string =>
  fs.readFileSync(path, 'utf8').toString();

/**
 * Creates the hash of a file.
 *
 * @param {string} path the path to the file
 * @returns {string} the hash
 */
export const getFileHash = (path: string): string =>
  crypto.createHash('sha512').update(fs.readFileSync(path)).digest('hex');
