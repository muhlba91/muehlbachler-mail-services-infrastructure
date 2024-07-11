import {
  Config,
  getOrganization,
  getStack,
  StackReference,
} from '@pulumi/pulumi';

import { DatabaseConfig } from '../model/config/database';
import { DNSConfig } from '../model/config/dns';
import { MailConfig } from '../model/config/mail';
import { NetworkConfig } from '../model/config/network';
import { ServerConfig } from '../model/config/server';
import { SimpleLoginConfig } from '../model/config/simplelogin';

export const environment = getStack();

const config = new Config();
export const bucketId = config.require('bucketId');
export const backupBucketId = config.require('backupBucketId');
export const dnsConfig = config.requireObject<DNSConfig>('dns');
export const networkConfig = config.requireObject<NetworkConfig>('network');
export const serverConfig = config.requireObject<ServerConfig>('server');
export const mailConfig = config.requireObject<MailConfig>('mail');
export const simpleloginConfig =
  config.requireObject<SimpleLoginConfig>('simplelogin');
export const databaseConfig = config.requireObject<DatabaseConfig>('database');

const sharedServicesStack = new StackReference(
  `${getOrganization()}/muehlbachler-shared-services/${environment}`,
);
const sharedServicesStackAws = sharedServicesStack.requireOutput('aws');
export const postgresqlConfig = sharedServicesStackAws.apply((output) => ({
  address: output.postgresql.address as string,
  port: output.postgresql.port as number,
  username: output.postgresql.username as string,
  password: output.postgresql.password as string,
}));

export const globalName = 'mail-services';

export const awsDefaultRegion = 'eu-west-1';

export const commonLabels = {
  environment: environment,
  purpose: globalName,
};
