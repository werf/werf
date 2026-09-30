---
title: werf images
permalink: reference/werf_images.html
---

The [release process]({{ site.url }}/about/release_channels.html) for werf includes the publication of images with werf, necessary utilities, and pre-configured settings for building with the Buildah backend.

> You can find examples of using werf images in the [Getting Started]({{ site.url }}/getting_started/).

The images follow the naming convention:

- `registry.werf.io/werf/werf:<group>` (e.g., `registry.werf.io/werf/werf:3`);
- `registry.werf.io/werf/werf:<group>-<channel>` (e.g., `registry.werf.io/werf/werf:3-stable`);
- `registry.werf.io/werf/werf:<group>-<channel>-<os>` (e.g., `registry.werf.io/werf/werf:3-stable-alpine`);
- `registry.werf.io/werf/werf:<version>` (e.g., `registry.werf.io/werf/werf:3.6.1`);
- `registry.werf.io/werf/werf:<version>-<os>` (e.g., `registry.werf.io/werf/werf:3.6.1-alpine`).

Where:

- `<group>`: version group; use `3`;
- `<channel>`: release channel, such as `alpha`, `beta`, `ea`, `stable` (default), or `rock-solid`;
- `<os>`: operating system, such as `alpine` (default), `ubuntu`, or `fedora`;
- `<version>`: version (e.g., `3.6.1`). If the release version includes `+fix` (e.g., `3.6.1+fix1`), it will be converted to `3.6.1.fix1` version.
