---
title: Run werf in Gitlab CI/CD with kubernetes executor
permalink: advanced/ci_cd/werf_in_container/use_kubernetes_for_gitlab_ci_cd.html
---

Werf currently supports building of images _with docker server_ or _without docker server_ (in experimental mode).

**NOTE** This page contains instructions, which are only applicable for experimental mode _without docker server_. Only dockerfile-images builder is available for this mode for now. Stapel-images builder will be available soon.

[comment]: <> (## 1. Prepare kubernetes cluster)

[comment]: <> (Create fuse device plugin daemonset, to access /dev/fuse on each kubernetes node. This is needed for werf to build images. Werf internally uses buildah with fuse-overlayfs overlayfs implementation.)

## 1. Setup kubernetes gitlab runner

Basic runner configuration (`/etc/gitlab-runner/config.toml`):

```
[[runners]]
  name = "kubernetes-runner-for-werf"
  url = GITLAB_URL
  token = RUNNER_TOKEN
  executor = "kubernetes"
  builds_dir = "/builds"
  environment = ["HOME=/builds"]
  [runners.kubernetes]
    privileged = true # FIXME
    [runners.kubernetes.pod_security_context]
      fs_group = 1000
      run_as_group = 1000
      run_as_non_root = true
      run_as_user = 1000
```

For more options consult [gitlab kubernetes executor documentation page](https://docs.gitlab.com/runner/executors/kubernetes.html).

## 2. Configure project access to target kubernetes cluster

There are 2 ways to access kubernetes cluster which is the target cluster to deploy your application into:

1. Using special service account from kubernetes executor. This method only suitable when kubernetes executor runs in the target kubernetes cluster.
2. Using configured kube config.

### Service account

Example service account configuration named `gitlab-kubernetes-runner-deploy`:

```
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gitlab-kubernetes-runner-deploy
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: werf
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: werf
    namespace: default
```

Adjust gitlab-runner configuration (`/etc/gitlab-runner/config.toml`) to use this service account:

```
[[runners]]
  name = "kubernetes-runner-for-werf"
  ...
  [runners.kubernetes]
    namespace = "gitlab-ci"
    service_account = "gitlab-kubernetes-runner-deploy"
    ...
```

### Kube config

Setup `WERF_KUBECONFIG_BASE64` secret environment variable in gitlab project with the content of `~/.kube/config` in base64 encoding. Werf will automatically use this configuration to connect to the kubernetes cluster which is the target for application.

This method is suitable when kubernetes executor and target kubernetes cluster are 2 different clusters.

## 2. Configure project gitlab-ci.yml

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
