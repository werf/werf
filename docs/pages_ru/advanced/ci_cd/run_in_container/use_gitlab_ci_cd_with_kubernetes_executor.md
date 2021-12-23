---
title: С помощью GitLab CI/CD с Kubernetes executor
permalink: advanced/ci_cd/run_in_container/use_gitlab_ci_cd_with_kubernetes_executor.html
---

> ПРИМЕЧАНИЕ: в настоящее время werf поддерживает сборку образов с _использованием Docker-сервера_ или _без его использования_ (в экспериментальном режиме). Эта страница содержит инструкции, которые подходят только для экспериментального режима _без Docker-сервера_. На данный момент для этого типа сборки доступен только сборщик образов на основе Dockerfile'ов. Сборщик Stapel будет доступен через некоторое время.

## 1. Подготовьте кластер Kubernetes

Выберите один из [доступных режимов работы]({{ "advanced/ci_cd/run_in_container/how_it_works.html#режимы-работы" | true_relative_url }}) и перейдите к нему.

### Ядро Linux с поддержкой OverlayFS в режиме rootless

Дополнительные действия не требуются.

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование привилегированного контейнера

Дополнительные действия не требуются.

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование непривилегированного контейнера

Для подключения устройства `/dev/fuse` в контейнерах под управлением werf необходим плагин [fuse device](https://github.com/kuberenetes-learning-group/fuse-device-plugin):

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

Примените приведенный выше манифест плагина в пространстве имен `kube-system`:

```
kubectl -n kube-system apply -f werf-fuse-device-plugin-ds.yaml
```

Также давайте создадим политику LimitRange с тем, чтобы Pod'ы, создаваемые в некотором пространстве имен, имели доступ к `/dev/fuse`:

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

Создайте пространство имен `gitlab-ci` и примените манифест LimitRange в этом пространстве имен (позже мы настроим раннер GitLab на использование этого пространства имен при создании Pod'ов для выполнения CI-заданий):

```
kubectl create namespace gitlab-ci
kubectl apply -f enable-fuse-pod-limit-range.yaml
```

## 2. Настройте GitLab-раннер в Kubernetes

Выберите один из [доступных режимов работы]({{ "advanced/ci_cd/run_in_container/how_it_works.html#режимы-работы" | true_relative_url }}) и перейдите к нему.

### Ядро Linux с поддержкой OverlayFS в режиме rootless

Базовая конфигурация раннера (`/etc/gitlab-runner/config.toml`):

```toml
[[runners]]
  name = "kubernetes-runner-for-werf"
  executor = "kubernetes"
  ...
  [runners.kubernetes]
    ...
    pod_annotations = ["container.apparmor.security.beta.kubernetes.io/werf-converge=unconfined"]
```

О дополнительных опциях можно узнать из [документации к Kubernetes executor](https://docs.gitlab.com/runner/executors/kubernetes.html).

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование привилегированного контейнера

Базовая конфигурация раннера (`/etc/gitlab-runner/config.toml`):

```toml
[[runners]]
  name = "kubernetes-runner-for-werf"
  executor = "kubernetes"
  ...
  [runners.kubernetes]
    ...
    privileged = true
```

О дополнительных опциях можно узнать из документации к [Kubernetes executor](https://docs.gitlab.com/runner/executors/kubernetes.html).

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование непривилегированного контейнера

Базовая конфигурация раннера (`/etc/gitlab-runner/config.toml`):

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

Обратите внимание, что было указано пространство имен `gitlab-ci`. Это пространство имен должно быть таким же, что и на шаге 1, для автоматической генерации лимитов Pod'ов.

О дополнительных опциях можно узнать из документации к [Kubernetes executor](https://docs.gitlab.com/runner/executors/kubernetes.html).

## 3. Настройте доступ проекта к целевому кластеру Kubernetes

Существует 2 способа доступа к целевому кластеру Kubernetes, в который разворачивается приложение:

1. Используя Service Account для Kubenetes executor. Этот метод подходит только в том случае, если Kubernetes executor работает в целевом кластере Kubernetes.
2. Используя kubeconfig с соответствующим настройками.

### Service account

Вот пример конфигурации Service Account'а с именем `gitlab-kubernetes-runner-deploy`:

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

Скорректируйте конфигурацию GitLab-раннера (`/etc/gitlab-runner/config.toml`), чтобы использовать этот Service Account:

```toml
[[runners]]
  name = "kubernetes-runner-for-werf"
  ...
  [runners.kubernetes]
    service_account = "gitlab-kubernetes-runner-deploy"
    ...
```

### Kubeconfig

Присвойте переменной окружения `WERF_KUBECONFIG_BASE64` в GitLab-проекте содержимое файла `~/.kube/config`, закодированное в base64. werf будет автоматически использовать эту конфигурацию для подключения к целевому кластеру Kubernetes.

Этот метод подходит для случаев, когда Kubernetes executor и целевой кластер Kubernetes — два разных кластера.

## 4. Настройте файл gitlab-ci.yml проекта

Ниже приведено описание базового задания по сборке и развертыванию проекта:

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

## Устранение проблем

Если у вас возникли какие-либо сложности, пожалуйста, обратитесь к разделу [Устранение проблем]({{ "advanced/ci_cd/run_in_container/how_it_works.html#устранение-проблем" | true_relative_url }})
