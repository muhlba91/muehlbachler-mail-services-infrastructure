#!/bin/sh

### cron ###
chmod +x /bin/ntfy-backup
systemctl daemon-reload
systemctl restart cron
