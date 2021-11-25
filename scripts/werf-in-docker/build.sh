#!/bin/bash

set -e

BUILD_VERSION="$(git rev-parse HEAD)"
IMAGE_NAME=ghcr.io/werf/werf:$BUILD_VERSION
docker build -f scripts/werf-in-docker/Dockerfile --build-arg build_version=$BUILD_VERSION -t $IMAGE_NAME .
