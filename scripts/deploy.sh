#!/usr/bin/env bash
set -euo pipefail

cd /opt/video-feedsystem

previous_commit=$(git rev-parse HEAD)
git pull --ff-only
if [[ $(git rev-parse HEAD) != "$previous_commit" ]]; then
  exec bash /opt/video-feedsystem/scripts/deploy.sh
fi

docker compose config --quiet
docker compose pull backend frontend prometheus alertmanager
docker compose up -d --no-build --pull never --force-recreate --wait --wait-timeout 120 backend frontend prometheus alertmanager

curl --fail --silent --show-error --max-time 10 http://127.0.0.1/api/ping
printf '\n'
curl --fail --silent --show-error --retry 5 --retry-delay 2 --retry-connrefused --max-time 10 http://127.0.0.1:9090/-/healthy
printf '\n'
curl --fail --silent --show-error --retry 5 --retry-delay 2 --retry-connrefused --max-time 10 http://127.0.0.1:9093/-/healthy
printf '\n'
docker compose ps
