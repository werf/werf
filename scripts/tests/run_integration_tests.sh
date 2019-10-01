#!/bin/bash -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
project_dir=$script_dir/../..

run_bats() {
  if [ "$1" == "all" ]
  then
    test_glob="$project_dir/tests/**"
  else
    test_glob="$1"
  fi

  bats_options="-r $test_glob"

  if [ "$2" != "0" ]
  then
    bats_options="$bats_options --jobs $2"
  fi

  bats $bats_options
}

run_bats ${1:-all} ${2:-0}
