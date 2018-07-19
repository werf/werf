#!/bin/bash

set -e

path=${GOPATH%%:*}/src

# pin go.uuid because sprig builds with error
# github.com/Masterminds/sprig/crypto.go:35: multiple-value uuid.NewV4() in single-value context
go get -v github.com/satori/go.uuid
git -C $path/github.com/satori/go.uuid checkout v1.2.0

go get -v github.com/docker/cli/...
git -C $path/github.com/docker/cli checkout v18.06.0-ce-rc3

go get -v github.com/flant/dapp/...
