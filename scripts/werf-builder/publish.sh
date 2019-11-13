#!/bin/bash

set -e

if [ -z "$1" ] ; then
	echo "Provide new werf-builder version: $0 VERSION!" 1>&2
	exit 1
fi
NEW_VERSION=$1

IMAGE_NAME=flant/werf-builder:$NEW_VERSION

docker push $IMAGE_NAME
