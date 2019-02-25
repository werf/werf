#!/bin/bash

set -e

docker build -t flant/werf-stapel-toolchain:0.2.0 stapel/toolchain
docker login -u $DOCKER_HUB_LOGIN -p $DOCKER_HUB_PASSWORD || true
docker push flant/werf-stapel-toolchain:0.2.0
