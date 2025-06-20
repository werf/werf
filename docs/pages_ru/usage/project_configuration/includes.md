---
title: Подключение конфигурации
permalink: usage/project_configuration/includes.html
---

## Обзор

`werf` поддерживает импорт конфигурационных файлов из внешних источников с помощью механизма **includes**.   

Это удобно, когда у вас много однотипных приложений и вы хотите сопровождать и использовать конфигурацию централизованно.

### Что можно импортировать

Поддерживается импорт следующих типов файлов:

* Конфигурационные файлы `werf.yaml` и шаблоны `.werf/**/*.tmpl`, а также произвольные файлы, которые используются при шаблонизации.
* Helm-чарты и отдельные файлы.
* Dockerfile и `.dockerignore`.

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
2. Если файл присутствует в нескольких includes, будет использоваться версия из источника, который расположен ниже в списке `includes` в `werf-includes.yaml` (от общего к частному).

Ниже приведён пример конфигурации импортирования схожих конфигураций:

```yaml
# werf-includes.yaml
includes:
  - git: https://github.com/werf/examples
    branch: main
    add: local-dev
    to: /
    includePaths:
      - /.helm

  - git: https://example.com/helm_examples
    branch: main
    add: /
    to: /
    includePaths:
      - /.helm
````

Структура директорий:

* В `examples`: `.helm/Chart.yaml`, `.helm/values.yaml`
* В `helm_examples`: `.helm/Chart.yaml`

Результат `werf includes ls-files`:

```
.helm/Chart.yaml        https://example.com/werf/helm_examples
.helm/values.yaml       https://github.com/werf/examples
```

Файл `Chart.yaml` взят из последнего источника (`helm_examples`) согласно правилам наложения.

## Использование конфигурации из внешних источников

Ниже приведён пример конфигурации `werf-includes.yaml`.
Полное описание директив вы можете найти на соответствующей странице [werf-includes.yaml]({{"reference/werf_includes_yaml.html" | true_relative_url }})

```yaml
includes:
  - git: https://github.com/werf/examples
    branch: main
    add: local-dev
    to: /
    includePaths:
      - /.helm
```

После того как вы зафиксируете версии includes командой `werf includes update`, в корне проекта будет создан файл `werf-includes.lock`, который выглядит следующим образом:

```yaml
includes:
  - git: https://github.com/werf/examples
    branch: main
    commit: 21640b8e619ba4dd480fedf144f7424aa217a2eb
```

> **ВАЖНО.** Согласно политикам гитерминизма, файлы `werf-includes.yaml` и `werf-includes.lock` должны быть закомичены. Для первичной конфигурации и отладки мы предлагаем использовать флаг `--dev`. После завершения конфигурации файлы необходимо закоммитить.

## Обновление

### Детерминированное обновление версий

Существует два способа обновления версий includes:

* Редактирование файла `werf-includes.lock` вручную или с помощью средств поддержания зависимостей, таких как dependabot, renovate и прочие.
* Командой `werf includes update`. Данная команда обновит все includes на `HEAD` соответствующего референса (`branch` или `tag`).

### Автообновление версий (не рекомендовано)

Если необходимо использовать последние `HEAD`-версии без lock-файла, например, для быстрой проверки изменений, можно использовать опцию `--allow-includes-update`, а так же явно разрешить ее использование в `werf-giterminism.yaml`:

```yaml
includes:
  allowIncludesUpdate: true
```

> **ВАЖНО.** Мы не рекомендуем использование данного подхода, так как он может нарушить воспроизводимость сборок и развертываний.

## Отладка

### Получение списка файлов проекта с учётом источников

```bash
werf includes ls-files .helm
```

Пример вывода:

```
PATH                      SOURCE
.helm/Chart.yaml          local
.helm/values.yaml         https://github.com/werf/examples
```

### Получение содержимого файла проекта с учётом источников

```bash
werf includes get-file .helm/values.yaml
```

Пример вывода:

```yaml
backend:
  limits:
    cpu: 100m
    memory: 256Mi
```

### Использование локальных репозиториев

Допускается использование локальных репозиториев в качестве источника для include.
Порядок работы с такими репозиториями ничем не отличается от работы с удалёнными.

```yaml
# werf-includes.yaml
includes:
  - git: /path/to/repo
    branch: main
    add: local-dev
    to: /
    includePaths:
      - /.helm
```
