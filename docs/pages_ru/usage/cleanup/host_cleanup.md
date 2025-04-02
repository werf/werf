---
title: Автоматическая очистка хоста
permalink: usage/cleanup/host_cleanup.html
---

## Обзор

Очистка хоста удаляет неактуальных данные и сокращает размер кеша **автоматически** в рамках вызова основных команд werf и сразу **для всех проектов**. При необходимости очистку можно выполнять в ручном режиме с помощью команды [**werf host cleanup**]({{ "reference/cli/werf_host_cleanup.html" | true_relative_url }}).

## Переопределение директории хранилища бэкенда (Docker или Buildah)

Параметр `--backend-storage-path` (или переменная окружения `WERF_BACKEND_STORAGE_PATH`) позволяет явно задать директорию хранилища бэкенда в случае, если werf не может правильно определить её автоматически.

## Изменение порога занимаемого места и глубины очистки хранилища бэкенда (Docker или Buildah)

Параметр `--allowed-backend-storage-volume-usage` (`WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE`) позволяет изменить порог занимаемого места на томе, при достижении которого выполняется очистка хранилища бэкенда (по умолчанию 70%).

Параметр `--allowed-backend-storage-volume-usage-margin` (`WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE_MARGIN`) позволяет установить глубину очистки относительно установленного порога занимаемого места хранилища бэкенда (по умолчанию 5%).

## Сборочный кеш (Docker или Buildah)

werf не чистит сборочный кеш бекендов **Buildah** и **Docker**.

Для бекенда **Docker** (buildx / buildkit) пользователю доступны следующие стратегии очистки сборочного кеша:

+ Определить в [настройках](https://docs.docker.com/build/cache/garbage-collection/#configuration) сборщика порог хранимого кеша.
+ Настроить очистку по расписанию с использованием `cron` и `docker buildx prune --max-used-space=bytes`.

## Изменение порога занимаемого места и глубины очистки локального кэша

Параметр `--allowed-local-cache-volume-usage` (`WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE`) позволяет изменить порог занимаемого места на томе, при достижении которого выполняется очистка локального кэша (по умолчанию 70%).

Параметр `--allowed-local-storage-volume-usage-margin` (`WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN`) позволяет установить глубину очистки относительно установленного порога занимаемого места локального кэша (по умолчанию 5%).

## Выключение автоматической очистки

Пользователь может выключить автоматическую очистку неактуальных данных хоста с помощью параметра `--disable-auto-host-cleanup` (`WERF_DISABLE_AUTO_HOST_CLEANUP`). В этом случае рекомендуется добавить команду `werf host cleanup` в cron, например, следующим образом:

```shell
# /etc/cron.d/werf-host-cleanup
SHELL=/bin/bash
*/30 * * * * gitlab-runner source ~/.profile ; source $(trdl use werf 2 stable) ; werf host cleanup
```
