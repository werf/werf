#!/bin/bash

set -e

source scripts/stapel/version.sh

docker pull flant/werf-stapel-base:$PREVIOUS_STAPEL_VERSION
docker pull flant/werf-stapel:$PREVIOUS_STAPEL_VERSION
