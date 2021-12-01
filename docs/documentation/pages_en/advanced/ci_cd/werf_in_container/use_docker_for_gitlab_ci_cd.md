---
title: Run werf in Gitlab CI/CD with docker executor
permalink: advanced/ci_cd/werf_in_container/use_docker_for_gitlab_ci_cd.html
---

Werf currently supports building of images _with docker server_ or _without docker server_ (in experimental mode).

**NOTE** This page contains instructions, which are only applicable for experimental mode _without docker server_. Only dockerfile-images builder is available for this mode for now. Stapel-images builder will be available soon.

## 1. Setup kubernetes gitlab runner

Basic runner configuration:

```
[[runners]]
  name = "docker-runner-for-werf"
  url = GITLAB_URL
  token = RUNNER_TOKEN
  executor = "docker"
  [runners.docker]
    privileged = true
    ...
```

## 2. Setup access for kubernetes cluster

Setup `WERF_KUBECONFIG_BASE64` secret environment variable in gitlab project with the content of `~/.kube/config` in base64 encoding. Werf will automatically use this configuration to connect to the kubernetes cluster.

## 3. Configure project gitlab-ci.yml

Basic build and deploy job for a project:

```
stages:
  - build-and-deploy

Build and deploy application:
  stage: build-and-deploy
  image: ghcr.io/werf/werf
  script:
    - source $(werf ci-env gitlab --as-file)
    - werf converge
  tags: ["kubernetes-runner-for-werf"]
```
