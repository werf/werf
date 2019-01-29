#!/bin/bash

set -e

VERSION=$1

if [ -z "$VERSION" ] ; then
  echo "Usage: $0 VERSION"
  echo
  exit 1
fi

RELEASE_BUILD_DIR=release/build/

rm -rf $RELEASE_BUILD_DIR/$VERSION
mkdir -p $RELEASE_BUILD_DIR/$VERSION

for os in linux darwin windows ; do
  for arch in amd64 ; do
    outputFile=$RELEASE_BUILD_DIR/$VERSION/werf-$os-$arch-$VERSION
    if [ "$os" == "windows" ] ; then
      outputFile=$outputFile.exe
    fi

    echo "# Building werf $VERSION for $os $arch ..."

    GOOS=$os GOARCH=$arch \
      go build -ldflags="-s -w -X github.com/flant/werf/pkg/werf.Version=$VERSION" \
               -o $outputFile github.com/flant/werf/cmd/werf

    echo "# Built $outputFile"
  done
done

cd $RELEASE_BUILD_DIR/$VERSION/
sha256sum werf-* > SHA256SUMS
cd -
