---
title: Управление выкатом
sidebar: doc_sidebar
permalink: deploy_for_kube.html
folder: kube
---

Для деплоя в kubernetes используется [helm](https://helm.sh/) (kubernetes package manager).

В директории `.helm` в корне проекта описывается [helm chart](https://github.com/kubernetes/helm/blob/master/docs/charts.md#charts):

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

Структура chart-а включает в себя дополнительные файлы `secret-values.yaml`, директорию и `secret`, подробнее о которых см. в разделе [работа с секретами](secrets_for_kube.html).

## Настройки подключения к kubernetes

Подключение к kubernetes настраивается через тот же конфигурационный файл, что и kubectl: `~/.kube/config`.

* Используется контекст `current-context`, если он установлен или первый попавшийся контекст из списка `contexts`.
* Используется тот же kubernetes namespace по умолчанию, что и kubectl: из поля `namespace` активного контекста.
  * Если namespace по умолчанию не указан в `~/.kube/config`, то используется namespace=`default`.

## Управление выкатом

### dapp kube deploy

```
dapp kube deploy REPO [--tag=TAG --tag-branch --tag-commit --tag-build-id --tag-ci] [--namespace=NAMESPACE] [--set=<value>] [--values=<values-path>] [--secret-values=<secret-values-path>]
```

Запускает процесс выката helm-chart'а в kubernetes.

В helm будет создан или обновлен релиз с именем [`имя dapp`](definitions.html#имя-dapp)-\<NAMESPACE\>.

##### `REPO`

Адрес репозитория, из которого будут взяты образы. Данный параметр должен совпадать с параметром, указываемым в [`dapp dimg push`](base_commands.html#dapp-dimg-push).

При указании специального значения `:minikube` будет использован локальный proxy для docker-registry из minikube, см. [использование minikube](minikube_for_kube.html).

##### `--tag=TAG --tag-branch --tag-commit --tag-build-id --tag-ci`

Версия образа из указанного репозитория. Опции соответствуют указываемым в [`dapp dimg push`](base_commands.html#dapp-dimg-push).

##### `--namespace=NAMESPACE`

Использовать указанный kubernetes namespace. Если не указан, то будет использован namespace по умолчанию в соответствии с `~/.kube/config` или, если не указано, namespace=`default`.

##### `--set=<value>`

Передается без изменений в параметр [`helm --set`](https://github.com/kubernetes/helm/blob/master/docs/chart_template_guide/values_files.md#values-files).

##### `--values=<values-path>`

Передается без изменений в параметр [`helm --values`](https://github.com/kubernetes/helm/blob/master/docs/chart_template_guide/values_files.md#values-files).

Позволяет указать дополнительный values yaml файл, помимо стандартного `.helm/values.yaml`.

##### `--secret-values=<secret-values-path>`

Позволяет указать дополнительный secret-values yaml файл, помимо стандартного `.helm/secret-values.yaml`. Подробнее о секретах см. в разделе [работе с секретами](secrets_for_kube.html).

### dapp kube dismiss

```
dapp kube dismiss [--namespace=NAMESPACE] [--with-namespace]
```

Запускает процесс удаления релиза [`<имя dapp>`](definitions.html#имя-dapp)-`<NAMESPACE>` из helm.

##### `--namespace=NAMESPACE`

Использовать указанный kubernetes namespace. Если не указан, то будет использован namespace по умолчанию в соответствии с `~/.kube/config` или, если не указано, namespace=`default`.

##### `--with-namespace`

Удалить используемый kubernetes namespace после удаления релиза из helm. По умолчанию namespace удален не будет.
