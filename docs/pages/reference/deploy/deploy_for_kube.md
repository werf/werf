---
title: Управление выкатом
sidebar: reference
permalink: reference/deploy/deploy_for_kube.html
---

Для деплоя в kubernetes используется [helm](https://helm.sh/) (kubernetes package manager).

## Установка helm

Перед тем как использовать dapp для деплоя, необходимо [установить helm](https://docs.helm.sh/using_helm/#installing-helm) и его [серверную часть - tiller](https://docs.helm.sh/using_helm/#installing-tiller). Затем, необходимо установить для helm модуль template:

```
helm plugin install https://github.com/technosophos/helm-template
```

## Helm chart

В директории `.helm` в корне проекта описывается [helm chart](https://github.com/kubernetes/helm/blob/master/docs/charts.md#charts) (далее - chart, чарт), который описывает конфигурацию приложения и его компонент для последующего выката в кластер kubernetes средствами helm. Структура chart'а для dapp, т.е. папки `.helm`, следующая:

```
.helm/
  templates/
    <name>.yaml
    <name>.tpl
  charts/
  secret/
  Chart.yaml
  values.yaml
  secret-values.yaml
```

Файл `Chart.yaml` — это описание чарта приложения и в нем как минимум нужно указать имя приложения и его версию. Пример файла `Chart.yaml`:
```
apiVersion: v1 [ apiVersion: v1]
description: Test RabbitMQ chart for Kubernetes
name: rabbit
version: 0.1.0
```

В директории `templates` хранятся шаблоны YAML-файлов - элементов чарта - с описанием ресурсов для их размещения в кластере. Подробнее создание шаблонов рассматривается в отдельном [разделе]({{ site.baseurl }}/reference/deploy/templates.html). Директория `charts` используется при необходимости работы с внешними чартами.

Структура чарта включает в себя дополнительные элементы которых нет в структуре обычного helm chart, - это файл `secret-values.yaml`, и директория `secret`, подробнее о которых см. в разделе [работа с секретами]({{ site.baseurl }}/reference/deploy/secrets.html).

## Настройки подключения к kubernetes

Подключение к kubernetes настраивается через тот же конфигурационный файл, что и kubectl: `~/.kube/config`.

* Используется контекст `current-context`, если он установлен или первый попавшийся контекст из списка `contexts`.
* Используется тот же kubernetes namespace по умолчанию, что и kubectl: из поля `namespace` активного контекста.
  * Если namespace по умолчанию не указан в `~/.kube/config`, то используется namespace=`default`.

Если вы используете самоподписанный SSL-сертификат, то указать его можно с помощью переменной `SSL_CERT_FILE`.
* Скопируйте фаил сертификата в какую-либо папку и установите необходимые права доступа. Пример:
```bash
cp  CERTFILE.crt ~gitlab-runner/.
chown -R gitlab-runner:gitlab-runner ~gitlab-runner/CERTFILE.crt
```
* Установите переменную окружения `SSL_CERT_FILE=/home/gitlab-runner/<domain>.crt)` (экспортируйте ее, либо установите непосредственно перед вызовом `dapp kube deploy`). Пример
```bash
SSL_CERT_FILE=/home/gitlab-runner/CERTFILE.crt dapp kube deploy
      --namespace ${CI_ENVIRONMENT_SLUG}
      --tag-ci
      ${CI_REGISTRY_IMAGE}
```


## Управление выкатом

### dapp kube deploy

```
dapp kube deploy REPO [--tag=TAG --tag-branch --tag-commit --tag-build-id --tag-ci] [--namespace=NAMESPACE] [--set=<value>] [--values=<values-path>] [--secret-values=<secret-values-path>]
```

Запускает процесс выката helm-chart'а в kubernetes.

В helm будет создан или обновлен релиз с именем `имя dapp`-\<NAMESPACE\>.

##### `REPO`

Адрес репозитория, из которого будут взяты образы. Данный параметр должен совпадать с параметром, указываемым в [`dapp dimg push`]({{ site.baseurl }}/reference/cli/dimg_push.html).

При указании специального значения `:minikube` будет использован локальный proxy для docker-registry из minikube, см. [использование minikube]({{ site.baseurl }}/reference/deploy/minikube.html).

##### `--tag=TAG --tag-branch --tag-commit --tag-build-id --tag-ci`

Версия образа из указанного репозитория. Опции соответствуют указываемым в [`dapp dimg push`]({{ site.baseurl }}/reference/cli/dimg_push.html).

##### `--namespace=NAMESPACE`

Использовать указанный kubernetes namespace. Если не указан, то будет использован namespace по умолчанию в соответствии с `~/.kube/config` или, если не указано, namespace=`default`.

##### `--set=<value>`

Передается без изменений в параметр [`helm --set`](https://github.com/kubernetes/helm/blob/master/docs/chart_template_guide/values_files.md#values-files).

##### `--values=<values-path>`

Передается без изменений в параметр [`helm --values`](https://github.com/kubernetes/helm/blob/master/docs/chart_template_guide/values_files.md#values-files).

Позволяет указать дополнительный values yaml файл, помимо стандартного `.helm/values.yaml`.

##### `--secret-values=<secret-values-path>`

Позволяет указать дополнительный secret-values yaml файл, помимо стандартного `.helm/secret-values.yaml`. Подробнее о секретах см. в разделе [работе с секретами]({{ site.baseurl }}/reference/deploy/secrets.html).

### dapp kube dismiss

```
dapp kube dismiss [--namespace=NAMESPACE] [--with-namespace]
```

Запускает процесс удаления релиза `<имя dapp>`-`<NAMESPACE>` из helm.

##### `--namespace=NAMESPACE`

Использовать указанный kubernetes namespace. Если не указан, то будет использован namespace по умолчанию в соответствии с `~/.kube/config` или, если не указано, namespace=`default`.

##### `--with-namespace`

Удалить используемый kubernetes namespace после удаления релиза из helm. По умолчанию namespace удален не будет.
