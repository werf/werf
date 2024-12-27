---
title: Доступные образа werf
permalink: reference/werf_images.html
---

[Релизный процесс]({{ site.url }}/about/release_channels.html) werf включает публикацию образов с werf, необходимыми утилитами и предустановленными настройками для сборки со сборочным бэкендом Buildah. 

Поддерживаются образы, публикуемые по следующему принципу:

* `registry.werf.io/werf/werf:latest`,
* `registry.werf.io/werf/werf:<group>` (например, `registry.werf.io/werf/werf:2`);
* `registry.werf.io/werf/werf:<group>-<channel>` (например, `registry.werf.io/werf/werf:2-stable`);
* `registry.werf.io/werf/werf:<group>-<channel>-<os>` (например, `registry.werf.io/werf/werf:2-stable-alpine`);

Где:

* `<group>`: группа `1.2` или `2` (по умолчанию);
* `<channel>`: канал выпуска `alpha`, `beta`, `ea`, `stable` (по умолчанию) или `rock-solid`;
* `<os>`: операционная система `alpine` (по умолчанию), `ubuntu` или `fedora`.
