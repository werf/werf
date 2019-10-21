#!/bin/bash -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
project_dir=$script_dir/../..

go build github.com/google/go-containerregistry/cmd/crane
mv crane $project_dir/bin/tests
