---
title: Build process
sidebar: documentation
permalink: documentation/reference/build_process.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

## Build command

{% include /cli/werf_build.md %}

## Multiple builds on the same host

Multiple build commands can run at the same time on the same host. When building _stage_ Werf acquires a **lock** using _stage signature_ as ID so that only one build process is active for a stage with a particular signature at the same time.

When another build process is holding a lock for a stage, Werf waits until this process releases a lock. Then Werf proceeds to the next stage.

The reason is no need to build the same stage multiple times. Werf build process can wait until another process finishes build and puts _stage_ into the _stages storage_.
