#!/bin/bash

set -e

path=${GOPATH%%:*}/src

# pin go.uuid because sprig builds with error
# github.com/Masterminds/sprig/crypto.go:35: multiple-value uuid.NewV4() in single-value context
go get -v github.com/satori/go.uuid
git -C $path/github.com/satori/go.uuid checkout v1.2.0

go get -v github.com/docker/cli/...
git -C $path/github.com/docker/cli fetch
git -C $path/github.com/docker/cli checkout v18.06.3-ce

go get -u -v github.com/flant/kubedog/...
go get -u -v github.com/flant/logboek_py/...

go get -v github.com/flant/logboek/...
git -C $path/github.com/flant/logboek fetch
git -C $path/github.com/flant/logboek checkout 91fee3a3fc8d106213eb918ead7256eba41df426

go get -v github.com/flant/werf/cmd/werf
