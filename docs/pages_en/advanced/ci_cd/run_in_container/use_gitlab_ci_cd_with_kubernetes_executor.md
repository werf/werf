---
title: Use GitLab CI/CD with Kubernetes executor
permalink: advanced/ci_cd/run_in_container/use_gitlab_ci_cd_with_kubernetes_executor.html
---

> NOTICE: werf currently supports building images _with the Docker server_ or _without the Docker server_ (in experimental mode). This page contains information applicable only to the experimental mode _without the Docker server_. For now, only the Dockerfile image builder is available for this mode. The Stapel image builder will be available soon.

## 1. Prepare the Kubernetes cluster

Make sure you meet all [system requirements]({{ "advanced/buildah_mode.html#system-requirements" | true_relative_url }}) and select one of the [available operating modes]({{ "advanced/ci_cd/run_in_container/how_it_works.html#modes-of-operation" | true_relative_url }}).

### Linux kernel with rootless OverlayFS

No additional actions are required.

### Linux kernel without rootless OverlayFS and privileged container

No additional actions are required.

### Linux kernel without rootless OverlayFS and non-privileged container

The [fuse device plugin](https://github.com/kuberenetes-learning-group/fuse-device-plugin) is one of the ways to enable the `/dev/fuse` device in containers with werf:

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

Apply the above plugin manifest to the `kube-system` namespace:

```
kubectl -n kube-system apply -f werf-fuse-device-plugin-ds.yaml
```

Let's also create a LimitRange policy so that Pods created in some namespace have access to `/dev/fuse`:

```
# enable-fuse-limit-range.yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: enable-fuse-limit-range
  namespace: gitlab-ci
spec:
  limits:
  - type: "Container"
    default:
      github.com/fuse: 1
```

Create a `gitlab-ci` namespace and apply the LimitRange manifest in that namespace (we will later configure the GitLab runner to use that namespace when creating Pods for running CI jobs):

```
kubectl create namespace gitlab-ci
kubectl apply -f enable-fuse-pod-limit-range.yaml
```

## 2. Setup the GitLab runner in Kubernetes

Select and proceed with one of the [available modes of operation]({{ "advanced/ci_cd/run_in_container/how_it_works.html#modes-of-operation" | true_relative_url }}).

### Linux kernel with rootless OverlayFS

Basic runner configuration (`/etc/gitlab-runner/config.toml`):

```toml
[[runners]]
  name = "kubernetes-runner-for-werf"
  executor = "kubernetes"
  ...
  [runners.kubernetes]
    ...
    namespace = "gitlab-ci"
    pod_annotations = ["container.apparmor.security.beta.kubernetes.io/werf-converge=unconfined"]
```

For more options, consult the [Kubernetes executor for GitLab runner documentation](https://docs.gitlab.com/runner/executors/kubernetes.html).

### Linux kernel without rootless OverlayFS and privileged container

Basic runner configuration (`/etc/gitlab-runner/config.toml`):

```toml
[[runners]]
  name = "kubernetes-runner-for-werf"
  executor = "kubernetes"
  ...
  [runners.kubernetes]
    ...
    namespace = "gitlab-ci"
    privileged = true
```

For more options, consult the [Kubernetes executor for GitLab runner documentation](https://docs.gitlab.com/runner/executors/kubernetes.html).

### Linux kernel without rootless OverlayFS and non-privileged container

Basic runner configuration (`/etc/gitlab-runner/config.toml`):

```toml
[[runners]]
  name = "kubernetes-runner-for-werf"
  executor = "kubernetes"
  ...
  [runners.kubernetes]
    ...
    namespace = "gitlab-ci"
    pod_annotations = ["container.apparmor.security.beta.kubernetes.io/werf-converge=unconfined"]
```

For more options, consult the [Kubernetes executor for GitLab runner documentation](https://docs.gitlab.com/runner/executors/kubernetes.html).

## 3. Configure project access to the target Kubernetes cluster

There are 2 ways to access the target Kubernetes cluster to which the application is deployed:

1. Using Service Account for the Kubenetes executor. This method is only suitable if the Kubernetes executor is running in the target Kubernetes cluster.
2. Using kubeconfig with the appropriate settings.

### Service Account

Here is an example configuration of a Service Account named `gitlab-kubernetes-runner-deploy`:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gitlab-kubernetes-runner-deploy
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: gitlab-kubernetes-runner-deploy
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: gitlab-kubernetes-runner-deploy
    namespace: gitlab-ci
```

Adjust the GitLab runner configuration (`/etc/gitlab-runner/config.toml`) to use this Service Account:

```toml
[[runners]]
  name = "kubernetes-runner-for-werf"
  ...
  [runners.kubernetes]
    service_account = "gitlab-kubernetes-runner-deploy"
    ...
```

### Kubeconfig

Assign the base64-encoded contents of `~/.kube/config` to the `WERF_KUBECONFIG_BASE64` environment variable in the GitLab project. werf will automatically use this configuration to connect to the target Kubernetes cluster.

This method is appropriate when the Kubernetes executor and the target Kubernetes cluster are two different clusters.

## 4. Configure gitlab-ci.yml of the project

Below is a description of the basic build and deploy job for a project:

```yaml
stages:
  - build-and-deploy

Build and deploy application:
  stage: build-and-deploy
  image: registry.werf.io/werf/werf
  script:
    - source $(werf ci-env gitlab --as-file)
    - werf converge
  tags: ["kubernetes-runner-for-werf"]
```

## Troubleshooting

In case of problems, refer to the [Troubleshooting section]({{ "advanced/ci_cd/run_in_container/how_it_works.html#troubleshooting" | true_relative_url }})
