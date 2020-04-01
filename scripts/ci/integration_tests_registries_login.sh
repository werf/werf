#!/bin/bash -ex

#env_names=$(compgen -A variable | grep "^WERF_TEST_DOCKER_REGISTRY_IMPLEMENTATION_" || true)
#for env_name in ${env_names[@]}
#do
#  name=${env_name##*_}
  if [[ -z "$1" ]]; then
    echo "script requires argument <implementation name>" >&2
    exit 1
  fi

  name=${1^^}
  registry=$(printenv WERF_TEST_"$name"_REGISTRY || true)
  username=$(printenv WERF_TEST_"$name"_USERNAME || true)
  password=$(printenv WERF_TEST_"$name"_PASSWORD || true)
  base64config=$(printenv WERF_TEST_"$name"_BASE64_CONFIG || true)

  if [[ "$name" == "DOCKERHUB" ]]; then
    echo "$password" | docker login -u "$username" --password-stdin
  elif [[ -n "$username" ]] && [[ -n "$password" ]]; then
    echo "$password" | docker login -u "$username" --password-stdin "$registry"
  elif [[ -n "$base64config" ]]; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
      echo "$base64config" | base64 -D | docker login -u _json_key --password-stdin "$registry"
    else
      echo "$base64config" | base64 -d | docker login -u _json_key --password-stdin "$registry"
    fi
  else
    echo "script requires environment variables with credentials" >&2
    exit 1
  fi
#done
