#!/bin/bash
set -eu

docker rm -f rais || true
docker run --rm -v $(pwd):/opt/rais-src rais-build make
docker run -it --rm \
  --name rais \
  --privileged=true \
  -p 12415:12415 \
  -v /mnt/news2/outgoing/batch_oru_20160329203731_ver01/data/sn00063676/print/2015080701:/var/local/images/tmp \
  -v $(pwd):/opt/rais-src \
  rais-build /opt/rais-src/bin/rais-server --address ":12415" --iiif-url "http://localhost:12415/iiif" --iiif-tile-sizes "512,1024" --tile-path "/var/local/images"
