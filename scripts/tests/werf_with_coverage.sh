#!/bin/bash -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
project_dir=$script_dir/../..
project_bin_tests_dir=$project_dir/bin/tests

generate_binary() {
    cd $project_dir
    go test -tags "dfrunmount dfssh integration_coverage" -coverpkg=./... -c cmd/werf/main.go cmd/werf/main_test.go -o $project_bin_tests_dir/werf.test

    cat <<'EOF' > $project_bin_tests_dir/werf
#!/bin/bash

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
EOF

    # Project dir is embedded into script
    cat <<EOF >> $project_bin_tests_dir/werf
project_dir=$project_dir
EOF

    cat <<'EOF' >> $project_bin_tests_dir/werf

coverage_dir=${WERF_TEST_COVERAGE_DIR:-$project_dir/tests_coverage}
mkdir -p $coverage_dir

coverage_file_name="$(date +%s.%N | sha256sum | cut -c 1-10)-$(date +%s).out"
coverage_file_path="$coverage_dir/$coverage_file_name"

exec $script_dir/werf.test -test.coverprofile=$coverage_file_path "$@"
EOF
}

mkdir -p $project_bin_tests_dir
generate_binary
chmod +x $project_bin_tests_dir/werf
