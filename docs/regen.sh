#!/bin/bash

set -e

SOURCE=`dirname ${BASH_SOURCE[0]}`
PARTIALS_DIR=$SOURCE/_includes/cli

$SOURCE/../go-build.sh
rm -rf $PARTIALS_DIR
mkdir $PARTIALS_DIR
werf docs --dir $PARTIALS_DIR --log-terminal-width=100

cp -f $SOURCE/../README.md $SOURCE/_includes/README.md
