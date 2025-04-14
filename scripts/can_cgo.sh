#!/usr/bin/env bash

if [[ $(go env CGO_ENABLED) != '1' ]]; then
  echo "Your system cannot build RAIS. It appears that there may not be a C compiler,"
  echo "which is required for the openjpeg bindings. Install gcc, clang, or similar"
  echo "and try again."

  exit 1
fi

exit 0
