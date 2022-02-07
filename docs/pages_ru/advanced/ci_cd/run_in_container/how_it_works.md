---
title: Принципы работы
permalink: advanced/ci_cd/run_in_container/how_it_works.html
---

ПРИМЕЧАНИЕ: в настоящее время werf поддерживает сборку образов с _использованием Docker-сервера_ или _без его использования_ (в экспериментальном режиме). Информация на данной странице применима только к экспериментальному режиму _без использования Docker-сервера_. Пока для этого режима доступен только сборщик образов на основе Dockerfile'ов. Сборщик Stapel будет доступен через некоторое время.

Общая информация о том, как включить Buildah в werf, доступна на странице [режима сборки с использованием Buildah]({{ "/advanced/buildah_mode.html" | true_relative_url }}).

## Режимы работы

В зависимости от соответствия системы пользователя [требованиям режима Buildah]({{ "/advanced/buildah_mode.html#системные-требования" | true_relative_url }}) и в зависимости от потребностей пользователя существует 3 способа использования werf с Buildah внутри контейнеров:

### Ядро Linux с поддержкой OverlayFS в режиме rootless

В данном случае необходимо только **отключить** профили **seccomp** и **AppArmor** в контейнере с werf.

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование привилегированного контейнера

В случае, если ваше ядро Linux не поддерживает [OverlayFS в режиме rootless]({{ "/advanced/buildah_mode.html#системные-требования" | true_relative_url }}), Buildah и werf вместо нее будут использовать fuse-overlayfs.

Чтобы использовать fuse-overlayfs, можно воспользоваться **привилегированным контейнером** для запуска werf.

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование непривилегированного контейнера

В случае, если ваше ядро Linux не поддерживает [OverlayFS в режиме rootless]({{ "/advanced/buildah_mode.html#системные-требования" | true_relative_url }}), Buildah и werf вместо нее будут использовать fuse-overlayfs.

Чтобы воспользоваться fuse-overlayfs без привилегированного контейнера, запустите werf в контейнере со следующими параметрами:

1. **Отключены** профили **seccomp** и **AppArmor**.
2. Включено устройство `/dev/fuse`.

## Доступные образы werf

Ниже приведен список образов со встроенной утилитой werf. Каждый образ обновляется в рамках релизного процесса, основанного на менеджере пакетов trdl ([подробнее о каналах обновлений]({{ "installation.html#все-изменения-в-werf-проходят-через-цепочку-каналов-обновлений" | true_relative_url }})).

* `ghcr.io/werf/werf:latest` -> `ghcr.io/werf/werf:1.2-stable`;
* `ghcr.io/werf/werf:1.2-alpha` -> `ghcr.io/werf/werf:1.2-alpha-alpine`;
* `ghcr.io/werf/werf:1.2-beta` -> `ghcr.io/werf/werf:1.2-beta-alpine`;
* `ghcr.io/werf/werf:1.2-ea` -> `ghcr.io/werf/werf:1.2-ea-alpine`;
* `ghcr.io/werf/werf:1.2-stable` -> `ghcr.io/werf/werf:1.2-stable-alpine`;
* `ghcr.io/werf/werf:1.2-alpha-alpine`;
* `ghcr.io/werf/werf:1.2-beta-alpine`;
* `ghcr.io/werf/werf:1.2-ea-alpine`;
* `ghcr.io/werf/werf:1.2-stable-alpine`;
* `ghcr.io/werf/werf:1.2-alpha-ubuntu`;
* `ghcr.io/werf/werf:1.2-beta-ubuntu`;
* `ghcr.io/werf/werf:1.2-ea-ubuntu`;
* `ghcr.io/werf/werf:1.2-stable-ubuntu`;
* `ghcr.io/werf/werf:1.2-alpha-centos`;
* `ghcr.io/werf/werf:1.2-beta-centos`;
* `ghcr.io/werf/werf:1.2-ea-centos`;
* `ghcr.io/werf/werf:1.2-stable-centos`;
* `ghcr.io/werf/werf:1.2-alpha-fedora`;
* `ghcr.io/werf/werf:1.2-beta-fedora`;
* `ghcr.io/werf/werf:1.2-ea-fedora`;
* `ghcr.io/werf/werf:1.2-stable-fedora`.

## Устранение проблем

### `fuse: device not found`

Полностью данная ошибка может выглядеть следующим образом:

```
error mounting new container: error mounting build container "53f916e7a334a4bb0d9dbc38a0901718d40b99765002bb7f2f2e5464b1db4294": error creating overlay mount to /home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/merged, mount_data=",lowerdir=/home/build/.local/share/containers/storage/overlay/l/Z5GEVIFIIQ7H262DYUTX3YOVR6:/home/build/.local/share/containers/storage/overlay/l/PJBBW6UNUNGI37IX6R3LDNPX3J:/home/build/.local/share/containers/storage/overlay/l/MUYSUONLQVE4CJMQVDCH2UBAVQ:/home/build/.local/share/containers/storage/overlay/l/67JHKJDCKBTI4R3Q5S5YG44AD3:/home/build/.local/share/containers/storage/overlay/l/3S72G4SWKDXILGANUOCESP5LDK,upperdir=/home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/diff,workdir=/home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/work,volatile": using mount program /usr/bin/fuse-overlayfs: fuse: device not found, try 'modprobe fuse' first
fuse-overlayfs: cannot mount: No such file or directory
: exit status 1
time="2021-12-06T11:30:20Z" level=error msg="exit status 1"
```

Решение: включите fuse device для контейнера, в котором запущен werf ([подробности](#ядро-linux-без-поддержки-overlayfs-в-режиме-rootless-и-использование-непривилегированного-контейнера)).

### `flags: 0x1000: permission denied`

Полностью данная ошибка может выглядеть следующим образом:

```
time="2021-12-06T11:23:23Z" level=debug msg="unable to create kernel-style whiteout: operation not permitted"
time="2021-12-06T11:23:23Z" level=debug msg="[graphdriver] trying provided driver \"overlay\""
time="2021-12-06T11:23:23Z" level=debug msg="overlay: mount_program=/usr/bin/fuse-overlayfs"
Running time 0.01 seconds
Error: unable to get buildah client: unable to create new Buildah instance with mode "native-rootless": unable to get storage: mount /home/build/.local/share/containers/storage/overlay:/home/build/.local/share/containers/storage/overlay, flags: 0x1000: permission denied
time="2021-12-06T11:23:23Z" level=error msg="exit status 1"
```

Решение: отключите профили AppArmor и seccomp с помощью параметров `--security-opt seccomp=unconfined` и `--security-opt apparmor=unconfined`, добавьте специальные аннотации в Pod ([подробности](#ядро-linux-без-поддержки-overlayfs-в-режиме-rootless-и-использование-непривилегированного-контейнера)).

### `unshare(CLONE_NEWUSER): Operation not permitted`

Полностью данная ошибка может выглядеть следующим образом:

```
Error during unshare(CLONE_NEWUSER): Operation not permitted
ERRO[0000] error parsing PID "": strconv.Atoi: parsing "": invalid syntax 
ERRO[0000] (unable to determine exit status)            
```

Решение: отключите профили AppArmor и seccomp с помощью параметров `--security-opt seccomp=unconfined` и `--security-opt apparmor=unconfined`, добавьте специальные аннотации в Pod или используйте привилегированный контейнер ([подробности](#режимы-работы)).

### `User namespaces are not enabled in /proc/sys/kernel/unprivileged_userns_clone`

Решение:
```bash
# Включить непривилегированные user namespaces:
echo 'kernel.unprivileged_userns_clone = 1' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```
