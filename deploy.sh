#!/bin/bash
#
# Compiles and deploys RAIS as a service
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
make bin/rais-server
sudo service rais stop
sudo rm /opt/chronam-support/rais-server
sudo mkdir -p /opt/chronam-support/
sudo cp rh_config/init.sh /etc/init.d/rais
sudo cp bin/rais-server /opt/chronam-support/rais-server
sudo service rais start
