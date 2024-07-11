#!/bin/sh

### update system ###
apt-get update
DEBIAN_FRONTEND=noninteractive apt-get upgrade --yes


### docker ###
# add docker repository
DEBIAN_FRONTEND=noninteractive apt-get install ca-certificates curl
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
chmod a+r /etc/apt/keyrings/docker.asc
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  tee /etc/apt/sources.list.d/docker.list > /dev/null
DEBIAN_FRONTEND=noninteractive apt-get update

# install docker
DEBIAN_FRONTEND=noninteractive apt-get install --yes docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# daemon.json
cat << EOF > /etc/docker/daemon.json
{{ daemonJson }}
EOF

# start docker
systemctl enable docker
systemctl restart docker
