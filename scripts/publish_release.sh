#!/bin/bash

set -e

for f in $(find scripts/lib -type f -name "*.sh"); do
    source $f
done

VERSION=$1
if [ -z "$VERSION" ] ; then
    echo "Required version argument!" 1>&2
    echo 1>&2
    echo "Usage: $0 VERSION" 1>&2
    exit 1
fi

if [ -z "$PUBLISH_BINTRAY_AUTH"  ] ; then
    echo "\$PUBLISH_BINTRAY_AUTH required!" 1>&2
    exit 1
fi

if [ -z "$PUBLISH_GITHUB_TOKEN"  ] ; then
    echo "\$PUBLISH_GITHUB_TOKEN required!" 1>&2
    exit 1
fi

( which git > /dev/null ) || ( echo "Cannot find git command!" 1>&2 && exit 1 )
( which curl > /dev/null ) || ( echo "Cannot find curl command!" 1>&2 && exit 1 )

# ( go_build $VERSION ) || ( echo "Failed to build!" 1>&2 && exit 1 )
# ( publish_binaries $VERSION ) || ( echo "Failed to publish release binaries!" 1>&2 && exit 1 )
( create_github_release $VERSION ) || ( echo "Failed to create github release!" 1>&2 && exit 1 )
