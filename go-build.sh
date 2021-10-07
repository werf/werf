#!/bin/bash

set -e

CWD=`pwd`
SOURCE=`dirname ${BASH_SOURCE[0]}`

cd $SOURCE

export GO111MODULE=on
export CGO_ENABLED=1
#go install -ldflags="-linkmode=external -extldflags=-static" -tags "dfrunmount dfssh containers_image_openpgp osusergo" github.com/werf/werf/cmd/werf
#go install -compiler gccgo -gccgoflags="-static" -ldflags="-extldflags=-static" -tags "dfrunmount dfssh containers_image_openpgp osusergo" github.com/werf/werf/cmd/werf

go install -ldflags "-linkmode external -extldflags -static" -tags "dfrunmount dfssh containers_image_openpgp osusergo" github.com/werf/werf/cmd/buildah-test
#go install -gcflags="-static" -tags "dfrunmount dfssh containers_image_openpgp osusergo" github.com/werf/werf/cmd/buildah-test

#go install -tags "dfrunmount dfssh containers_image_openpgp" github.com/werf/werf/cmd/werf
#go install -tags "dfrunmount dfssh containers_image_openpgp" github.com/werf/werf/cmd/buildah-test

cd $CWD
