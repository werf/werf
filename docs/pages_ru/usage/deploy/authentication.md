---
title: Аутентификация и авторизация
permalink: usage/deploy/authentication.html
---

## Задание kubeconfig для доступа к Kubernetes

По умолчанию мы используем файл `~/.kube/config` для доступа к кластеру Kubernetes. Вы можете указать другой kubeconfig с помощью следующих параметров:

1. `--kube-config=<путь>` или `$WERF_KUBE_CONFIG=<путь>`: изменить путь к файлу kubeconfig.
1. `--kube-config-base64=<base64>` или `$WERF_KUBE_CONFIG_BASE64=<base64>`: передать файл kubeconfig, закодированный в base64, через командную строку или переменную окружения.

## Переопределение конфигурации kubeconfig

Вы можете переопределить конфигурацию kubeconfig с помощью следующих параметров:

1. `--kube-context=<контекст>` или `$WERF_KUBE_CONTEXT=<контекст>`: изменить контекст kubeconfig.
1. `--kube-token=<токен>` или `$WERF_KUBE_CONTEXT=<токен>`: задать токен для авторизации в Kubernetes.
1. `--kube-api-server=<url>` или `$WERF_KUBE_API_SERVER=<url>`: изменить URL Kubernetes API-сервера.
1. `--kube-tls-server=<url>` или `$WERF_KUBE_TLS_SERVER=<url>`: изменить имя сервера, используемое для валидации сертификата Kubernetes API сервера.
1. `--kube-ca-path=<путь>` или `$WERF_KUBE_CA_PATH=<путь>`: изменить путь к CA-сертификату.
1. `--skip-tls-verify-kube=<bool>` или `$WERF_SKIP_TLS_VERIFY_KUBE=<bool>`: нужно ли валидировать TLS-сертификат Kubernetes API сервера.

## Доступ к Helm-чартам или werf-бандлам в приватном репозитории

Используйте `werf cr login` для аутентификации в приватном OCI-репозитории с Helm-чартами или werf-бандлами:

```shell
werf cr login -u myuser -p mypassword localhost:5000
```

А для приватного HTTP-репозитория Helm-чартов используйте `werf helm registry login`:

```shell
werf helm registry login -u myuser -p mypassword localhost:5000
```

Теперь вы можете скачивать или загружать Helm-чарты или werf-бандлы из репозитория, будь то с помощью `werf helm dependency build/update`, `werf helm pull`, `werf bundle publish/apply` или других команд.
