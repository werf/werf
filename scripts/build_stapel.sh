#!/bin/bash

set -e

IMAGE_VERSION=0.1.2

docker build -t flant/werf-stapel:$IMAGE_VERSION stapel

docker push flant/werf-stapel:$IMAGE_VERSION
