#!/bin/bash
#
# docker.sh: Removes and rebuilds the rais container with an example of how to
# run it with various settings.
#
# After starting the container, test RAIS by browsing to
# http://localhost:12415/iiif/test-world.jp2/full/full/0/default.jpg
set -eu

# Cleanup running rais container
docker stop rais || true
docker rm rais || true

# Build the libs image: fedora + jp2/imagemagick libraries
docker build --rm -t uolibraries/rais:libs -f Dockerfile.libs .

# Build the "build" image: compiles our go binary on the libs image
docker build --rm -t uolibraries/rais:build -f Dockerfile.build ..

# Grab the built server image into our local filesystem
mkdir -p $(pwd)/bin
docker run --rm -it -v $(pwd)/bin:/tmp/hostbin uolibraries/rais:build cp /tmp/go/bin/rais-server /tmp/hostbin

# Build the production image from core and copy over the binary
docker build --rm -t uolibraries/rais -f Dockerfile.prod .

# Remove the binary directory to avoid copying old files if things break in a subsequent build
rm -rf $(pwd)/bin

# All possible environmental overrides are included below for clarity:
# - PORT: the port RAIS listens on, defaults to 12415
# - TILESIZES: what RAIS reports as valid IIIF tile sizes, defaults to 512
# - IIIFURL: what RAIS reports as its server URL, defaults to localhost:$PORT/iiif
cd .. && docker run -d \
  --name rais \
  -e PORT=12415 \
  -e TILESIZES=512,1024 \
  -e IIIFURL="http://localhost:12415/iiif" \
  -p 12415:12415 \
  -v $(pwd)/testfile:/var/local/images \
  uolibraries/rais
