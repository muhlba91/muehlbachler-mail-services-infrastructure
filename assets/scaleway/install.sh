#!/bin/sh

### scaleway ###
curl -s https://raw.githubusercontent.com/scaleway/scaleway-cli/master/scripts/get.sh | sh
apt-get install -y rclone
