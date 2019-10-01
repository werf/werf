#!/bin/bash -e

# ping stdout every 9 minutes or Travis kills build
# https://docs.travis-ci.com/user/common-build-problems/#Build-times-out-because-no-output-was-received
while sleep 9m; do echo "=====[ $SECONDS seconds still running ]====="; done
