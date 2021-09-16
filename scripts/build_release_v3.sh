#!/bin/bash

set -e

VERSION=$1
if [ -z "$VERSION" ] ; then
    echo "Required version argument!" 1>&2
    echo 1>&2
    echo "Usage: $0 VERSION" 1>&2
    exit 1
fi

export CGO_ENABLED=0

gox -osarch="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64" \
    -output="release-build/$VERSION/{{.OS}}-{{.Arch}}/bin/werf" \
    -tags="dfrunmount dfssh" \
    -ldflags="-s -w -X github.com/werf/werf/pkg/werf.Version=$VERSION" \
        github.com/werf/werf/cmd/werf
