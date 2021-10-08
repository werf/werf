#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:?ERROR: Version should be specified as the first argument.}"

export GO111MODULE="on"

GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
  go build -compiler gc -o "release-build/$VERSION/linux-amd64/bin/werf" \
  -ldflags="-s -w -X github.com/werf/werf/pkg/werf.Version=$VERSION -linkmode external -extldflags=-static" \
  -tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" \
  github.com/werf/werf/cmd/werf

GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc \
  go build -compiler gc -o "release-build/$VERSION/linux-arm64/bin/werf" \
  -ldflags="-s -w -X github.com/werf/werf/pkg/werf.Version=$VERSION -linkmode external -extldflags=-static" \
  -tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" \
  github.com/werf/werf/cmd/werf

GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
  go build -o "release-build/$VERSION/darwin-amd64/bin/werf" \
  -ldflags="-s -w -X github.com/werf/werf/pkg/werf.Version=$VERSION"
  -tags="dfrunmount dfssh containers_image_openpgp" \
  github.com/werf/werf/cmd/werf

GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 \
  go build -o "release-build/$VERSION/darwin-arm64/bin/werf" \
  -ldflags="-s -w -X github.com/werf/werf/pkg/werf.Version=$VERSION"
  -tags="dfrunmount dfssh containers_image_openpgp" \
  github.com/werf/werf/cmd/werf

GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
  go build -o "release-build/$VERSION/windows-amd64/bin/werf.exe" \
  -ldflags="-s -w -X github.com/werf/werf/pkg/werf.Version=$VERSION"
  -tags="dfrunmount dfssh containers_image_openpgp" \
  github.com/werf/werf/cmd/werf
