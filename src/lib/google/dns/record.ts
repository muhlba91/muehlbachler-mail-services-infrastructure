import * as gcp from '@pulumi/gcp';
import { Output } from '@pulumi/pulumi';

import { dnsConfig } from '../../configuration';
import { sanitizeText } from '../../util/string';

export const defaultTtl = 300;

/**
 * Creates a record.
 *
 * @param {string} domain the record's domain
 * @param {string | Output<string>} zoneId the zone's ID
 * @param {string} type the record's type
 * @param {readonly string[] | readonly Output<string>[]} records the records to apply
 * @param {number} ttl the TTL to set
 * @param {string} project the project to use
 * @returns {gcp.dns.RecordSet} the record
 */
export const createRecord = (
  domain: string,
  zoneId: string | Output<string>,
  type: string,
  records: readonly string[] | readonly Output<string>[],
  {
    ttl = defaultTtl,
    project,
  }: {
    readonly ttl?: number;
    readonly project?: string;
  },
): gcp.dns.RecordSet =>
  new gcp.dns.RecordSet(
    `dns-record-${type}-${sanitizeText(domain)}`,
    {
      managedZone: zoneId,
      name: `${domain}.`,
      type: type,
      rrdatas: records.map((val) => val),
      ttl: ttl,
      project: project ?? dnsConfig.project,
    },
    {},
  );
