#!/bin/bash -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

$script_dir/run_unit_tests.sh
$script_dir/run_integration_tests.sh all "${1:-0}"
