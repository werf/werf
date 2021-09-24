#!/bin/bash

set -e

VERSION=v1.22.3

IMAGE_NAME=ghcr.io/werf/buildah:$VERSION

docker build --build-arg version=$VERSION -t $IMAGE_NAME .

docker push $IMAGE_NAME
