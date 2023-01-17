---
title: Работа в контейнерах
permalink: usage/build/run_in_containers.html
---

### Готовые образы

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/advanced/ci_cd/run_in_container/how_it_works.html и https://werf.io/documentation/v1.2/advanced/ci_cd/run_in_container/use_docker_container.html -->

Рекомендуемым вариантом запуска werf в контейнерах является использование официальных образов werf из `registry.werf.io/werf/werf`. Запуск werf в контейнере выглядит следующим образом:

```shell
docker run \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    registry.werf.io/werf/werf:1.2-stable WERF_COMMAND
```

Если ядро Linux не предоставляет overlayfs для пользовательского режима без root, то следует использовать fuse-overlayfs:

```shell
docker run \
    --device /dev/fuse \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    registry.werf.io/werf/werf:1.2-stable WERF_COMMAND
```

Рекомендуется использовать образ `registry.werf.io/werf/werf:1.2-stable`, также доступны следующие образы:
- `registry.werf.io/werf/werf:latest` — алиас для `registry.werf.io/werf/werf:1.2-stable`;
- `registry.werf.io/werf/werf:1.2-{alpha|beta|ea|stable|rock-solid}` — все образы основаны на alpine, выбираем только канал стабильности;
- `registry.werf.io/werf/werf:1.2-{alpha|beta|ea|stable|rock-solid}-{alpine|ubuntu|fedora}` — выбираем канал стабильности и базовый дистрибутив.

При использовании этих образов автоматически будет использоваться бэкенд Buildah. Опций для переключения на использование бекэнда Docker Server не предусматривается.

### Работа через Docker Server (не рекомендуется)

> **ЗАМЕЧАНИЕ:** Для этого метода официальных образов с werf не предоставляется.

Для использования локального docker-сервера необходимо примонтировать docker-сокет в контейнер и использовать привилегированный режим:

```shell
docker run \
    --privileged \
    --volume $HOME/.werf:/root/.werf \
    --volume /tmp:/tmp \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    IMAGE WERF_COMMAND
```

— этот метод поддерживает сборку Dockerfile-образов или stapel-образов.

Для работы с docker-сервером по tcp можно использовать следующую команду:

```shell
docker run --env DOCKER_HOST="tcp://HOST:PORT" IMAGE WERF_COMMAND
```

— такой метод поддерживает только сборку Dockerfile-образов. Образы stapel не поддерживаются, поскольку сборщик stapel-образов использует монтирование директорий из хост-системы в сборочные контейнеры.

### Устранение проблем

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/advanced/ci_cd/run_in_container/how_it_works.html#troubleshooting -->

#### `fuse: device not found`

Полностью данная ошибка может выглядеть следующим образом:

```
error mounting new container: error mounting build container "53f916e7a334a4bb0d9dbc38a0901718d40b99765002bb7f2f2e5464b1db4294": error creating overlay mount to /home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/merged, mount_data=",lowerdir=/home/build/.local/share/containers/storage/overlay/l/Z5GEVIFIIQ7H262DYUTX3YOVR6:/home/build/.local/share/containers/storage/overlay/l/PJBBW6UNUNGI37IX6R3LDNPX3J:/home/build/.local/share/containers/storage/overlay/l/MUYSUONLQVE4CJMQVDCH2UBAVQ:/home/build/.local/share/containers/storage/overlay/l/67JHKJDCKBTI4R3Q5S5YG44AD3:/home/build/.local/share/containers/storage/overlay/l/3S72G4SWKDXILGANUOCESP5LDK,upperdir=/home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/diff,workdir=/home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/work,volatile": using mount program /usr/bin/fuse-overlayfs: fuse: device not found, try 'modprobe fuse' first
fuse-overlayfs: cannot mount: No such file or directory
: exit status 1
time="2021-12-06T11:30:20Z" level=error msg="exit status 1"
```

Решение: включите fuse device для контейнера, в котором запущен werf ([подробности](#работа-в-контейнерах)).

#### `flags: 0x1000: permission denied`

Полностью данная ошибка может выглядеть следующим образом:

```
time="2021-12-06T11:23:23Z" level=debug msg="unable to create kernel-style whiteout: operation not permitted"
time="2021-12-06T11:23:23Z" level=debug msg="[graphdriver] trying provided driver \"overlay\""
time="2021-12-06T11:23:23Z" level=debug msg="overlay: mount_program=/usr/bin/fuse-overlayfs"
Running time 0.01 seconds
Error: unable to get buildah client: unable to create new Buildah instance with mode "native-rootless": unable to get storage: mount /home/build/.local/share/containers/storage/overlay:/home/build/.local/share/containers/storage/overlay, flags: 0x1000: permission denied
time="2021-12-06T11:23:23Z" level=error msg="exit status 1"
```

Решение: отключите профили AppArmor и seccomp с помощью параметров `--security-opt seccomp=unconfined` и `--security-opt apparmor=unconfined`, добавьте специальные аннотации в Pod ([подробности](#работа-в-контейнерах)).

#### `unshare(CLONE_NEWUSER): Operation not permitted`

Полностью данная ошибка может выглядеть следующим образом:

```
Error during unshare(CLONE_NEWUSER): Operation not permitted
ERRO[0000] error parsing PID "": strconv.Atoi: parsing "": invalid syntax
ERRO[0000] (unable to determine exit status)
```

Решение: отключите профили AppArmor и seccomp с помощью параметров `--security-opt seccomp=unconfined` и `--security-opt apparmor=unconfined`, добавьте специальные аннотации в Pod или используйте привилегированный контейнер ([подробности](#работа-в-контейнерах)).

#### `User namespaces are not enabled in /proc/sys/kernel/unprivileged_userns_clone`

Решение:
```bash
# Включить непривилегированные user namespaces:
echo 'kernel.unprivileged_userns_clone = 1' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```
