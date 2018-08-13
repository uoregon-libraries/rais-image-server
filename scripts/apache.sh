#!/bin/bash
#
# apache.sh finds all files in the docker/images directory and builds an
# Apache-enabled container running Open Seadragon and pointing to a container
# running the RAIS build image.  Each JP2 or TIFF found in docker/images will
# be exposed via IIIF and DeepZoom endpoints as a way to verify both protocols
# semi-side-by-side.
set -eu

# Gotta have a URL or else default to localhost
url=${1:-}
if [[ $url == "" ]]; then
  echo "No URL provided; defaulting to 'http://localhost'"
  echo "If you can't see images, try an explicitly-set URL"
  echo
  echo "(e.g., './apache.sh http://192.168.0.5')"
  url="http://localhost"
fi

# Insert a tile source for every file found under docker/images
sources=""
for file in $(find docker/images -name "*.jp2" -o -name "*.tiff"); do
  relpath=${file##docker/images/}
  relpath=${relpath//\//%2F}
  if [[ $sources != "" ]]; then
    sources="$sources,"
  fi

  sources="$sources\"$url:12415/images/iiif/$relpath/info.json\""
done
sed "s|%SRCS%|$sources|g" docker/apache/template.html > docker/apache/iiif.html

sources=""
for file in $(find docker/images -name "*.jp2" -o -name "*.tiff"); do
  relpath=${file##docker/images/}
  relpath=${relpath//\//%2F}
  if [[ $sources != "" ]]; then
    sources="$sources,"
  fi

  sources="$sources\"$url:12415/images/dzi/${relpath}.dzi\""
done
sed "s|%SRCS%|$sources|g" docker/apache/template.html > docker/apache/dzi.html

docker rm -f "rais-osd-example" || true
docker rm -f "rais-test" || true

cp ./rais-example.toml ./temprais.toml
sed -i 's|^\s*IIIFURL.*$|IIIFURL = "'$url:12415'/images/iiif"|' temprais.toml

docker run -d -it --name "rais-osd-example" -p 80:80 \
  -v $(pwd)/docker/apache:/usr/local/apache2/htdocs/ \
  httpd:2.4
docker run -it --rm --name "rais-test" --privileged=true -p 12415:12415 \
  -v $(pwd):/opt/rais-src \
  -v $(pwd)/docker/images:/var/local/images \
  -v $(pwd)/temprais.toml:/etc/rais.toml \
  uolibraries/rais
