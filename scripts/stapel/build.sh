#!/bin/bash

set -e

docker build -t ghcr.io/werf/stapel-base:dev --target base --file stapel/Dockerfile .
docker build -t ghcr.io/werf/stapel:dev --target final --file stapel/Dockerfile .
