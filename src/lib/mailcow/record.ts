import { Output } from '@pulumi/pulumi';

import { MailDomainConfig } from '../../model/config/mail';
import { mailConfig } from '../configuration';
import { createRecord } from '../google/dns/record';
import { getMailname } from '../util/mail';

const MAIN_SERVER = `${getMailname(mailConfig.main.name)}.`;

/**
 * Creates the necessary base DNS records.
 *
 * @param {Output<string>} ipv4 the IPv4 address
 * @param {Output<string>} ipv6 the IPv6 address
 */
export const createDNSRecords = (
  ipv4: Output<string>,
  ipv6: Output<string>,
) => {
  // main server A/AAAA records
  const mainServer = getMailname(mailConfig.main.name);
  createRecord(mainServer, mailConfig.main.zoneId, 'A', [ipv4], {
    project: mailConfig.main.project,
  });
  createRecord(mainServer, mailConfig.main.zoneId, 'AAAA', [ipv6], {
    project: mailConfig.main.project,
  });

  // main domain
  createDomainDNSRecords(mailConfig.main, true);

  // additional domains have a CNAME pointing to the main domain
  (mailConfig.additional || []).forEach((domain) => {
    createDomainDNSRecords(domain, false);
  });
};

/**
 * Creates the necessary DNS records for a (sub)domain.
 *
 * @param {MailDomainConfig} domain the domain configuration
 * @param {boolean} main whether this is the main domain
 */
const createDomainDNSRecords = (domain: MailDomainConfig, main: boolean) => {
  // if this is not the main domain, create the 'mail' record
  if (!main) {
    createRecord(`mail.${domain.name}`, domain.zoneId, 'CNAME', [MAIN_SERVER], {
      project: domain.project,
    });
  }

  // create the necessary autodiscover, autoconfig, and mta-sts records
  createRecord(
    `autodiscover.${domain.name}`,
    domain.zoneId,
    'CNAME',
    [MAIN_SERVER],
    {
      project: domain.project,
    },
  );
  createRecord(
    `autoconfig.${domain.name}`,
    domain.zoneId,
    'CNAME',
    [MAIN_SERVER],
    {
      project: domain.project,
    },
  );
  createRecord(
    `mta-sts.${domain.name}`,
    domain.zoneId,
    'CNAME',
    [MAIN_SERVER],
    {
      project: domain.project,
    },
  );
};
