#!/usr/bin/env bash
set -euo pipefail

export GO111MODULE=on

if [[ "$(uname)" == "Linux" ]]; then
  CGO_ENABLED=1 go test -compiler gc \
    -ldflags="-linkmode external -extldflags=-static" \
    -tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" \
    $@
elif [[ "$(uname)" == "Darwin" ]]; then
  CGO_ENABLED=0 go test \
    -tags="dfrunmount dfssh containers_image_openpgp" \
    $@
else
  CGO_ENABLED=0 go test \
    -tags="dfrunmount dfssh containers_image_openpgp" \
    $@
fi
