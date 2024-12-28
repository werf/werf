---
title: werf images
permalink: reference/werf_images.html
---

The [release process]({{ site.url }}/about/release_channels.html) for werf includes the publication of images with werf, necessary utilities, and pre-configured settings for building with the Buildah backend.

> You can find examples of using werf images in the [Getting Started]({{ site.url }}/getting_started/).

The images follow the naming convention:

- `registry.werf.io/werf/werf:latest`.
- `registry.werf.io/werf/werf:<group>` (e.g., `registry.werf.io/werf/werf:2`);
- `registry.werf.io/werf/werf:<group>-<channel>` (e.g., `registry.werf.io/werf/werf:2-stable`);
- `registry.werf.io/werf/werf:<group>-<channel>-<os>` (e.g., `registry.werf.io/werf/werf:2-stable-alpine`);

Where:

- `<group>`: version group, such as `1.2` or `2` (default);
- `<channel>`: release channel, such as `alpha`, `beta`, `ea`, `stable` (default), or `rock-solid`;
- `<os>`: operating system, such as `alpine` (default), `ubuntu`, or `fedora`.
