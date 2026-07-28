---
title: Переменные окружения для отладки
permalink: reference/debug_env_vars.html
---

werf поддерживает набор переменных окружения, включающих дополнительный отладочный вывод. Они предназначены для диагностики проблем: формат их вывода не является частью стабильного интерфейса и может меняться между релизами.

## Уровни логирования

| Переменная | Значение | Описание |
|------------|----------|----------|
| `WERF_LOG_DEBUG` (синоним: `WERF_DEBUG`) | boolean | То же, что `--log-debug`: включает debug-уровень логирования |
| `WERF_LOG_VERBOSE` (синоним: `WERF_VERBOSE`) | boolean | То же, что `--log-verbose`: включает verbose-уровень логирования |

## Тематические отладочные каналы

Переменные ниже включают узкие отладочные каналы. Каналы, пишущие на debug-уровень, дополнительно требуют `--log-debug` (или `WERF_LOG_DEBUG=1`) — одной тематической переменной недостаточно. Если не указано иное, переменная включается значением `1`.

### Сборка

| Переменная | Описание |
|------------|----------|
| `WERF_DEBUG_STAGES_STORAGE` | Внутренности хранилища стадий: трейсы вызовов методов хранилища, операции manifest cache, перечисление стейдж-тегов, дампы описаний выбранных стадий и базового образа. Требует `--log-debug` |
| `WERF_DEBUG_CONVEYOR_PHASES` | Обёртки фаз сборочного конвейера (BeforeImages, OnImageStage, AfterImages и др.). Требует `--log-debug` |
| `WERF_DEBUG_IMAGE_SPEC` | YAML-снимки конфигурации image spec (информация об исходном образе, подготовленный и публикуемый конфиг). Требует `--log-debug` |
| `WERF_DEBUG_IMPORT_SERVER` | Трейсы rsync-сервера импортов. Требует `--log-debug` |
| `WERF_DEBUG_STAGE_DIGEST` | Аргументы расчёта digest'а стадии (именованные значения, из которых он вычисляется). Сам итоговый digest логируется при `--log-debug` независимо от этой переменной. Требует `--log-debug` |
| `WERF_DEBUG_USER_STAGE_CHECKSUM` | Детали расчёта контрольной суммы пользовательских стадий. Требует `--log-debug` |
| `WERF_DEBUG_IMPORT_SOURCE_CHECKSUM` | Детали расчёта контрольной суммы источников импорта. Требует `--log-debug` |
| `WERF_DEBUG_DOCKERFILE_STAGE_DEPENDENCIES` | Управляет двумя каналами: расчёт зависимостей стадий Dockerfile и перечисление совпавших путей build-контекста при расчёте его контрольной суммы. Требует `--log-debug` |
| `WERF_DEBUG_CONTAINER_RUNTIME` (устаревший синоним: `WERF_CONTAINER_RUNTIME_DEBUG`) | Внутренности сборочного бэкенда: удаление путей в контейнере, mount/unmount buildah-контейнеров, пофайловый расчёт контрольных сумм. Требует `--log-debug` |
| `WERF_DEBUG_BUILDAH` (устаревший синоним: `WERF_BUILDAH_DEBUG`) | Отладочный вывод бэкенда Buildah |
| `WERF_DEBUG_DOCKER` | Debug-режим Docker CLI |
| `WERF_DEBUG_DOCKER_RUN_COMMAND` | Печать полных команд `docker run` |

### Git

| Переменная | Описание |
|------------|----------|
| `WERF_DEBUG_TRUE_GIT` | Трассировка выполнения git-команд. Требует `--log-debug` |
| `WERF_DEBUG_TRUE_GIT_ARCHIVE` (устаревший синоним: `WERF_TRUE_GIT_DEBUG_ARCHIVE`) | Пофайловый вывод при создании git-архивов. Требует `--log-debug` |
| `WERF_DEBUG_TRUE_GIT_PATCH` (устаревший синоним: `WERF_TRUE_GIT_DEBUG_PATCH`) | Отладочный вывод создания git-патчей |
| `WERF_DEBUG_TRUE_GIT_PATCH_PARSER` (устаревший синоним: `WERF_TRUE_GIT_DEBUG_PATCH_PARSER`) | Отладочный вывод парсера git-патчей |
| `WERF_DEBUG_LS_TREE_PROCESS` | Обход git ls-tree и пофайловые записи при расчёте контрольных сумм. Требует `--log-debug` |
| `WERF_DEBUG_GIT_STATUS` | Детали результата git status. Требует `--log-debug` |
| `WERF_DEBUG_GITERMINISM_MANAGER` | Операции чтения файлов giterminism manager. Требует `--log-debug` |

### Container registry

| Переменная | Описание |
|------------|----------|
| `WERF_DEBUG_DOCKER_REGISTRY` (устаревший синоним: `WERF_DOCKER_REGISTRY_DEBUG`) | Трассировка вызовов API registry и progress-тики push/pull. Для progress-тиков требует `--log-debug` |
| `WERF_DEBUG_DOCKER_REGISTRY_API` | Отладочные логи библиотеки go-containerregistry |

### Конфигурация и шаблоны

| Переменная | Описание |
|------------|----------|
| `WERF_DEBUG_TEMPLATES` | boolean; то же, что `--debug-templates`: debug-режим Go-шаблонов |
| `WERF_DEBUG_SECRET_VALUES` | Отладочный вывод декодирования секретных значений. Требует `--log-debug` |

### Деплой

| Переменная | Описание |
|------------|----------|
| `WERF_NELM_TRACE` | boolean; включает trace-уровень логирования движка деплоя Nelm |
| `WERF_DEBUG_HELM_V3_EXTRA_ANNOTATIONS_AND_LABELS` (устаревший синоним: `WERF_HELM_V3_EXTRA_ANNOTATIONS_AND_LABELS_DEBUG`) | Отладочный вывод post-renderer'а дополнительных аннотаций и лейблов |
| `WERF_SHOW_VERBOSE_DIFFS` | boolean, по умолчанию `true`; то же, что `--show-verbose-diffs` в `werf plan`: подробные строки диффов |
| `WERF_SHOW_VERBOSE_CRD_DIFFS` | boolean, по умолчанию `false`; то же, что `--show-verbose-crd-diffs` в `werf plan`: подробные строки диффов CRD |

### Сборщик Ansible

| Переменная | Описание |
|------------|----------|
| `WERF_DEBUG_ANSIBLE_ARGS` | Строка с дополнительными аргументами для `ansible-playbook` |
| `WERF_DEBUG_ANSIBLE_LIVE_PY_PATH` | Путь к файлу, заменяющему встроенный ansible-колбэк `live.py` |
| `WERF_DEBUG_ANSIBLE_WERF_PY_PATH` | Путь к файлу, заменяющему встроенный ansible-колбэк `werf.py` |

### Диагностика процесса

| Переменная | Описание |
|------------|----------|
| `WERF_PRINT_STACK_TRACES` | Периодическая печать стек-трейсов горутин |
| `WERF_PRINT_STACK_TRACES_PERIOD` | Целое число секунд между печатью стек-трейсов (по умолчанию `5`); действует только вместе с `WERF_PRINT_STACK_TRACES=1` |
