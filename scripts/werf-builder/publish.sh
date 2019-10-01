#!/bin/bash

set -e

source scripts/werf-builder/version.sh

IMAGE_NAME=flant/werf-builder:$WERF_BUILDER_VERSION

docker push $IMAGE_NAME
