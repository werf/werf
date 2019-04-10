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
go get -u -v github.com/flant/logboek/...
go get -u -v github.com/flant/logboek_py/...

if [ ! -d "$path/k8s.io/helm" ]; then
  mkdir -p $path/k8s.io
  git clone https://github.com/helm/helm $path/k8s.io/helm
else
  git -C $path/k8s.io/helm fetch
  git -C $path/k8s.io/helm checkout master
  git -C $path/k8s.io/helm reset --hard origin/master
fi

cwd=`pwd`
cd $path/k8s.io/helm
make bootstrap
find . -type f -regex './vendor/golang.org/x/net/trace/.*' -delete
cd $cwd

go get -v github.com/flant/werf/cmd/werf
