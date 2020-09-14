#!/usr/bin/env sh

# demo-rais-entry.sh is the RAIS entrypoint script, which sets up the
# configuration and runs the rais server

# Copy the config and edit it in-place; this allows customizing most pieces of
# configuration for demoing
url=${URL:-}
if test "$url" = ""; then
  echo "No URL provided; defaulting to 'http://localhost'"
  echo "If you can't see images, try an explicitly-set URL, e.g.:"
  echo
  echo "    URL="http://192.168.0.5" docker-compose up"
  url="http://localhost"
fi

exec /opt/rais/rais-server
