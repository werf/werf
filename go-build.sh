#!/bin/bash

set -e

CWD=`pwd`
SOURCE=`dirname ${BASH_SOURCE[0]}`

cd $SOURCE

export GO111MODULE=on
export CGO_ENABLED=0
go install -tags "dfrunmount dfssh containers_image_openpgp" github.com/werf/werf/cmd/werf

cd $CWD
