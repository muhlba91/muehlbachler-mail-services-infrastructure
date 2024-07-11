#!/bin/bash

# set env
echo "[init] setting env..."
set -o allexport && source /code/.env && set +o allexport

# needed for importing gpg keys
echo "[init] creating gpg directory..."
mkdir /sl/gnupg || true

# application initialization
echo "[init] initializing application..."
alembic upgrade head
python init_app.py

# set all users as premium users
echo "[init] set users as premium..."
apt-get update
apt-get install --yes postgresql-client
psql -c "UPDATE users SET lifetime = TRUE;"
