#!/bin/bash

set -e

BUILDAH_VERSION=v1.22.3
IMAGE_VERSION=$BUILDAH_VERSION-1

IMAGE_NAME=registry-write.werf.io/werf/buildah:$IMAGE_VERSION

docker build --build-arg version=$BUILDAH_VERSION -t $IMAGE_NAME scripts/buildah

docker push $IMAGE_NAME
