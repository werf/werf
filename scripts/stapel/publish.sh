#!/bin/bash

set -e

source scripts/stapel/version.sh

#docker push flant/werf-stapel-base:$CURRENT_STAPEL_VERSION
docker push flant/werf-stapel:$CURRENT_STAPEL_VERSION
