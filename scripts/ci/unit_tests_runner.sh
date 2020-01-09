#!/bin/bash -e

test_binaries=$(find . -type f -name '*.test')
for test_binary in $test_binaries; do
  coverage_file_name="$(openssl rand -hex 6)-$(date +"%H_%M_%S")_coverage.out"
  $test_binary -test.v -test.coverprofile="$WERF_TEST_COVERAGE_DIR"/"$coverage_file_name"
done
