#!/bin/bash

set -e

$(dirname $0)/scripts/update_pylogger_so.sh

go install github.com/flant/werf/cmd/werf
