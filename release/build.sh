#!/bin/bash

set -e

VERSION=$1

if [ -z "$VERSION" ] ; then
  echo "Usage: $0 VERSION"
  echo
  exit 1
fi

RELEASE_BUILD_DIR=$(pwd)/release/build

rm -rf $RELEASE_BUILD_DIR

for arch in linux darwin ; do
  outputDir=$RELEASE_BUILD_DIR/$arch-amd64

  mkdir -p $outputDir

  echo "Building dapp for $arch, version $VERSION"
  GOOS=$arch GOARCH=amd64 go build -ldflags="-s -w -X github.com/flant/dapp/pkg/dapp.Version=$VERSION" -o $outputDir/dapp github.com/flant/dapp/cmd/dapp

  echo "Calculating checksum dapp.sha"
  sha256sum $outputDir/dapp | cut -d' ' -f 1 > $outputDir/dapp.sha
done
