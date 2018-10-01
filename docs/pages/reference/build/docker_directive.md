---
title: Adding docker instructions
sidebar: reference
permalink: reference/build/docker_directive.html
---

```yaml
docker:
  VOLUME:
  - <volume>
  EXPOSE:
  - <expose>
  ENV:
    <env_name>: <env_value>
  LABEL:
    <label_name>: <label_value>
  ENTRYPOINT:
  - <entrypoint>
  CMD:
  - <cmd>
  ONBUILD:
  - <onbuild>
  WORKDIR: <workdir>
  USER: <user>
```
