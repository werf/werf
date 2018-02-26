#!/bin/bash

set -e

if [ ! -f dappdeps-toolchain.tar ] ; then
  docker pull dappdeps/toolchain:0.1.1
  docker save dappdeps/toolchain:0.1.1 -o dappdeps-toolchain.tar
fi
docker build -t dappdeps/ansible:2.4.4.0-1 .

docker login -u $DOCKER_HUB_LOGIN -p $DOCKER_HUB_PASSWORD || true
docker push dappdeps/ansible:2.4.4.0-1
