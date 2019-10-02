#!/bin/bash

set -e

source scripts/stapel/version.sh

docker build -t flant/werf-stapel-base:$CURRENT_STAPEL_VERSION --target base --cache-from flant/werf-stapel-base:$PREVIOUS_STAPEL_VERSION --file stapel/Dockerfile .

docker build -t flant/werf-stapel:$CURRENT_STAPEL_VERSION --target final --cache-from flant/werf-stapel-base:$CURRENT_STAPEL_VERSION --cache-from flant/werf-stapel:$PREVIOUS_STAPEL_VERSION --file stapel/Dockerfile .

sed -e "s/export PREVIOUS_STAPEL_VERSION.*/export PREVIOUS_STAPEL_VERSION=$CURRENT_STAPEL_VERSION/" -i scripts/stapel/version.sh
