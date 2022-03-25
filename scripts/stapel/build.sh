#!/bin/bash

set -e

docker build -t registry-write.werf.io/werf/stapel-base:dev --target base --file stapel/Dockerfile .
docker build -t registry-write.werf.io/werf/stapel:dev --target final --file stapel/Dockerfile .
