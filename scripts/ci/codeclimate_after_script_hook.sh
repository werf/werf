#!/bin/bash -e

coverage_files=$WERF_TEST_COVERAGE_DIR/*.out
for file in ${coverage_files[@]}
do
  file_name=$(basename $file)
  ./cc-test-reporter format-coverage \
    -t=gocov \
    -o="coverage/$file_name.codeclimate.json" \
    "$file"
done

./cc-test-reporter sum-coverage \
  -p=$(ls -1q coverage/*.codeclimate.json | wc -l) \
  coverage/*.codeclimate.json

./cc-test-reporter upload-coverage
