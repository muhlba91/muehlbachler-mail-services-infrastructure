#!/bin/sh

### cron ###
chmod +x /bin/simplelogin-backup
systemctl daemon-reload
systemctl restart cron
