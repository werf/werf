#!/bin/bash

set -e

DOCKER_IMAGE_VERSION=$(cat stapel/ansible/omnibus/config/projects/werf-stapel-ansible.rb | \
grep "DOCKER_IMAGE_VERSION =" | \
cut -d"=" -f2 | \
cut -d'"' -f2)

if [ ! -f stapel/ansible/werf-stapel-toolchain.tar ] ; then
  docker pull flant/werf-stapel-toolchain:0.2.0
  docker save flant/werf-stapel-toolchain:0.2.0 -o stapel/ansible/werf-stapel-toolchain.tar
fi

docker build -t flant/werf-stapel-ansible:$DOCKER_IMAGE_VERSION stapel/ansible

docker push flant/werf-stapel-ansible:$DOCKER_IMAGE_VERSION
