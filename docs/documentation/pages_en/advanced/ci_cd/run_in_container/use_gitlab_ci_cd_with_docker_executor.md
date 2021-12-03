---
title: Use Gitlab CI/CD with docker executor
permalink: advanced/ci_cd/run_in_container/use_gitlab_ci_cd_with_docker_executor.html
---

werf currently supports building of images _with docker server_ or _without docker server_ (in experimental mode).

> NOTICE: This page contains instructions, which are only applicable for experimental mode _without docker server_. Only dockerfile-images builder is available for this mode for now. Stapel-images builder will be available soon.

## 1. Setup kubernetes gitlab runner

Basic runner configuration:

```toml
[[runners]]
  name = "docker-runner-for-werf"
  executor = "docker"
  ...
  [runners.docker]
    privileged = true
    ...
```

## 2. Setup access for kubernetes cluster

Setup `WERF_KUBECONFIG_BASE64` secret environment variable in gitlab project with the content of `~/.kube/config` in base64 encoding. werf will automatically use this configuration to connect to the kubernetes cluster.

## 3. Configure project gitlab-ci.yml

Basic build and deploy job for a project:

```yaml
stages:
  - build-and-deploy

Build and deploy application:
  stage: build-and-deploy
  image: ghcr.io/werf/werf
  script:
    - source $(werf ci-env gitlab --as-file)
    - werf converge
  tags: ["docker-runner-for-werf"]
```
