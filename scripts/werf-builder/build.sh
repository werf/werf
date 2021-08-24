#!/bin/bash

set -e

IMAGE_NAME=ghcr.io/werf/builder:latest
docker build -f scripts/werf-builder/Dockerfile -t $IMAGE_NAME .
