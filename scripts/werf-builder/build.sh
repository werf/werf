#!/bin/bash

set -e

IMAGE_NAME=ghcr.io/werf/builder:"$(git rev-parse HEAD)"
docker build -f scripts/werf-builder/Dockerfile -t $IMAGE_NAME .
