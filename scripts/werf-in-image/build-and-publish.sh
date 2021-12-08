#!/bin/bash

REPO="ghcr.io/werf/werf"

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd $SCRIPT_DIR

werf export --repo "${REPO}" --tag "${REPO}:%image%"
