#!/bin/bash

set -e

apt-get update -qq
apt-get install -y git autoconf build-essential devscripts curl


git config --global user.email timofey.kirillov@flant.com
git config --global user.name 'Timofey Kirillov'
gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3
curl -sSL https://get.rvm.io | bash
source /etc/profile.d/rvm.sh
rvm install 2.3.1
gem install bundler
cd /vagrant
bundle install --without development
