#!/bin/bash

set -e

DOCKER_IMAGE_VERSION=$(cat omnibus/config/projects/dappdeps-ansible.rb | \
grep "DOCKER_IMAGE_VERSION =" | \
cut -d"=" -f2 | \
cut -d'"' -f2)

if [ ! -f dappdeps-toolchain.tar ] ; then
  docker pull dappdeps/toolchain:0.1.1
  docker save dappdeps/toolchain:0.1.1 -o dappdeps-toolchain.tar
fi

echo "# Building dappdeps/ansible:$DOCKER_IMAGE_VERSION"
echo "# How to change versions:"
echo "#   * Change docker image version DOCKER_IMAGE_VERSION in omnibus/config/projects/dappdeps-ansible.rb"
echo "#   * Change ansible source code version tag ANSIBLE_GIT_TAG in omnibus/config/software/ansible.rb"
echo
docker build -t dappdeps/ansible:$DOCKER_IMAGE_VERSION .

docker login -u $DOCKER_HUB_LOGIN -p $DOCKER_HUB_PASSWORD || true
docker push dappdeps/ansible:$DOCKER_IMAGE_VERSION
