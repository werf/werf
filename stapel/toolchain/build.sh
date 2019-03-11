#!/bin/bash

set -e

docker build -t flant/werf-stapel-toolchain:0.2.0 stapel/toolchain

docker push flant/werf-stapel-toolchain:0.2.0
