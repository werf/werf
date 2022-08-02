#!/bin/bash -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
project_dir=$script_dir/../..

find_dir=${1:-.}
tests_binaries_output_dirname=${2:-$project_dir/precompiled_test_binaries}

if [[ "$OSTYPE" == "darwin"* ]]; then
  if ! [[ -x "$(command -v gfind)" ]]; then
    brew install findutils
  fi

  package_paths=$(gfind -L "$find_dir" -type f -name '*_test.go' -printf '%h\n' | sort -u)
else
  package_paths=$(find  -L "$find_dir" -type f -name '*_test.go' -printf '%h\n' | sort -u)
fi

unameOut="$(uname -s)"
case "${unameOut}" in
    CYGWIN*|MINGW*|MSYS*) ext=".test.exe";;
    *)                    ext=".test"
esac

for package_path in $package_paths; do
  if [[ $package_path == "./integration/suites/docs/docs/backend" ]] || [[ $package_path == "./integration/suites/docs/docs/site/_site/backend" ]] || [[ $package_path == "./integration/ci_suites/default/docs/_fixtures/cli/docs/site/backend" ]]; then
    continue
  fi

  test_binary_filename=$(basename -- "$package_path")$ext
	test_binary_path="$tests_binaries_output_dirname"/"$package_path"/"$test_binary_filename"
	./scripts/ci/go-test.sh "$package_path" -coverpkg=./... -c -o "$test_binary_path"

  if [[ ! -f $test_binary_path ]]; then # cmd/werf/main_test.go
     continue
  fi

#	if [[ -x "$(command -v upx)" ]]; then
#	  upx "$test_binary_path"
#  fi
done
