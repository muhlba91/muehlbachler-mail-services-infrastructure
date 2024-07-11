#!/bin/sh

### traefik ###
systemctl daemon-reload
systemctl enable traefik
systemctl restart traefik
