#!/bin/bash
#
# rundocker.sh: Provides an example of how to run the docker container with various settings
#
# After starting the container, test RAIS by browsing to
# http://localhost:12415/iiif/test-world.jp2/full/full/0/default.jpg

# All possible environmental overrides are included below for clarity:
# - PORT: the port RAIS listens on, defaults to 12415
# - TILESIZES: what RAIS reports as valid IIIF tile sizes, defaults to 512
# - IIIFURL: what RAIS reports as its server URL, defaults to localhost:$PORT/iiif
docker run -d \
  --name rais \
  --privileged=true \
  -e PORT=12415 \
  -e TILESIZES=512,1024 \
  -e IIIFURL="http://localhost:12415/iiif" \
  -p 12415:12415 \
  -v $(pwd)/testfile:/var/local/images \
  uolibraries/rais
