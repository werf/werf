---
title: Подключение конфигурации
permalink: usage/project_configuration/includes.html
---

## Обзор

`werf` поддерживает импорт конфигурационных файлов из внешних источников с помощью механизма **includes**.   

Это удобно, когда у вас много однотипных приложений и вы хотите сопровождать и использовать конфигурацию централизованно.

### Что можно импортировать

Поддерживается импорт следующих типов файлов:

* Основной файл конфигурации (`werf.yaml`), шаблоны (`.werf/**/*.tmpl`), а также произвольные файлы, которые используются при шаблонизации.
* Helm-чарты и отдельные файлы.
* `Dockerfile` и `.dockerignore`.

### Что нельзя импортировать

Механизм не позволяет импортировать файлы сборочного контекста, а также следующие конфигурационные файлы:

* `werf-includes.yaml`
* `werf-includes.lock`
* `werf-giterminism.yaml`

### Как работает механизм includes

1. Вы описываете внешние источники и импортируемые файлы в `werf-includes.yaml`.
2. Фиксируете версии внешних источников с помощью команды `werf includes update`, которая создаёт `werf-includes.lock`. Lock-файл необходим для воспроизводимости сборок и развертываний.
3. Все команды `werf` (например, `werf build`, `werf converge`) будут учитывать этот lock-файл и работать с конфигурацией в соответствии с правилами наложения.

### Правила наложения

Если один и тот же файл существует в нескольких местах, применяются следующие правила:

1. Локальные файлы проекта всегда имеют приоритет над импортируемыми.
2. Если файл присутствует в нескольких `includes`, будет использоваться версия из источника, который расположен ниже по списку `includes` в `werf-includes.yaml` (от общего к частному).

#### Пример применения правил наложения

Допустим в локальном репозитории есть файл `c`, а также следующая конфигурация:

```yaml
# werf-includes.yaml
includes:
  - git: repo1 # импортирует a, b, c
    branch: main
    add: /
    to: /
  - git: repo2 # импортирует b, c
    branch: main
    add: /
    to: /
```

Тогда результат наложения будет выглядеть так:

```bash
.
├── a      # из repo1
├── b      # из repo2
├── c      # из локального репо
```

## Подключение конфигурации из внешних источников

Ниже приведён пример конфигурации внешних источников (полное описание директив вы можете найти на соответствующей странице [werf-includes.yaml]({{"reference/werf_includes_yaml.html" | true_relative_url }})):

```yaml
# werf-includes.yaml
includes:
  - git: https://github.com/werf/werf
    branch: main
    add: /docs/examples/includes/common
    to: /
    includePaths:
      - .werf
  - git: https://gitlab.company.name/common/helper-utils.git
    basicAuth:
      username: token
      password:
        env: GITLAB_TOKEN
    commit: fedf144f7424aa217a2eb21640b8e619ba4dd480
    add: /.helm
    to: /
```

Порядок работы с приватными репозитриями аналогичен описанному [здесь]({{ "/usage/build/stapel/git.html#работа-с-удаленными-репозиториями" | true_relative_url }}).

После конфигурации необходимо зафиксировать версии внешних зависимостей. Для этого можно использовать команду `werf includes update`, после вызова которой, в корне проекта будет создан файл `werf-includes.lock`:

```yaml
includes:
  - git: https://github.com/werf/werf
    branch: main
    commit: 21640b8e619ba4dd480fedf144f7424aa217a2eb
  - git: https://gitlab.company.name/common/helper-utils.git
    commit: fedf144f7424aa217a2eb21640b8e619ba4dd480
```

> **ВАЖНО.** Согласно политикам гитерминизма, файлы `werf-includes.yaml` и `werf-includes.lock` должны быть закомичены. При конфигурации и отладке для удобства предлагается использовать флаг `--dev`.

### Пример использования внешних источников при конфигурации однотипных приложений

Предположим вы решили централизованно обслуживать конфигурацию приложений в вашей организации, сохраняя общую конфигурацию в одном источнике и конфигурацию под каждый тип проекта в других (в одном или нескольких Git-репозиториях — на ваше усмотрение). 

Конфигурация типового проекта в таком случае могла бы выглядеть так:

* Проект:
{% tree_file_viewer 'examples/includes/app' default_file='werf-includes.yaml' %}

* Источник с общей конфигурацией:
{% tree_file_viewer 'examples/includes/werf-common' default_file='.werf/cleanup.tpl' %}

* Источник со специфичной для вашего типа проекта конфигурацией:
{% tree_file_viewer 'examples/includes/common_js' default_file='backend.Dockerfile' %}

## Обновление `includes`

### Детерминированное

Существует два способа обновления версий includes:

* Командой `werf includes update`. Данная команда обновит все `includes` на `HEAD` соответствующего референса (`branch` или `tag`).
* Редактирование файла `werf-includes.lock` вручную или с помощью таких как инструментов как `dependabot`, `renovate` и прочих.

### Автообновление (не рекомендовано)

Если необходимо использовать последние `HEAD`-версии без lock-файла, можно использовать опцию `--allow-includes-update`, а так же явно разрешить ее использование в `werf-giterminism.yaml`:

```yaml
includes:
  allowIncludesUpdate: true
```

> **ВАЖНО.** Мы не рекомендуем использование данного подхода, так как он может нарушать воспроизводимость сборок и развертываний.

## Отладка

### Получение списка файлов проекта с учётом источников

```bash
werf includes ls-files .helm
```

Пример вывода:

```
PATH                                             SOURCE
.helm/Chart.yaml                                 https://github.com/werf/werf
.helm/charts/backend/Chart.yaml                  https://github.com/werf/werf
.helm/charts/backend/templates/deployment.yaml   https://github.com/werf/werf
.helm/charts/backend/templates/service.yaml      https://github.com/werf/werf
.helm/charts/frontend/Chart.yaml                 https://github.com/werf/werf
.helm/charts/frontend/templates/deployment.yaml  https://github.com/werf/werf
.helm/charts/frontend/templates/service.yaml     https://github.com/werf/werf
.helm/requirements.lock                          https://github.com/werf/werf
.helm/values.yaml                                local
```

### Получение содержимого файла проекта с учётом источников

```bash
werf includes get-file .helm/charts/backend/templates/deployment.yaml
```

Пример вывода:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ $.Chart.Name }}
  annotations:
    werf.io/weight: "30"
spec:
  selector:
    matchLabels:
      app: {{ $.Chart.Name }}
  template:
    metadata:
      labels:
        app: {{ $.Chart.Name }}
    spec:
      containers:
        - name: backend
          image: {{ $.Values.werf.image.backend }}
          ports:
            - containerPort: 3000
          resources:
            requests:
              cpu: {{ $.Values.limits.cpu }}
              memory: {{ $.Values.limits.memory}}
            limits:
              cpu: {{ $.Values.limits.cpu }}
              memory: {{ $.Values.limits.memory }}
```

### Использование локальных репозиториев

Допускается использование локальных репозиториев в качестве источника для include.
Порядок работы с такими репозиториями ничем не отличается от работы с удалёнными.

```yaml
# werf-includes.yaml
includes:
  - git: /local/repo
    branch: main
    add: local-dev
    to: /
    includePaths:
      - /.helm
```
