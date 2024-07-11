#!/bin/sh

### cron ###
chmod +x /bin/mailcow-backup
systemctl daemon-reload
systemctl restart cron
