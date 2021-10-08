#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "$0")"; pwd)"

export GO111MODULE=on

cd "$script_dir"
if [[ "$(uname)" == "Linux" ]]; then
  CGO_ENABLED=1 go install -compiler gc \
    -ldflags="-linkmode external -extldflags=-static" \
    -tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" \
    github.com/werf/werf/cmd/werf
elif [[ "$(uname)" == "Darwin" ]]; then
  CGO_ENABLED=0 go install \
    -tags="dfrunmount dfssh containers_image_openpgp" \
    github.com/werf/werf/cmd/werf
else
  CGO_ENABLED=0 go install \
    -tags="dfrunmount dfssh containers_image_openpgp" \
    github.com/werf/werf/cmd/werf
fi
cd -
