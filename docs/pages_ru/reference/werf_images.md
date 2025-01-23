---
title: Образы werf
permalink: reference/werf_images.html
---

[Релизный процесс]({{ site.url }}/about/release_channels.html) werf включает публикацию образов с werf, необходимыми утилитами и предустановленными настройками для сборки со сборочным бэкендом Buildah.

> Примеры использования образов werf можно посмотреть в разделе [Быстрый старт]({{ site.url }}/getting_started/).

Поддерживаются образы, публикуемые по следующему принципу:

- `registry.werf.io/werf/werf:<group>` (например, `registry.werf.io/werf/werf:2`);
- `registry.werf.io/werf/werf:<group>-<channel>` (например, `registry.werf.io/werf/werf:2-stable`);
- `registry.werf.io/werf/werf:<group>-<channel>-<os>` (например, `registry.werf.io/werf/werf:2-stable-alpine`);
- `registry.werf.io/werf/werf:<version>` (например, `registry.werf.io/werf/werf:2.16.2`);
- `registry.werf.io/werf/werf:<version>-<os>` (например, `registry.werf.io/werf/werf:2.16.2-alpine`).

Где:

- `<group>`: группа `1.2` или `2` (по умолчанию);
- `<channel>`: канал выпуска `alpha`, `beta`, `ea`, `stable` (по умолчанию) или `rock-solid`;
- `<os>`: операционная система `alpine` (по умолчанию), `ubuntu` или `fedora`.
- `<version>`: версия (например, `2.16.2`). Если версия релиза содержит `+fix` (например, `2.16.2+fix1`), то версия будет приведена к `2.16.2.fix1`.
