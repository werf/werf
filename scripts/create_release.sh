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

if [ -z "$BINTRAY_AUTH"  ] ; then
    echo "\$BINTRAY_AUTH required!" 1>&2
    exit 1
fi

if [ -z "$GITHUB_TOKEN"  ] ; then
    echo "\$GITHUB_TOKEN required!" 1>&2
    exit 1
fi

( which git > /dev/null ) || ( echo "Cannot find git command!" 1>&2 && exit 1 )
( which curl > /dev/null ) || ( echo "Cannot find curl command!" 1>&2 && exit 1 )

( create_release_message $VERSION ) || ( echo "Failed to create release message!" 1>&2 && exit 1 )

docker run --rm \
    --env SSH_AUTH_SOCK=$SSH_AUTH_SOCK \
    --volume $SSH_AUTH_SOCK:$SSH_AUTH_SOCK \
    --volume ~/.ssh/known_hosts:/root/.ssh/known_hosts \
    --volume $(pwd):/werf \
    --workdir /werf \
    flant/werf-builder:1.2.0 \
    bash -ec "set -e; source scripts/lib/release/global_data.sh && source scripts/lib/release/build.sh && go_build $VERSION"

( publish_binaries $VERSION ) || ( echo "Failed to publish release binaries!" 1>&2 && exit 1 )
( sign_binaries $VERSION ) || ( echo "Failed to sign release binaries!" 1>&2 && exit 1 )
( bintray_publish_files_in_version "$VERSION" ) || ( echo "Failed to publish uploaded files in version $VERSION" 1>&2 && exit 1 )
( create_github_release $VERSION ) || ( echo "Failed to create github release!" 1>&2 && exit 1 )
