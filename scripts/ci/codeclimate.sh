#!/bin/bash -e

curl -L https://codeclimate.com/downloads/test-reporter/test-reporter-latest-linux-amd64 > ./cc-test-reporter
chmod +x ./cc-test-reporter

coverage_files=$(find $WERF_TEST_COVERAGE_DIR -name '*.out')
for file in ${coverage_files[@]}
do
  file_name=$(echo $file | tr / _)
  ./cc-test-reporter format-coverage \
    -t=gocov \
    -o="coverage/$file_name.codeclimate.json" \
    -p=github.com/flant/werf/ \
    "$file"
done

./cc-test-reporter sum-coverage \
  -p=$(ls -1q coverage/*.codeclimate.json | wc -l) \
  coverage/*.codeclimate.json

./cc-test-reporter upload-coverage
