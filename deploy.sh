#!/bin/bash
#
# Compiles and deploys the tile server as a service
#
# This is meant as an example for going from development to production.  MANY
# assumptions are made:
#
# - You've already installed the service once and done `chkconfig` magic
# - You are using a RedHat-6-based system
# - You have Go installed on your production system
# - You have sudo access
# - You are using this with chronam

make clean
make bin/jp2tileserver
sudo service tileserver stop
sudo rm /opt/chronam-support/jp2tileserver
sudo mkdir -p /opt/chronam-support/
sudo cp rh_config/init.sh /etc/init.d/tileserver
sudo cp bin/jp2tileserver /opt/chronam-support/jp2tileserver
sudo service tileserver start
