import * as hcloud from '@pulumi/hcloud';
import { Output } from '@pulumi/pulumi';

import { mailConfig } from '../configuration';
import { hetznerIdentifierToNumber } from '../util/hetzner';
import { getMailname } from '../util/mail';

/**
 * Creates the necessary reverse DNS records.
 *
 * @param {hcloud.PrimaryIp} ipv4 the IPv4 address
 * @param {hcloud.PrimaryIp} ipv6 the IPv6 address
 * @param {Output<string>} ipv6Address the IPv6 address as a string
 */
export const createReverseDNSRecords = (
  ipv4: hcloud.PrimaryIp,
  ipv6: hcloud.PrimaryIp,
  ipv6Address: Output<string>,
) => {
  const mainServer = getMailname(mailConfig.main.name);

  // reverse DNS entries in Hetzner
  new hcloud.Rdns('hcloud-rdns-ipv4', {
    primaryIpId: ipv4.id.apply(hetznerIdentifierToNumber),
    ipAddress: ipv4.ipAddress,
    dnsPtr: mainServer,
  });
  new hcloud.Rdns('hcloud-rdns-ipv6', {
    primaryIpId: ipv6.id.apply(hetznerIdentifierToNumber),
    ipAddress: ipv6Address,
    dnsPtr: mainServer,
  });
};
