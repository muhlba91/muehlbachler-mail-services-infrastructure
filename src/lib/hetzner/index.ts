import { Output } from '@pulumi/pulumi';

import { ServerData } from '../../model/server';
import { networkConfig, serverConfig } from '../configuration';
import { hetznerIdentifierToNumber } from '../util/hetzner';

import { createFirewall } from './firewall';
import { createSubnet, getOrCreateNetwork } from './network';
import { createPrimaryIP } from './primary_ip';
import { createReverseDNSRecords } from './reverse_dns';
import { createServer } from './server';
import { registerSSHKey } from './ssh_key';

/**
 * Creates the setup with Hetzner Cloud.
 *
 * @param {Output<string>} publicSSHKey the public SSH key
 * @returns {Promise<ServerData>} the server
 */
export const createHetznerSetup = async (
  publicSSHKey: Output<string>,
): Promise<ServerData> => {
  // location & datacenter
  // FIXME: nbg1
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const location = 'nbg1';
  const datacenter = 'nbg1-dc3';

  // ssh key
  const hetznerSSHKey = registerSSHKey(publicSSHKey);

  // network
  const network = await getOrCreateNetwork();
  createSubnet(network, networkConfig.subnetCidr);
  const firewall = createFirewall();
  const primaryIPs = {
    ipv4: createPrimaryIP('ipv4', 'fsn1-dc14'),
    ipv6: createPrimaryIP('ipv6', 'fsn1-dc14'),
  };
  const primaryIPAddresses = {
    ipv4: primaryIPs.ipv4.ipAddress,
    ipv6: primaryIPs.ipv6.ipAddress.apply((ip) => `${ip}1`),
  };
  // FIXME: nbg1
  const primaryIPsNbg = {
    ipv4: createPrimaryIP('ipv4', datacenter),
    ipv6: createPrimaryIP('ipv6', datacenter),
  };
  const primaryIPAddressesNbg = {
    ipv4: primaryIPsNbg.ipv4.ipAddress,
    ipv6: primaryIPsNbg.ipv6.ipAddress.apply((ip) => `${ip}1`),
  };

  // dns
  createReverseDNSRecords(
    primaryIPs.ipv4,
    primaryIPs.ipv6,
    primaryIPAddresses.ipv6,
    'fsn1-dc14',
  );
  // FIXME: nbg1
  createReverseDNSRecords(
    primaryIPsNbg.ipv4,
    primaryIPsNbg.ipv6,
    primaryIPAddressesNbg.ipv6,
    datacenter,
  );

  // server
  const server = createServer(
    'fsn1',
    serverConfig.type,
    hetznerSSHKey.id,
    firewall.id.apply(hetznerIdentifierToNumber),
    network,
    serverConfig.ipv4,
    primaryIPs.ipv4.id.apply(hetznerIdentifierToNumber),
    primaryIPs.ipv6.id.apply(hetznerIdentifierToNumber),
  );
  // FIXME: nbg1
  // const serverNbg = createServer(
  //   location,
  //   serverConfig.type,
  //   hetznerSSHKey.id,
  //   firewall.id.apply(hetznerIdentifierToNumber),
  //   network,
  //   // serverConfig.ipv4,
  //   '10.20.0.11',
  //   primaryIPsNbg.ipv4.id.apply(hetznerIdentifierToNumber),
  //   primaryIPsNbg.ipv6.id.apply(hetznerIdentifierToNumber),
  // );

  return {
    resource: server,
    privateIPv4: Output.create(serverConfig.ipv4),
    publicIPv4: primaryIPAddresses.ipv4,
    publicIPv6: primaryIPAddresses.ipv6,
    sshIPv4: serverConfig.publicSsh
      ? primaryIPAddresses.ipv4
      : Output.create(serverConfig.ipv4),
    network: Output.create(networkConfig.name),
  };
};
