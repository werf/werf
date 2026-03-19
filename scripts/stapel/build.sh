#!/bin/bash

set -e

REGISTRY="registry-write.werf.io/werf"
TAG="dev"
DOCKERFILE="stapel/Dockerfile"
PLATFORM="${STAPEL_PLATFORM:-}"

if [ -n "$PLATFORM" ]; then
    case "$PLATFORM" in
        linux/amd64)
            TARGETARCH="amd64"
            ;;
        linux/arm64)
            TARGETARCH="arm64"
            ;;
        *)
            echo "Unsupported STAPEL_PLATFORM: $PLATFORM" 1>&2
            exit 1
            ;;
    esac
else
    case "$(uname -m)" in
        x86_64|amd64)
            TARGETARCH="amd64"
            ;;
        arm64|aarch64)
            TARGETARCH="arm64"
            ;;
        *)
            echo "Unsupported host architecture: $(uname -m)" 1>&2
            exit 1
            ;;
    esac
fi

case "$TARGETARCH" in
    amd64)
        LFS_TGT="x86_64-lfs-linux-gnu"
        ;;
    arm64)
        LFS_TGT="aarch64-lfs-linux-gnu"
        ;;
    *)
        echo "Unsupported target architecture: $TARGETARCH" 1>&2
        exit 1
        ;;
esac

build() {
    local image=$1
    local target=$2

    local build_args=(
        --build-arg "TARGETARCH=${TARGETARCH}"
        --build-arg "LFS_TGT=${LFS_TGT}"
    )

    if [ -n "$PLATFORM" ]; then
        build_args+=(--platform "$PLATFORM")
    fi

    docker build -t "${REGISTRY}/${image}:${TAG}" --target "$target" --file "$DOCKERFILE" "${build_args[@]}" .
}

build stapel-base base
build stapel final
