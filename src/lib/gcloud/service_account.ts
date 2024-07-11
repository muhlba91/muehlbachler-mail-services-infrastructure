import { interpolate } from '@pulumi/pulumi';

import { ServiceAccountData } from '../../model/google/service_account_data';
import { dnsConfig } from '../configuration';
import { createIAMMember } from '../google/iam/member';
import { createGCPServiceAccountAndKey } from '../util/google/service_account_user';

/**
 * Creates the Google Cloud service account for Traefik.
 *
 * @returns {ServiceAccountData} the generated service account
 */
export const createServiceAccount = (): ServiceAccountData => {
  const iam = createGCPServiceAccountAndKey('mail', {
    roles: ['roles/storage.objectUser'],
  });

  createIAMMember(
    'mail-dns-admin',
    interpolate`serviceAccount:${iam.serviceAccount.email}`,
    ['roles/dns.admin'],
    { project: dnsConfig.project },
  );

  return iam;
};
