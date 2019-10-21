#!/bin/bash -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
project_dir=$script_dir/../..

go build github.com/onsi/ginkgo/ginkgo
mv ginkgo $project_dir/bin/tests
