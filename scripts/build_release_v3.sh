#!/usr/bin/env bash
set -xeuo pipefail

VERSION="${1:?ERROR: Version should be specified as the first argument.}"

export GO111MODULE="on"

COMMON_LDFLAGS="-s -w -X github.com/werf/werf/pkg/werf.Version=$VERSION"
COMMON_TAGS="dfrunmount dfssh containers_image_openpgp"
PKG="github.com/werf/werf/cmd/werf"

parallel -j0 --halt now,fail=1 --line-buffer -k --tag --tagstring '{= @cmd = split(" ", $_); $_ = "[".$cmd[0]." ".$cmd[1]."]" =}' <<-EOF
  GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
    go build -compiler gc -o "release-build/$VERSION/linux-amd64/bin/werf" \
    -ldflags="$COMMON_LDFLAGS -linkmode external -extldflags=-static" \
    -tags="$COMMON_TAGS osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" \
    "$PKG"

  GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc \
    go build -compiler gc -o "release-build/$VERSION/linux-arm64/bin/werf" \
    -ldflags="$COMMON_LDFLAGS -linkmode external -extldflags=-static" \
    -tags="$COMMON_TAGS osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" \
    "$PKG"

  GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
    go build -o "release-build/$VERSION/darwin-amd64/bin/werf" -ldflags="$COMMON_LDFLAGS" -tags="$COMMON_TAGS" "$PKG"

  GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 \
    go build -o "release-build/$VERSION/darwin-arm64/bin/werf" -ldflags="$COMMON_LDFLAGS" -tags="$COMMON_TAGS" "$PKG"

  GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
    go build -o "release-build/$VERSION/windows-amd64/bin/werf.exe" -ldflags="$COMMON_LDFLAGS" -tags="$COMMON_TAGS" "$PKG"
EOF
