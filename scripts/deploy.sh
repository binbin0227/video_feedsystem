#!/usr/bin/env bash
set -euo pipefail

cd /opt/video-feedsystem

git pull --ff-only
docker compose config --quiet
docker compose pull backend frontend
docker compose up -d --no-build --pull never --force-recreate --wait --wait-timeout 120 backend frontend

curl --fail --silent --show-error --max-time 10 http://127.0.0.1/api/ping
printf '\n'
docker compose ps
