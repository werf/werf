---
title: Use Gitlab CI/CD with kubernetes executor
permalink: advanced/ci_cd/run_in_container/use_gitlab_ci_cd_with_kubernetes_executor.html
---

> NOTICE: werf currently supports building of images _with docker server_ or _without docker server_ (in experimental mode). This page contains instructions, which are only applicable for experimental mode _without docker server_. Only dockerfile-images builder is available for this mode for now. Stapel-images builder will be available soon.

## 1. Prepare kubernetes cluster

Select and proceed with one of the [available modes of operation]({{ "advanced/ci_cd/run_in_container/how_it_works.html#modes-of-operation" | true_relative_url }}).

### Linux kernel with rootless overlayfs

No actions needed.

### Linux kernel without rootless overlayfs and privileged container

No actions needed.

### Linux kernel without rootless overlayfs and non-privileged container

[Fuse device plugin](https://github.com/kuberenetes-learning-group/fuse-device-plugin) needed to enable `/dev/fuse` device in containers running werf:

```
# werf-fuse-device-plugin-ds.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: werf-fuse-device-plugin
spec:
  selector:
    matchLabels:
      name: werf-fuse-device-plugin
  template:
    metadata:
      labels:
        name: werf-fuse-device-plugin
    spec:
      hostNetwork: true
      containers:
      - image: soolaugust/fuse-device-plugin:v1.0
        name: fuse-device-plugin-ctr
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop: ["ALL"]
        volumeMounts:
          - name: device-plugin
            mountPath: /var/lib/kubelet/device-plugins
      volumes:
        - name: device-plugin
          hostPath:
            path: /var/lib/kubelet/device-plugins
```

Apply provided device plugin in the `kube-system` namespace:

```
kubectl -n kube-system apply -f werf-fuse-device-plugin-ds.yaml
```

Also let's define a limit range so that Pods created in some namespace will have an access to the `/dev/fuse`:

```
# enable-fuse-limit-range.yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: enable-fuse-limit-range
spec:
  limits:
  - type: "Container"
    default:
      github.com/fuse: 1
```

Create namespace `gitlab-ci` and limit range in this namespace (later we will setup gitlab-runner to use this namespace when creating Pods to run ci-jobs):

```
kubectl create namespace gitlab-ci
kubectl apply -f enable-fuse-pod-limit-range.yaml
```

## 2. Setup kubernetes gitlab runner

Select and proceed with one of the [available modes of operation]({{ "advanced/ci_cd/run_in_container/how_it_works.html#modes-of-operation" | true_relative_url }}).

### Linux kernel with rootless overlayfs

Basic runner configuration (`/etc/gitlab-runner/config.toml`):

```toml
[[runners]]
  name = "kubernetes-runner-for-werf"
  executor = "kubernetes"
  builds_dir = "/builds"
  environment = ["HOME=/builds"]
  ...
  [runners.kubernetes]
    ...
    pod_annotations = ["container.apparmor.security.beta.kubernetes.io/werf-converge=unconfined"]
    [runners.kubernetes.pod_security_context]
      fs_group = 1000
      run_as_group = 1000
      run_as_non_root = true
      run_as_user = 1000
```

For more options consult [gitlab kubernetes executor documentation page](https://docs.gitlab.com/runner/executors/kubernetes.html).

### Linux kernel without rootless overlayfs and privileged container

Basic runner configuration (`/etc/gitlab-runner/config.toml`):

```toml
[[runners]]
  name = "kubernetes-runner-for-werf"
  executor = "kubernetes"
  builds_dir = "/builds"
  environment = ["HOME=/builds"]
  ...
  [runners.kubernetes]
    ...
    privileged = true
    [runners.kubernetes.pod_security_context]
      fs_group = 1000
      run_as_group = 1000
      run_as_non_root = true
      run_as_user = 1000
```

For more options consult [gitlab kubernetes executor documentation page](https://docs.gitlab.com/runner/executors/kubernetes.html).

### Linux kernel without rootless overlayfs and non-privileged container

Basic runner configuration (`/etc/gitlab-runner/config.toml`):

```toml
[[runners]]
  name = "kubernetes-runner-for-werf"
  executor = "kubernetes"
  builds_dir = "/builds"
  environment = ["HOME=/builds"]
  ...
  [runners.kubernetes]
    ...
    namespace = "gitlab-ci"
    pod_annotations = ["container.apparmor.security.beta.kubernetes.io/werf-converge=unconfined"]
    [runners.kubernetes.pod_security_context]
      fs_group = 1000
      run_as_group = 1000
      run_as_non_root = true
      run_as_user = 1000
```

Note that `gitlab-ci` namespace has been specified. The same namespace should be used as in the step 1 to create auto pod limit settings.

For more options consult [gitlab kubernetes executor documentation page](https://docs.gitlab.com/runner/executors/kubernetes.html).

## 3. Configure project access to target kubernetes cluster

There are 2 ways to access kubernetes cluster which is the target cluster to deploy your application into:

1. Using special service account from kubernetes executor. This method only suitable when kubernetes executor runs in the target kubernetes cluster.
2. Using configured kube config.

### Service account

Example service account configuration named `gitlab-kubernetes-runner-deploy`:

```yaml
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

```toml
[[runners]]
  name = "kubernetes-runner-for-werf"
  ...
  [runners.kubernetes]
    service_account = "gitlab-kubernetes-runner-deploy"
    ...
```

### Kube config

Setup `WERF_KUBECONFIG_BASE64` secret environment variable in gitlab project with the content of `~/.kube/config` in base64 encoding. werf will automatically use this configuration to connect to the kubernetes cluster which is the target for application.

This method is suitable when kubernetes executor and target kubernetes cluster are 2 different clusters.

## 4. Configure project gitlab-ci.yml

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
  tags: ["kubernetes-runner-for-werf"]
```

## Troubleshooting

If you have any problems please refer to the [troubleshooting section]({{ "advanced/ci_cd/run_in_container/how_it_works.html#troubleshooting" | true_relative_url }})
