#!/bin/bash
#
# Compiles and deploys RAIS as a service
#
# This is meant as an example for going from development to production.  MANY
# assumptions are made:
#
# - You've already installed the service once
# - You are using a RedHat-7-based system
# - You have Go installed on your production system
# - You have sudo access
# - You are using this with ONI

set -eu

make clean
make
sudo systemctl stop rais
sudo rm -f /usr/local/rais/rais-server
sudo mkdir -p /usr/local/rais
sudo cp rh_config/rais.service /usr/local/rais/rais.service

if [ ! -f /etc/rais.toml ]; then
  sudo cp rais-example.toml /etc/rais.toml
  echo "New install detected - modify /etc/rais.toml as necessary"
fi

sudo cp bin/rais-server /usr/local/rais/rais-server
sudo systemctl daemon-reload
sudo systemctl start rais
