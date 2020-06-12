#!/bin/bash

set -e

CWD=`pwd`
SOURCE=`dirname ${BASH_SOURCE[0]}`

cd $SOURCE

export GO111MODULE=on
go install -tags "dfrunmount dfssh" github.com/werf/werf/cmd/werf

cd $CWD
