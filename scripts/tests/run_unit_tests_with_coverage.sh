#!/bin/bash -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
project_dir=$script_dir/../..

if [ -z "$WERF_TEST_COVERAGE_DIR" ]
then
  unit_tests_dir=$script_dir/unit-tests
  coverage_dir=$unit_tests_dir/tests-coverage
  rm -rf $unit_tests_dir
else
  coverage_dir=$WERF_TEST_COVERAGE_DIR
fi

mkdir -p $coverage_dir

coverage_file_name="$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 10 | head -n 1)-$(date +%s).out"
coverage_file_path="$coverage_dir/$coverage_file_name"

cd $project_dir
GO111MODULE=on CGO_ENABLED=0 go test -coverpkg=./... -coverprofile=$coverage_file_path ./...
