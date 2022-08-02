#!/bin/bash

set -e

IMAGE_NAME=registry-write.werf.io/werf/builder:"$(git rev-parse HEAD)"
docker push $IMAGE_NAME
