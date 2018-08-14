#!/bin/bash
set -eu

docker rm -f rais || true
docker run --rm -v $(pwd):/opt/rais-src rais-build:f28 make
docker run -it --rm \
  --name rais \
  --privileged=true \
  -p 12415:12415 \
  -v $(pwd)/docker/images:/var/local/images \
  -v $(pwd):/opt/rais-src \
  rais-build:f28 /opt/rais-src/bin/rais-server
