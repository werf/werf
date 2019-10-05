#!/bin/bash -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
project_dir=$script_dir/../..
cd $project_dir

tests_bin_path=$project_dir/bin/tests
mkdir -p $tests_bin_path

GO111MODULE=on CGO_ENABLED=0 go test --tags integration -coverpkg=./... -c cmd/werf/main.go cmd/werf/main_test.go -o $tests_bin_path/werf.test

cat <<'EOF' > $tests_bin_path/werf
#!/bin/bash

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
EOF
# Project dir is embedded into script
cat <<EOF >> $tests_bin_path/werf
project_dir=$project_dir
EOF
cat <<'EOF' >> $tests_bin_path/werf

if [ -z "$WERF_TEST_COVERAGE_DIR" ]
then
  coverage_dir=$project_dir/tests-coverage
else
  coverage_dir=$WERF_TEST_COVERAGE_DIR
fi

coverage_file_name="$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 10 | head -n 1)-$(date +%s).out"
coverage_file_path="$coverage_dir/$coverage_file_name"

mkdir -p $coverage_dir
exec $script_dir/werf.test -test.coverprofile=$coverage_file_path "$@"
EOF

chmod +x $tests_bin_path/werf
export PATH="$tests_bin_path:$PATH"
export WERF_INTEGRATION_TEST_COVERAGE_DIR=$coverage_dir

$script_dir/run_integration_tests.sh "${1:-all}" "${2:-0}"
