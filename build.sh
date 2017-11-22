#!/bin/bash

set -e

if [ ! -f dappdeps-toolchain.tar ] ; then
  docker save dappdeps/toolchain:0.1.0 -o dappdeps-toolchain.tar
fi
docker build -t dappdeps/chefdk:2.3.17-1 .

docker login -u $DOCKER_HUB_LOGIN -p $DOCKER_HUB_PASSWORD || true
docker push dappdeps/chefdk:2.3.17-1
