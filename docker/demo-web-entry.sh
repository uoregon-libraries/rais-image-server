#!/usr/bin/env bash

# demo-apache-entry.sh is the Apache entrypoint script, which scans for images
# and writes out HTML files to include all images in the RAIS / Open Seadragon
# demo

# Copy the templates
cp -r /static/* /usr/share/nginx/html

# Insert a tile source for every file found under /var/local/images
sources=""
for file in $(find /var/local/images -name "*.jp2" -o -name "*.tiff" -o -name "*.jpg"); do
  relpath=${file##/var/local/images/}
  relpath=${relpath//\//%2F}
  if [[ $sources != "" ]]; then
    sources="$sources,"
  fi

  sources="$sources\"/images/iiif/$relpath/info.json\""
done
sed "s|%SRCS%|$sources|g" /usr/share/nginx/html/template.html > /usr/share/nginx/html/iiif.html

sources=""
for file in $(find /var/local/images -name "*.jp2" -o -name "*.tiff" -o -name "*.jpg"); do
  relpath=${file##/var/local/images/}
  relpath=${relpath//\//%2F}
  if [[ $sources != "" ]]; then
    sources="$sources,"
  fi

  sources="$sources\"/images/dzi/${relpath}.dzi\""
done
sed "s|%SRCS%|$sources|g" /usr/share/nginx/html/template.html > /usr/share/nginx/html/dzi.html

nginx -g "daemon off;"
