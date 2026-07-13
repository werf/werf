---
title: Сборочные бэкенды
permalink: usage/build/backends.html
---

## Обзор

werf собирает образы через демон [buildkitd](https://github.com/moby/buildkit). Эндпоинт выбирается следующим образом:

-	Задана переменная `WERF_BUILDKIT_HOST` (или стандартная `BUILDKIT_HOST`) — werf использует указанный эндпоинт buildkitd.
-	Ни одна из переменных не задана — werf автоматически запускает (или переиспользует) локальный контейнер buildkitd с именем `werf-buildkitd` на локальном Docker-демоне и работает через `docker-container://werf-buildkitd`. В этом случае требуется доступный Docker.

> Необходимые требования и подготовка системы для работы со сборочными бэкендами описаны в разделе сайта [Быстрый старт]({{ site.url }}/getting_started/).

## BuildKit

### Эндпоинты

Поддерживаются следующие схемы эндпоинтов:

*	`unix://` — локальный Unix-сокет;
*	`tcp://` — TCP-эндпоинт;
*	`docker-container://` — buildkitd, запущенный в Docker-контейнере;
*	`kube-pod://` — buildkitd, запущенный в Kubernetes-поде;
*	`podman-container://` — buildkitd, запущенный в Podman-контейнере;
*	`ssh://` — buildkitd, доступный по SSH.

### Быстрый старт

Если локально доступен Docker, настройка не требуется: при первой сборке werf автоматически запустит контейнер `werf-buildkitd`.

Чтобы использовать внешний buildkitd:

```shell
export BUILDKIT_HOST=tcp://my-buildkitd:1234
```

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

### Локальный registry

Адрес registry должен быть доступен и с хоста werf, и изнутри контейнера buildkitd:

*	На нативном Linux Docker-демоне werf-контейнер buildkitd использует host-сеть, поэтому registry на `localhost:<port>` работает как есть.
*	На Docker Desktop (macOS/Windows) вместо `localhost` используйте LAN IP хоста: `werf build --repo <host-ip>:5000/myproject --insecure-registry --skip-tls-verify-registry`.

С флагами `--insecure-registry` / `--skip-tls-verify-registry` werf автоматически настраивает werf-контейнер buildkitd для работы с plain-HTTP или self-signed registry.

### Insecure и self-signed registries

Доступ к insecure registry, пользовательские CA и отключение TLS-верификации настраиваются на стороне демона buildkitd (обычно через `buildkitd.toml`). werf не транслирует свои флаги `--insecure-registry` / `--skip-tls-verify-registry` в buildkitd.

### Очистка хоста

При использовании удалённого buildkitd на host-машине werf нет локального хранилища образов. Команды `werf host purge` и другие команды host-cleanup очищают только служебные директории werf на хосте; сборочный кэш buildkitd в первой итерации werf не прунит — он управляется сборщиком мусора buildkitd (см. `buildkitd.toml`) или вручную через `buildctl prune`.
