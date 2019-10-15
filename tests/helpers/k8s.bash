#!/bin/bash

test_skip_if_k8s_disabled() {
  if [[ ! -z "$WERF_TEST_K8S_DISABLED" ]]; then
    skip "k8s test was disabled by \$WERF_TEST_K8S_DISABLED"
  fi
}

test_requires_k8s_docker_registry() {
  if [[ -z "$WERF_TEST_K8S_DOCKER_REGISTRY" ]] || [[ -z "$WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME" ]] || [[ -z "$WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD" ]]; then
    skip "\$WERF_TEST_K8S_DOCKER_REGISTRY, \$WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME and \$WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD are required"
  fi
}

test_k8s_init_project_name() {
  export WERF_TEST_K8S_PROJECT_NAME=${1:-project}-$(generate_random_string)
}
