#!/bin/bash
set -eu

docker compose down
docker compose run --rm rais-build make clean
docker compose run --rm rais-build make
docker compose up -d rais
docker compose logs -f rais
