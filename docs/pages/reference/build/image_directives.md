---
title: All directives
sidebar: reference
permalink: reference/build/image_directives.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

```yaml
image: <image name... || ~>
from: <image[:<tag>]>
fromLatest: <bool>
fromCacheVersion: <arbitrary string>
fromImage: <image name>
fromImageArtifact: <artifact name>
git:
# local git
- add: <absolute path>
  to: <absolute path>
  owner: <owner>
  group: <group>
  includePaths:
  - <relative path or glob>
  excludePaths:
  - <relative path or glob>
  stageDependencies:
    install:
    - <relative path or glob>
    beforeSetup:
    - <relative path or glob>
    setup:
    - <relative path or glob>
# remote git
- url: <git repo url>
  branch: <branch name>
  commit: <commit>
  tag: <tag>
  add: <absolute path>
  to: <absolute path>
  owner: <owner>
  group: <group>
  includePaths:
  - <relative path or glob>
  excludePaths:
  - <relative path or glob>
  stageDependencies:
    install:
    - <relative path or glob>
    beforeSetup:
    - <relative path or glob>
    setup:
    - <relative path or glob>
shell:
  beforeInstall:
  - <bash command>
  install:
  - <bash command>
  beforeSetup:
  - <bash command>
  setup:
  - <bash command>
  cacheVersion: <arbitrary string>
  beforeInstallCacheVersion: <arbitrary string>
  installCacheVersion: <arbitrary string>
  beforeSetupCacheVersion: <arbitrary string>
  setupCacheVersion: <arbitrary string>
ansible:
  beforeInstall:
  - <task>
  install:
  - <task>
  beforeSetup:
  - <task>
  setup:
  - <task>
  cacheVersion: <arbitrary string>
  beforeInstallCacheVersion: <arbitrary string>
  installCacheVersion: <arbitrary string>
  beforeSetupCacheVersion: <arbitrary string>
  setupCacheVersion: <arbitrary string>
mount:
- from: build_dir
  to: <absolute path>
- from: tmp_dir
  to: <absolute path>
- fromPath: <absolute path>
  to: <absolute path>
import:
- artifact: <artifact name>
  before: <install || setup>
  after: <install || setup>
  add: <absolute path>
  to: <absolute path>
  owner: <owner>
  group: <group>
  includePaths:
  - <relative path or glob>
  excludePaths:
  - <relative path or glob>
docker:
  VOLUME:
  - <volume>
  EXPOSE:
  - <expose>
  ENV:
    <env name>: <env value>
  LABEL:
    <label name>: <label value>
  ENTRYPOINT:
  - <entrypoint>
  CMD:
  - <cmd>
  ONBUILD:
  - <onbuild>
  WORKDIR: <workdir>
  USER: <user>
  STOPSIGNAL: <stopsignal>
  HEALTHCHECK: <healthcheck>
asLayers: <bool>
```
