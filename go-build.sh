#!/bin/bash

set -e

CWD=`pwd`
SOURCE=`dirname ${BASH_SOURCE[0]}`

cd $SOURCE

export GO111MODULE=on
export CGO_ENABLED=1
go install -compiler gc -ldflags="-linkmode external -extldflags=-static" -tags="dfrunmount dfssh osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" github.com/werf/werf/cmd/werf
go install -compiler gc -ldflags="-linkmode external -extldflags=-static" -tags="dfrunmount dfssh osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" github.com/werf/werf/cmd/buildah-test

cd $CWD
