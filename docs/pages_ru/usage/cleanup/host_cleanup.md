---
title: Автоматическая очистка хоста
permalink: usage/cleanup/host_cleanup.html
---

## Обзор

Очистка хоста удаляет неактуальных данные и сокращает размер кеша **автоматически** в рамках вызова основных команд werf и сразу **для всех проектов**. При необходимости очистку можно выполнять в ручном режиме с помощью команды [**werf host cleanup**]({{ "reference/cli/werf_host_cleanup.html" | true_relative_url }}).

## Переопределение директории хранилища Docker

Параметр `--docker-server-storage-path` (или переменная окружения `WERF_DOCKER_SERVER_STORAGE_PATH`) позволяет явно задать директорию хранилища Docker в случае, если werf не может правильно определить её автоматически.

## Изменение порога занимаемого места и глубины очистки хранилища Docker

Параметр `--allowed-docker-storage-volume-usage` (`WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE`) позволяет изменить порог занимаемого места на томе, при достижении которого выполняется очистка хранилища Docker (по умолчанию 70%).

Параметр `--allowed-docker-storage-volume-usage-margin` (`WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE_MARGIN`) позволяет установить глубину очистки относительно установленного порога занимаемого места хранилища Docker (по умолчанию 5%).

## Изменение порога занимаемого места и глубины очистки локального кэша

Параметр `--allowed-local-cache-volume-usage` (`WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE`) позволяет изменить порог занимаемого места на томе, при достижении которого выполняется очистка локального кэша (по умолчанию 70%).

Параметр `--allowed-docker-storage-volume-usage-margin` (`WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN`) позволяет установить глубину очистки относительно установленного порога занимаемого места локального кэша (по умолчанию 5%).

## Выключение автоматической очистки

Пользователь может выключить автоматическую очистку неактуальных данных хоста с помощью параметра `--disable-auto-host-cleanup` (`WERF_DISABLE_AUTO_HOST_CLEANUP`). В этом случае рекомендуется добавить команду `werf host cleanup` в cron, например, следующим образом:

```shell
# /etc/cron.d/werf-host-cleanup
SHELL=/bin/bash
*/30 * * * * gitlab-runner source ~/.profile ; source $(trdl use werf 1.2 stable) ; werf host cleanup
```
