#!/bin/bash

set -e

go get -v github.com/satori/go.uuid
git -C ../../satori/go.uuid/ checkout v1.2.0

go get -v github.com/flant/dapp/...
