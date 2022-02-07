---
title: С помощью GitLab CI/CD с Docker executor
permalink: advanced/ci_cd/run_in_container/use_gitlab_ci_cd_with_docker_executor.html
---

> ПРИМЕЧАНИЕ: в настоящее время werf поддерживает сборку образов с _использованием Docker-сервера_ или _без его использования_ (в экспериментальном режиме). Эта страница содержит инструкции, которые подходят только для экспериментального режима _без Docker-сервера_. На данный момент для этого типа сборки доступен только сборщик образов на основе Dockerfile'ов. Сборщик Stapel будет доступен через некоторое время.

## 1. Настройте GitLab-раннер в Kubernetes

Убедитесь, что удовлетворены [системные требования]({{ "advanced/buildah_mode.html#системные-требования" | true_relative_url }}) и выберите один из [доступных режимов работы]({{ "advanced/ci_cd/run_in_container/how_it_works.html#режимы-работы" | true_relative_url }}) (в зависимости от возможностей вашего GitLab-раннера) и перейдите к нему.

### Ядро Linux с поддержкой OverlayFS в режиме rootless

Базовая конфигурация раннера:

```toml
[[runners]]
  name = "docker-runner-for-werf"
  executor = "docker"
  ...
  [runners.docker]
    security_opt = ["seccomp:unconfined", "apparmor:unconfined"]
    ...
```

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование привилегированного контейнера

Базовая конфигурация раннера:

```toml
[[runners]]
  name = "docker-runner-for-werf"
  executor = "docker"
  ...
  [runners.docker]
    privileged = true
    ...
```

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование непривилегированного контейнера

Базовая конфигурация раннера:

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

## 2. Настройте доступ к кластеру Kubernetes

Присвойте переменной окружения `WERF_KUBECONFIG_BASE64` в GitLab-проекте значение из файла `~/.kube/config`, закодированное в base64. werf будет автоматически использовать данную конфигурацию для подключения к кластеру Kubernetes.

## 3. Настройте файл gitlab-ci.yml проекта

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
  tags: ["docker-runner-for-werf"]
```

## Устранение проблем

Если у вас возникли какие-либо сложности, пожалуйста, обратитесь к разделу [Устранение проблем]({{ "advanced/ci_cd/run_in_container/how_it_works.html#устранение-проблем" | true_relative_url }})
