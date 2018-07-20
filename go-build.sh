#!/bin/bash

set -e

go install github.com/flant/dapp/cmd/dappfile-yml
go install github.com/flant/dapp/cmd/git-artifact
go install github.com/flant/dapp/cmd/git-repo
go install github.com/flant/dapp/cmd/image
go install github.com/flant/dapp/cmd/builder
