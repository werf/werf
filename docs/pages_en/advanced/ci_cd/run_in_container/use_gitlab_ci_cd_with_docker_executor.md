---
title: Use GitLab CI/CD with Docker executor
permalink: advanced/ci_cd/run_in_container/use_gitlab_ci_cd_with_docker_executor.html
---

> NOTICE: werf currently supports building images _with the Docker server_ or _without the Docker server_ (in experimental mode). This page contains information applicable only to the experimental mode _without the Docker server_. For now, only the Dockerfile image builder is available for this mode. The Stapel image builder will be available soon.

## 1. Configure GitLab rinner for Kubernetes

Make sure you meet all [system requirements]({{ "advanced/buildah_mode.html#system-requirements" | true_relative_url }}) and select one of the [available operating modes]({{ "advanced/ci_cd/run_in_container/how_it_works.html#modes-of-operation" | true_relative_url }}) (depending on the capabilities of your GitLab runner) and navigate to it.

### Linux kernel with rootless OverlayFS

Basic runner configuration:

```toml
[[runners]]
  name = "docker-runner-for-werf"
  executor = "docker"
  ...
  [runners.docker]
    security_opt = ["seccomp:unconfined", "apparmor:unconfined"]
    ...
```

### Linux kernel without rootless OverlayFS and privileged container

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

### Linux kernel without rootless OverlayFS and non-privileged container

Basic runner configuration:

```toml
[[runners]]
  name = "docker-runner-for-werf"
  executor = "docker"
  ...
  [runners.docker]
    security_opt = ["seccomp:unconfined", "apparmor:unconfined"]
    devices = ["/dev/fuse"]
    ...
```

## 2. Configure access to the Kubernetes cluster

Assign the `WERF_KUBECONFIG_BASE64` environment variable in the GitLab project a base64-encoded value from `~/.kube/config`. werf will automatically use this configuration to connect to the Kubernetes cluster.

## 3. Configure gitlab-ci.yml of the project

Below is a basic build and deploy job for a project:

```yaml
stages:
  - build-and-deploy

Build and deploy application:
  stage: build-and-deploy
  image: registry.werf.io/werf/werf
  script:
    - source $(werf ci-env gitlab --as-file)
    - werf converge
  tags: ["docker-runner-for-werf"]
```

## Troubleshooting

In case of problems, refer to the [Troubleshooting section]({{ "advanced/ci_cd/run_in_container/how_it_works.html#troubleshooting" | true_relative_url }})
