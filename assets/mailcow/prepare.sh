#!/bin/sh

### mailcow ###
# create directories
mkdir -p /opt/backup/mailcow || true

# clone mailcow
git clone https://github.com/mailcow/mailcow-dockerized /opt/mailcow || true

# install pre-requisites
apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install --yes jq
