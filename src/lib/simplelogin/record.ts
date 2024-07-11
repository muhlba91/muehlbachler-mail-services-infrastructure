import { interpolate, Output } from '@pulumi/pulumi';

import { dnsConfig, mailConfig, simpleloginConfig } from '../configuration';
import { createRecord } from '../google/dns/record';

const DKIM_SELECTORS = ['dkim', 'dkim02', 'dkim03'];

/**
 * Creates the base DNS records.
 *
 * @param {Output<string>} dkimPublicKey the public DKIM key
 */
export const createDNSRecords = (dkimPublicKey: Output<string>) => {
  DKIM_SELECTORS.forEach((selector) =>
    createRecord(
      `${selector}._domainkey.${simpleloginConfig.mail.domain}`,
      simpleloginConfig.mail.zoneId ?? mailConfig.main.zoneId,
      'TXT',
      [
        interpolate`v=DKIM1; k=rsa; t=s; s=email; p=${dkimPublicKey}`.apply(
          (entry) => splitByLength(entry, 'TXT'),
        ),
      ],
      {
        project:
          simpleloginConfig.mail.project ??
          mailConfig.main.project ??
          dnsConfig.project,
      },
    ),
  );
};

/**
 * Splits the value by the allowed maximum length.
 *
 * @param {string} value the value
 * @param {string} type the entry type
 * @returns
 */
const splitByLength = (value: string, type: string) => {
  const split = value.split(/(.{200})/).filter((x) => x.length > 0);

  return split.length > 1 || type == 'TXT'
    ? `"${split.join('" "')}"`
    : split.join();
};
