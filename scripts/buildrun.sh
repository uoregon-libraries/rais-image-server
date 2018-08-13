#!/usr/bin/env bash
#
# buildrun.sh: runs the build container with any extra parameters specified on
# the command line.  e.g., `./buildrun.sh make test`
docker run --rm -v $(pwd):/opt/rais-src rais-build:f28 $@
