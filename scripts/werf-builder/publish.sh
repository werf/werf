#!/bin/bash

set -e

IMAGE_NAME=ghcr.io/werf/builder:latest
docker push $IMAGE_NAME
