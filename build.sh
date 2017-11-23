#!/bin/bash

set -e

docker build -t dappdeps/toolchain:0.1.1 .
docker login -u $DOCKER_HUB_LOGIN -p $DOCKER_HUB_PASSWORD || true
docker push dappdeps/toolchain:0.1.1
