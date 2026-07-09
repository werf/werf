---
title: Сборочные бэкенды
permalink: usage/build/backends.html
---

## Обзор

werf позволяет работать со следующими сборочными бэкендами:

-	Docker — традиционный способ, использующий системный Docker Daemon. Используется по умолчанию, если BuildKit-эндпоинт не задан. Поддерживает только Dockerfile-сборки.
-	BuildKit — сборка образов через внешний демон [buildkitd](https://github.com/moby/buildkit). Включается заданием BuildKit-эндпоинта через переменную окружения. Необходим для stapel-сборок.

> Необходимые требования и подготовка системы для работы со сборочными бэкендами описаны в разделе сайта [Быстрый старт]({{ site.url }}/getting_started/).

## BuildKit

Бэкенд BuildKit включается установкой переменной окружения `WERF_BUILDKIT_HOST` (или стандартной `BUILDKIT_HOST`) в адрес запущенного демона buildkitd. Если ни одна из этих переменных не задана, werf использует Docker-бэкенд.

### Эндпоинты

Поддерживаются следующие схемы эндпоинтов:

*	`unix://` — локальный Unix-сокет;
*	`tcp://` — TCP-эндпоинт;
*	`docker-container://` — buildkitd, запущенный в Docker-контейнере;
*	`kube-pod://` — buildkitd, запущенный в Kubernetes-поде;
*	`podman-container://` — buildkitd, запущенный в Podman-контейнере;
*	`ssh://` — buildkitd, доступный по SSH.

### Быстрый старт

Запустите buildkitd в локальном Docker-контейнере и укажите его адрес werf:

```shell
docker run -d --name buildkitd --privileged moby/buildkit
export BUILDKIT_HOST=docker-container://buildkitd
```

После этого любая команда сборки werf будет использовать BuildKit-бэкенд.

### Требуется container registry

BuildKit-бэкенд требует наличия удалённого container registry. Опция `--repo` (или переменная `WERF_REPO`) обязательна — локальное хранилище стадий (`:local`) не поддерживается. Собранные стадии пушатся по digest напрямую из buildkitd в указанный repo; werf затем накладывает свой тег стадии на опубликованный манифест на стороне registry.

### Поддерживаемые возможности

BuildKit-бэкенд поддерживает оба режима сборки на паритете с Docker-бэкендом:

*	Stapel-сборки (shell- и ansible-сборщики).
*	Dockerfile-сборки, staged и non-staged.

Паритет включает:

*	проброс ssh-agent в сборку;
*	сборочные секреты;
*	настраиваемую сборочную сеть (`default`, `host`, `none`);
*	пользовательские mounts.

### Семантика host mounts

Host mounts stapel (`fromPath`, `mount: build_dir`) отображаются на persistent cache mounts BuildKit с ключом по host-пути. Данные живут в кэше buildkitd на стороне демона, а не в директории на host-машине werf. Кэш сохраняется между сборками и разделяется по ключу host-пути. Обратите внимание: существующее содержимое host-директории НЕ доставляется в mount — cache mount при первом использовании пуст и накапливает только данные, записанные во время сборок.

### Insecure и self-signed registries

Доступ к insecure registry, пользовательские CA и отключение TLS-верификации настраиваются на стороне демона buildkitd (обычно через `buildkitd.toml`). werf не транслирует свои флаги `--insecure-registry` / `--skip-tls-verify-registry` в buildkitd.

### Очистка хоста

При использовании удалённого buildkitd на host-машине werf нет локального хранилища образов. Команды `werf host purge` и другие команды host-cleanup очищают только служебные директории werf на хосте; сборочный кэш buildkitd в первой итерации werf не прунит — он управляется сборщиком мусора buildkitd (см. `buildkitd.toml`) или вручную через `buildctl prune`.
