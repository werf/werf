#!/bin/bash -ex

env_names=$(compgen -A variable | grep "^WERF_TEST_DOCKER_REGISTRY_IMPLEMENTATION_")
for env_name in ${env_names[@]}
do
  name=$(printenv "$env_name")
  server=$(printenv WERF_TEST_"$name"_SERVER)
  username=$(printenv WERF_TEST_"$name"_USERNAME)
  password=$(printenv WERF_TEST_"$name"_PASSWORD)
  base64config=$(printenv WERF_TEST_"$name"_BASE64_CONFIG)

  if [[ -n "$username" ]] && [[ -n "$password" ]]; then
    echo "$password" | docker login -u "$username" --password-stdin "$server"
  elif [[ -n "$base64config" ]]; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
      echo "$base64config" | base64 -D | docker login -u _json_key --password-stdin "$server"
    else
      echo "$base64config" | base64 -d | docker login -u _json_key --password-stdin "$server"
    fi
  fi
done
