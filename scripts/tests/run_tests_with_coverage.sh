#!/bin/bash -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

export WERF_TEST_COVERAGE_DIR=${WERF_TEST_COVERAGE_DIR:-$script_dir/tests_coverage}
rm -rf $WERF_TEST_COVERAGE_DIR
mkdir $WERF_TEST_COVERAGE_DIR

$script_dir/run_unit_tests_with_coverage.sh
$script_dir/run_integration_tests_with_coverage.sh all "${1:-0}"
