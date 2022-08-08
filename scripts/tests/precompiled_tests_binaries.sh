#!/bin/bash -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
project_dir=$script_dir/../..

find_dir=${1:-.}
tests_binaries_output_dirname=${2:-$project_dir/precompiled_test_binaries}

package_paths=$(find -L "$find_dir" -type f -name '*_test.go' -print0 | xargs -I{} -0 dirname '{}' | sort -u)

unameOut="$(uname -s)"
case "${unameOut}" in
    CYGWIN*|MINGW*|MSYS*) ext=".test.exe";;
    *)                    ext=".test"
esac

find -L "$find_dir" -type f -name '*_test.go' -print0 | xargs -I{} -0 dirname '{}' | sort -u | \
  while read src_pkg_rel_path; do
    test_bin_output_path="$tests_binaries_output_dirname/$src_pkg_rel_path/$(basename -- $src_pkg_rel_path)$ext"
    task -p test:go-test paths="'$src_pkg_rel_path'" -- -cover -coverpkg=./... -c -o "$test_bin_output_path"
  done
