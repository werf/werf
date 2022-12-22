---
title: Overview
permalink: usage/build_draft/stapel/overview.html
---

werf has a built-in alternative syntax for describing assembly instructions called Stapel. Here are its distinctive features:

1. Easily support and parameterize complex configurations, reuse common snippets and generate configurations of the images of the same type using YAML format and templating.
2. Dedicated commands for integrating with git to enable incremental rebuilds based on the git repository history.
3. Inheriting images and importing files from images (similar to Dockerfile's multi-stage).
4. Run arbitrary build instructions, specify directory mount options, and use other advanced tools to build images.
5. More efficient caching mechanics for layers (a similar scheme is supported for Dockerfile layers when building with Buildah (currently pre-alpha)).

<!-- TODO(staged-dockerfile): удалить 5 пункт как неактуальный -->

To build images using the Stapel builder, you need to describe the build instructions in the `werf.yaml` configuration file. Stapel is supported for both the docker server builder backend (building via shell instructions or ansible) and for buildah (shell instructions only).

This section describes how to build images with the Stapel builder, its advanced features and how to use them.
