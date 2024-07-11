import * as hcloud from '@pulumi/hcloud';

import {
  commonLabels,
  environment,
  globalName,
  networkConfig,
  serverConfig,
} from '../configuration';

/**
 * Creates a Hetzner firewall.
 *
 * @returns {hcloud.Firewall} the generated network
 */
export const createFirewall = (): hcloud.Firewall =>
  new hcloud.Firewall(
    'hcloud-firewall-mail',
    {
      name: `${globalName}-${environment}`,
      rules: [
        {
          description: 'Allow incoming SSH traffic',
          direction: 'in',
          port: '22',
          protocol: 'tcp',
          sourceIps: serverConfig.publicSsh
            ? ['0.0.0.0/0', '::/0']
            : [networkConfig.cidr],
        },
        {
          description: 'Allow incoming Prometheus traffic (Mailcow)',
          direction: 'in',
          port: '9099',
          protocol: 'tcp',
          sourceIps: [networkConfig.cidr],
        },
        {
          description: 'Allow incoming mail traffic (SMTP)',
          direction: 'in',
          port: '25',
          protocol: 'tcp',
          sourceIps: ['0.0.0.0/0', '::/0'],
        },
        {
          description: 'Allow incoming mail traffic (SMTPS)',
          direction: 'in',
          port: '465',
          protocol: 'tcp',
          sourceIps: ['0.0.0.0/0', '::/0'],
        },
        {
          description: 'Allow incoming mail traffic (IMAPS)',
          direction: 'in',
          port: '993',
          protocol: 'tcp',
          sourceIps: ['0.0.0.0/0', '::/0'],
        },
        {
          description: 'Allow incoming mail traffic (Sieve)',
          direction: 'in',
          port: '4190',
          protocol: 'tcp',
          sourceIps: ['0.0.0.0/0', '::/0'],
        },
        {
          description: 'Allow incoming web traffic (HTTP)',
          direction: 'in',
          port: '80',
          protocol: 'tcp',
          sourceIps: ['0.0.0.0/0', '::/0'],
        },
        {
          description: 'Allow incoming web traffic (HTTPS)',
          direction: 'in',
          port: '443',
          protocol: 'tcp',
          sourceIps: ['0.0.0.0/0', '::/0'],
        },
      ],
      labels: commonLabels,
    },
    {},
  );
