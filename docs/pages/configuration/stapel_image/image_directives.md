---
title: All directives
sidebar: documentation
permalink: documentation/configuration/stapel_image/image_directives.html
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
- add: <absolute path in git repository>
  to: <absolute path inside image>
  owner: <owner>
  group: <group>
  includePaths:
  - <path or glob relative to path in add>
  excludePaths:
  - <path or glob relative to path in add>
  stageDependencies:
    install:
    - <path or glob relative to path in add>
    beforeSetup:
    - <path or glob relative to path in add>
    setup:
    - <path or glob relative to path in add>
# remote git
- url: <git repo url>
  branch: <branch name>
  commit: <commit>
  tag: <tag>
  add: <absolute path in git repository>
  to: <absolute path inside image>
  owner: <owner>
  group: <group>
  includePaths:
  - <path or glob relative to path in add>
  excludePaths:
  - <path or glob relative to path in add>
  stageDependencies:
    install:
    - <path or glob relative to path in add>
    beforeSetup:
    - <path or glob relative to path in add>
    setup:
    - <path or glob relative to path in add>
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
- fromPath: <absolute or relative path>
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
  ENTRYPOINT: <entrypoint>
  CMD: <cmd>
  WORKDIR: <workdir>
  USER: <user>
  HEALTHCHECK: <healthcheck>
asLayers: <bool>
```
