#!/bin/bash

set -e

scripts/update_pylogger_so.sh

go install github.com/flant/werf/cmd/werf
