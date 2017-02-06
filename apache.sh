#!/bin/bash
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

sed "s|%SRCS%|$sources|g" docker/apache/index.template > docker/apache/index.html

docker rm -f "rais-osd-example" || true
docker rm -f "rais-test" || true
docker run --rm -v $(pwd):/opt/rais-src rais-build make

docker run -d -it --name "rais-osd-example" -p 80:80 \
  -v $(pwd)/docker/apache:/usr/local/apache2/htdocs/ \
  httpd:2.4
docker run -it --rm --name "rais-test" --privileged=true -p 12415:12415 \
  -v $(pwd):/opt/rais-src \
  -v $(pwd)/docker/images:/var/local/images \
  rais-build /opt/rais-src/bin/rais-server --tile-path /var/local/images --iiif-url $url:12415/images/iiif
