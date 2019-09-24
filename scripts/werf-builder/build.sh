#!/bin/bash

set -e

source scripts/werf-builder/version.sh

IMAGE_NAME=flant/werf-builder:$WERF_BUILDER_VERSION

docker build -f scripts/werf-builder/Dockerfile -t $IMAGE_NAME .
