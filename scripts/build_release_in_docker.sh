#!/bin/bash

docker run -ti --rm --volume $GOPATH/mod:/go/mod --volume $GOPATH/pkg:/go/pkg --volume $(pwd):/werf --workdir /werf golang:1.14 scripts/do_build_release.sh "$@"
