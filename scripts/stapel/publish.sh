#!/bin/bash

set -e

if [ -z "$1" ] ; then
	echo "Provide image version: $0 VERSION" 1>&2
	exit 1
fi

VERSION=$1

docker tag flant/werf-stapel:dev flant/werf-stapel:$VERSION
docker push flant/werf-stapel:$VERSION
