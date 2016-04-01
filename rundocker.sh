#!/bin/bash
#
# rundocker.sh: Provides an example of how to run the production docker
# container with various settings
#
# After starting the container, test RAIS by browsing to
# http://localhost:12415/iiif/test-world.jp2/full/full/0/default.jpg

# All possible environmental overrides are included below for clarity:
# - PORT: the port RAIS listens on, defaults to 12415
# - IIIFURL: what RAIS reports as its server URL, defaults to localhost:$PORT/iiif
# - IIIFINFOCACHESIZE is the number of items in the info cache.  Each item is
#   very small, and 10,000 should use fewer than 5 megs of RAM.
docker run -d \
  --name rais \
  --privileged=true \
  -e PORT=12415 \
  -e IIIFURL="http://localhost:12415/iiif" \
  -e IIIFINFOCACHESIZE=10000
  -p 12415:12415 \
  -v $(pwd)/testfile:/var/local/images \
  uolibraries/rais
