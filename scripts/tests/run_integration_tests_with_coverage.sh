#!/bin/bash -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
project_dir=$script_dir/../..

integration_tests_dir=$script_dir/integration_tests
rm -rf $integration_tests_dir
mkdir $integration_tests_dir

cd $project_dir
GO111MODULE=on CGO_ENABLED=0 go test --tags integration -coverpkg=./... -c cmd/werf/main.go cmd/werf/main_test.go -o $integration_tests_dir/werf.test

cd $integration_tests_dir

if [ -z "$WERF_TEST_COVERAGE_DIR" ]
then
  coverage_dir=$integration_tests_dir/tests-coverage
else
  coverage_dir=$WERF_TEST_COVERAGE_DIR
fi

mkdir -p $coverage_dir

export WERF_INTEGRATION_TEST_COVERAGE_DIR=$coverage_dir

cat <<'EOF' > werf
#!/bin/bash

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

coverage_dir=$WERF_INTEGRATION_TEST_COVERAGE_DIR
coverage_file_name="$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 10 | head -n 1)-$(date +%s).out"
coverage_file_path="$coverage_dir/$coverage_file_name"

$script_dir/werf.test -test.coverprofile=$coverage_file_path "$@"
EOF

chmod +x werf
export PATH="`pwd`:$PATH"

$script_dir/run_integration_tests.sh "${1:-all}" "${2:-0}"
