#!/bin/bash

set -e

path=${GOPATH%%:*}/src

for os in linux darwin windows ; do
  for arch in amd64 ; do
    export GOOS=$os
    export GOARCH=$arch
    source go-get.sh
  done
done
