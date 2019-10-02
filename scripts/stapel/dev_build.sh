#!/bin/bash

set -e

#docker build -t flant/werf-stapel-base:dev --target base --file stapel/Dockerfile .
#docker build -t flant/werf-stapel:dev --target final --file stapel/Dockerfile .
docker build -t flant/werf-stapel:dev --file stapel/Dockerfile .
