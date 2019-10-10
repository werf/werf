#!/bin/bash -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
project_dir=$script_dir/../..

if [[ ! -z "$1" ]]; then
  bats_jobs_option="--jobs $1"
fi

if [[ ! -z "$WERF_TEST_K8S_INSECURE_DOCKER_REGISTRY" ]]; then
  export WERF_INSECURE_REGISTRY=true
fi

export WERF_TEST_K8S_DOCKER_REGISTRY=${WERF_TEST_K8S_DOCKER_REGISTRY:-$WERF_TEST_K8S_INSECURE_DOCKER_REGISTRY}
export KUBECONFIG=$(mktemp -d)/config
env_names=$(compgen -A variable | grep "^WERF_TEST_K8S_BASE64_KUBECONFIG_")
for env_name in ${env_names[@]}
do
  export WERF_TEST_K8S_VERSION=$(echo "$env_name" | grep -oE "[^_]+$")
  printenv $env_name | base64 -d > $KUBECONFIG
  bats -r $project_dir/tests -f '^\[k8s\]' $bats_jobs_option
done
