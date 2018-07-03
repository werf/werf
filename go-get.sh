#!/bin/bash

set -e

# pin go.uuid because sprig builds with error
# github.com/Masterminds/sprig/crypto.go:35: multiple-value uuid.NewV4() in single-value context
go get -v github.com/satori/go.uuid
path=${GOPATH%%:*}/src
git -C $path/github.com/satori/go.uuid checkout v1.2.0

go get -v github.com/flant/dapp/...
