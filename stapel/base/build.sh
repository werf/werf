#!/bin/bash

set -e

DOCKER_IMAGE_VERSION=$(cat stapel/base/omnibus/config/projects/werf-stapel-base.rb | \
grep "DOCKER_IMAGE_VERSION =" | \
cut -d"=" -f2 | \
cut -d'"' -f2)

if [ ! -f stapel/base/werf-stapel-toolchain.tar ] ; then
  docker pull flant/werf-stapel-toolchain:0.2.0
  docker save flant/werf-stapel-toolchain:0.2.0 -o stapel/base/werf-stapel-toolchain.tar
fi

docker build -t flant/werf-stapel-base:$DOCKER_IMAGE_VERSION stapel/base

docker push flant/werf-stapel-base:$DOCKER_IMAGE_VERSION
