# Mail Services - Infrastructure

[![Build status](https://img.shields.io/github/actions/workflow/status/muhlba91/muehlbachler-mail-services-infrastructure/pipeline.yml?style=for-the-badge)](https://github.com/muhlba91/muehlbachler-mail-services-infrastructure/actions/workflows/pipeline.yml)
[![License](https://img.shields.io/github/license/muhlba91/muehlbachler-mail-services-infrastructure?style=for-the-badge)](LICENSE.md)

This repository contains the infrastructure as code (IaC) for mail services using [Pulumi](http://pulumi.com).

---

## Requirements

- [NodeJS](https://nodejs.org/en), and [yarn](https://yarnpkg.com)
- [Pulumi](https://www.pulumi.com/docs/install/)

## Creating the Infrastructure

To create the infrastructure and deploy the virtual machine, a [Pulumi Stack](https://www.pulumi.com/docs/concepts/stack/) with the correct configuration needs to exists.

The stack can be deployed via:

```bash
yarn install
yarn build; pulumi up
```

## Destroying the Infrastructure

The entire infrastructure can be destroyed via:

```bash
yarn install
yarn build; pulumi destroy
```

## Environment Variables

To successfully run, and configure the Pulumi plugins, you need to set a list of environment variables. Alternatively, refer to the used Pulumi provider's configuration documentation.

- `CLOUDSDK_CORE_PROJECT`: the Google Cloud (GCP) project
- `CLOUDSDK_COMPUTE_REGION` the Google Cloud (GCP) region
- `GOOGLE_APPLICATION_CREDENTIALS`: reference to a file containing the Google Cloud (GCP) service account credentials
- `HCLOUD_TOKEN`: the token to interact with Hetzner Cloud

---

## Configuration

The following section describes the configuration which must be set in the Pulumi Stack.

***Attention:*** do use [Secrets Encryption](https://www.pulumi.com/docs/concepts/secrets/#:~:text=Pulumi%20never%20sends%20authentication%20secrets,“secrets”%20for%20extra%20protection.) provided by Pulumi for secret values!

### Network

```yaml
network:
  name: the network name to use (if it exists, `cidr` is ignored)
  cidr: the network CIDR
  subnetCidr: the subnet CIDR (must be within the network CIDR `cidr`)
```

### Server

```yaml
server:
  type: the Hetzner cloud server type/size
  ip: the internal IP address (must be within the subnet CIDR `network.subnetCidr`)
  publicSsh: connect to the server through its public ip address (`true`) or private ip address (`false`) (optional, default: `false`)
```

### Mail

```yaml
mail:
  main: the main domain of the mail server (mailname will be `mail.<DOMAIN_NAME>`)
    name: the domain
    zoneId: the zone identifier in Google Cloud to set the DNS entries
    project: the Google Cloud project where the zone is located (optional)
  additional: additional domains to use (optional)
    name: the domain
    zoneId: the zone identifier in Google Cloud to set the DNS entries
    project: the Google Cloud project where the zone is located (optional)
  dkimSignHeaders: the list of headers to sign with DKIM (see note below)
```

When using an outbound relay, the e-mail will be signed twice with DKIM.
Usually, this doesn't create any problems. However, to increase compatibility it's advised to skip signing `message-id` and `date`.
You can define the list of headers to signed in `dkimSignHeaders`.

### DNS

```yaml
dns:
  project: the Google Cloud project where the zone is located (will be overwritten by each `mail.XXX.project` if set)
  email: the e-mail address for ACME to use
```

### Database

A database is created with the corresponding user.

```yaml
database:
  databases: a map of databases to create and their owner
  users: a list of users to create
```

### SimpleLogin

```yaml
simplelogin:
  domain: the URL/domain for the SimpleLogin web interface
  mail: the email domain configuration
    domain: the domain to use for relaying emails
    mx: the expected MX record name
    zoneId: the Google Cloud zone identifier (optional)
    project: the Google Cloud project (optional)
  oidc: the OIDC configuration
    wellKnownUrl: the well-known URL to set
    clientId: the OIDC client id for the application
    clientSecret: the OIDC client secret for the application
```

### Bucket

```yaml
bucketId: the bucket identifier to store output assets in
backupBucketId: the backup bucket identifier
```

---

## Continuous Integration and Automations

- [GitHub Actions](https://docs.github.com/en/actions) are linting, and verifying the code.
- [Renovate Bot](https://github.com/renovatebot/renovate) is updating NodeJS packages, and GitHub Actions.
