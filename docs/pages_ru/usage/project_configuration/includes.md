# Импортирование конфигураций из удаленных репозиториев

**`werf`** поддерживает импорт конфигурационных файлов и шаблонов проекта из удалённых Git-репозиториев (includes). Это помогает избежать дублирования кода при сопровождении однотипных приложений. Импорт осуществляется с помощью виртуальной файловой системы, которая работает по определённым правилам наложения и ограничениям.

## Конфигурация

Конфигурация includes описывается в файле `werf-includes.yaml`, который должен находиться в корне проекта. Этот файл не подлежит импорту и обрабатывается в соответствии с политиками гитерминизма.

Пример `werf-includes.yaml`:

```yaml
includes:
  - git: https://github.com/werf/examples
    branch: main
    add: local-dev
    to: /
    includePaths:
      - /.helm

  - git: https://github.com/werf/helm_examples
    tag: v0.0.1
    add: /
    to: examples/helm

  - git: https://github.com/werf/werf_examples
    commit: 831fb9ff936ed6f1db5175c3e65891cfd25580dd
    add: /
    to: /
    includePaths:
      - /.werf
    excludePaths:
      - /.helm
```

## Фиксация версий

Для обеспечения предсказуемости и воспроизводимости сборок и развёртываний необходимо зафиксировать используемые коммиты веток или тегов в lock-файле `werf-includes.lock`. Его можно сгенерировать или обновить с помощью команды:

```bash
werf includes update
```

В lock-файл будут записаны текущие значения `HEAD` для каждого указанного источника.

Пример `werf-includes.lock`:

```yaml
includes:
  - git: https://github.com/werf/examples
    branch: main
    commit: 21640b8e619ba4dd480fedf144f7424aa217a2eb

  - git: https://github.com/werf/helm_examples
    tag: v0.0.1
    commit: e9eff747f82d6959d1c0b4da284e23fd650f4be3

  - git: https://github.com/werf/werf_examples
    commit: 831fb9ff936ed6f1db5175c3e65891cfd25580dd
```

> **Важно.** Lock-файл также не может быть импортирован в другие проекты и подчиняется политике гитерминизма.

Если необходимо использовать последние версии включаемых источников без фиксированных коммитов, можно разрешить обновление includes в файле `werf-giterminism.yaml`:

```yaml
includes:
  allowIncludesUpdate: true
```

Однако такой подход не обеспечивает воспроизводимость сборок и развёртываний и не рекомендуется для использования.


## Ограничения

### Что можно импортировать:

* файлы шаблонов `.werf`
* чарты Helm из директории `.helm`
* основной конфигурационный файл `werf.yaml`
* `Dockerfile` и `.dockerignore`

### Что **нельзя** импортировать:

* `werf-includes.yaml` и `werf-includes.lock`
* файл гитерминизма `werf-giterminism.yaml`

> **Важно.** Импорт работает **только** на уровне конфигурации и **не добавляет** файлы из внешних репозиториев в сборочный контекст Docker.

## Правила наложения

Если в нескольких источниках имеются файлы с одинаковыми путями, применяются следующие правила:

1. Если файл существует в локальном проекте — **он имеет приоритет** и используется вместо импортированного.
2. Если файл присутствует в нескольких includes — будет использован файл из того источника, который **расположен выше** в списке `includes` в `werf-includes.yaml`.

Пример:

```yaml
includes:
  - git: https://github.com/werf/examples
    branch: main
    add: local-dev
    to: /
    includePaths:
      - /.helm

  - git: https://github.com/werf/helm_examples
    branch: main
    to: /
    includePaths:
      - /.helm
```

## Отладка

Рассмотрим пример наложения содержимого директорий `.helm` из двух разных репозиториев:

```yaml
includes:
  - git: https://github.com/werf/examples
    branch: main
    add: local-dev
    to: /
    includePaths:
      - /.helm

  - git: https://github.com/werf/helm_examples
    branch: main
    to: /
    includePaths:
      - /.helm
```

Структуры директорий в исходных репозиториях:

**`https://github.com/werf/examples`**

```bash
.helm/
└── Chart.yaml
```

**`https://github.com/werf/helm_examples`**

```bash
.helm/
├── Chart.yaml
└── values.yaml
```

Команда для просмотра всех импортированных файлов:

```bash
werf includes ls-files .helm
```

Пример вывода:

```
PATH                                                SOURCE
.helm/Chart.yaml                                    https://github.com/werf/examples
.helm/values.yaml                                   https://github.com/werf/helm_examples
```

Как видно, файл `.helm/Chart.yaml` был взят из `examples`, так как этот источник указан первым в списке `includes`.

Команда для просмотра содержимого импортированного файла:

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
