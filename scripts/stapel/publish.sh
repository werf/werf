#!/bin/bash

set -e

if [ -z "$1" ] ; then
	echo "Provide image version: $0 VERSION" 1>&2
	exit 1
fi

VERSION=$1

#docker tag ghcr.io/werf/stapel-base:dev ghcr.io/werf/stapel-base:$VERSION
#docker push ghcr.io/werf/stapel-base:$VERSION

docker tag ghcr.io/werf/stapel:dev ghcr.io/werf/stapel:$VERSION
docker push ghcr.io/werf/stapel:$VERSION
