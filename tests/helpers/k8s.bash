#!/bin/bash

test_skip_if_k8s_disabled() {
  if [[ ! -z "$WERF_TEST_K8S_DISABLED" ]]; then
    skip "k8s test was disabled by \$WERF_TEST_K8S_DISABLED"
  fi
}

test_requires_k8s_docker_registry() {
  if [ -z "$WERF_TEST_K8S_DOCKER_REGISTRY" ]; then
    skip "\$WERF_TEST_K8S_DOCKER_REGISTRY is required"
  fi
}

test_k8s_version() {
  if [[ ! -z "$WERF_TEST_K8S_VERSION" ]]; then
    echo " ($WERF_TEST_K8S_VERSION)"
  fi
}
