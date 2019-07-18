#!/bin/bash

set -e

SOURCE=`dirname ${BASH_SOURCE[0]}`
CLI_PARTIALS_DIR=$SOURCE/_includes/cli
README=$SOURCE/../README.md
README_PARTIALS_DIR=$SOURCE/_includes/readme

$SOURCE/../go-build.sh
rm -rf $CLI_PARTIALS_DIR
rm -rf $README_PARTIALS_DIR
mkdir -p $CLI_PARTIALS_DIR
mkdir -p $README_PARTIALS_DIR

werf docs --dir $CLI_PARTIALS_DIR --log-terminal-width=100
werf docs --split-readme --readme $README --dir $README_PARTIALS_DIR
