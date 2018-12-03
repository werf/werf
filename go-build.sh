#!/bin/bash

set -e

go install github.com/flant/dapp/cmd/dappfile-yml
go install github.com/flant/dapp/cmd/git-artifact
go install github.com/flant/dapp/cmd/git-repo
go install github.com/flant/dapp/cmd/image
go install github.com/flant/dapp/cmd/builder
go install github.com/flant/dapp/cmd/dappdeps
go install github.com/flant/dapp/cmd/docker_registry
go install github.com/flant/dapp/cmd/cleanup
go install github.com/flant/dapp/cmd/slug
go install github.com/flant/dapp/cmd/deploy-watcher
go install github.com/flant/dapp/cmd/deploy
go install github.com/flant/dapp/cmd/build
