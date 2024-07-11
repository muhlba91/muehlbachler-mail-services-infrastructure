#!/bin/sh

### mailcow ###
# create directories
mkdir -p /opt/backup/mailcow

# clone mailcow
git clone https://github.com/mailcow/mailcow-dockerized /opt/mailcow
