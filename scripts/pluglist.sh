#!/usr/bin/env sh
#
# Spits out a list of plugin binaries we can build with "make" based on what's in src/plugins
for plugdir in $(find ./src/plugins -mindepth 1 -maxdepth 1 -type d); do
  echo bin/plugins/${plugdir##*/}.so
done
