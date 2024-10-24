#!/usr/bin/env bash
#
# buildrun.sh: runs the build container with any extra parameters specified on
# the command line.  e.g., `./buildrun.sh make test`
docker compose -f compose.build.yml run --rm rais-build $@
