#!/bin/bash

set -e

if echo $TRAVIS_COMMIT_MESSAGE | grep -Pq "\[ci skip tests?\]" ; then
  exit 0
fi

dappdeps=(
  dappdeps/chefdk:2.3.17-2
  dappdeps/gitartifact:0.2.1
  dappdeps/base:0.2.1
  dappdeps/berksdeps:0.1.0
  dappdeps/toolchain:0.1.1
)
dappdeps_md5=`echo ${dappdeps[@]} | md5sum | awk '{ print $1 }'`

if [[ -e dappdeps/dappdeps.tar.gz ]] && [[ -e dappdeps/$dappdeps_md5 ]]
then
    gunzip -c dappdeps/dappdeps.tar.gz | docker load
else
    rm -rf dappdeps
    mkdir -p dappdeps
    touch dappdeps/$dappdeps_md5
    for image in ${dappdeps[@]}; do docker pull $image; done
    docker save ${dappdeps[@]} | gzip -c > dappdeps/dappdeps.tar.gz
fi
