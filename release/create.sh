#!/bin/bash

set -e

export VERSION=$1

if [ -z "$VERSION" ] ; then
  echo "Usage: $0 VERSION"
  echo
  exit 1
fi

DIR=$(dirname $0)

TAG_TEMPLATE=$DIR/git_tag_template.md

cat $TAG_TEMPLATE | envsubst | git tag --annotate --file - --edit $VERSION

git push --tags
