#!/bin/bash
set -eu

tagprod() {
  docker rmi "$1" || true
  docker tag uolibraries/rais:prod "$1"
}

make docker

ver=$(grep Version ./src/version/version.go | sed 's/^.*"\(.*\)".*$/\1/')
tagprod "uolibraries/rais:$ver-$(date +"%Y-%m-%d")"
tagprod "uolibraries/rais:latest-dev"
