#!/bin/bash

docker build -t dappdeps/toolchain:0.1.0 .
docker login -u $DOCKER_HUB_LOGIN -p $DOCKER_HUB_PASSWORD
docker push dappdeps/toolchain:0.1.0
