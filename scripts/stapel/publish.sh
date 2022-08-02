#!/bin/bash

set -e

if [ -z "$1" ] ; then
	echo "Provide image version: $0 VERSION" 1>&2
	exit 1
fi

VERSION=$1

#docker tag registry-write.werf.io/werf/stapel-base:dev registry-write.werf.io/werf/stapel-base:$VERSION
#docker push registry-write.werf.io/werf/stapel-base:$VERSION

docker tag registry-write.werf.io/werf/stapel:dev registry-write.werf.io/werf/stapel:$VERSION
docker push registry-write.werf.io/werf/stapel:$VERSION
