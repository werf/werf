#!/bin/bash

set -e

REGISTRY="registry-write.werf.io/werf"
TAG="dev"
DOCKERFILE="stapel/Dockerfile"

build() {
    local image=$1
    local target=$2
    docker build -t "${REGISTRY}/${image}:${TAG}" --target "$target" --file "$DOCKERFILE" .
}

build stapel-base base
build stapel final
