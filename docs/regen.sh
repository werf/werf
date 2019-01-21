#!/bin/bash

SOURCE=`dirname ${BASH_SOURCE[0]}`
PARTIALS_DIR=$SOURCE/_includes/cli

$SOURCE/../go-build.sh
rm -rf $PARTIALS_DIR
mkdir $PARTIALS_DIR
$GOPATH/bin/werf docs --dir $PARTIALS_DIR
