#!/usr/bin/env sh
#
# Spits out a list of plugin binaries we can build with "make" based on what's
# in src/plugins.  The ImageMagick decoder is explicitly skipped to avoid
# unnecessary dependencies since JP2s are the primary need.
for plugdir in $(find ./src/plugins -mindepth 1 -maxdepth 1 -type d -not -name "*magick*"); do
  echo bin/plugins/${plugdir##*/}.so
done
