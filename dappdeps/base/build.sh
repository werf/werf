#!/bin/bash

set -e

DOCKER_IMAGE_VERSION=$(cat omnibus/config/projects/dappdeps-base.rb | \
grep "DOCKER_IMAGE_VERSION =" | \
cut -d"=" -f2 | \
cut -d'"' -f2)

if [ ! -f dappdeps-toolchain.tar ] ; then
  docker pull dappdeps/toolchain:0.1.1
  docker save dappdeps/toolchain:0.1.1 -o dappdeps-toolchain.tar
fi
docker build -t dappdeps/base:$DOCKER_IMAGE_VERSION .

docker login -u $DOCKER_HUB_LOGIN -p $DOCKER_HUB_PASSWORD || true
docker push dappdeps/base:$DOCKER_IMAGE_VERSION
