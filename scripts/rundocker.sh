#!/bin/bash
#
# rundocker.sh: Provides an example of how to run the production docker
# container with various settings
#
# After starting the container, test RAIS by browsing to
# http://localhost:12415/iiif/test-world.jp2/full/full/0/default.jpg

# All possible environmental overrides are included below for clarity:
# - RAIS_ADDRESS: the address RAIS listens on, defaults to ":12415", responding to all requests on port 12415
# - RAIS_IIIFURL: what RAIS reports as its server URL, defaults to localhost:$PORT/iiif
# - RAIS_IIIFINFOCACHESIZE is the number of items in the info cache.  Each item is
#   very small, and 10,000 should use fewer than 5 megs of RAM.
# - RAIS_TILEPATH is where RAIS looks for images.  This defaults to
#   /var/local/images.
docker run -d \
  --name rais \
  --privileged=true \
  -e RAIS_ADDRESS=":12415" \
  -e RAIS_IIIFURL="http://localhost:12415/iiif" \
  -e RAIS_IIIFINFOCACHESIZE=10000 \
  -e RAIS_TILEPATH="/var/local/images" \
  -p 12415:12415 \
  -v $(pwd)/testfile:/var/local/images \
  uolibraries/rais

# If you want to use a config file rather than the environment variables, you
# might use this instead:
docker run -d \
  --name rais \
  --privileged=true \
  -p 12415:12415 \
  -v $(pwd)/testfile:/var/local/images \
  -v $(pwd)/rais-example.toml:/etc/rais.toml \
  uolibraries/rais
