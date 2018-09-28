---
title: Adding docker-instructions
sidebar: reference
permalink: reference/build/docker_directive.html
---

## Проблематика

TODO Отсылка к Dokerfile

Если оглянуться на главу [Первое приложение на dapp]({{ site.baseurl }}/how_to/build_run_and_push.html), то встанет вопрос: "На каких стадиях выполняются from (аналог docker-инструкции FROM) и docker.RUN"? На самом деле для этих директив выделены отдельные, внутренние, стадии. from выполняется на одноименной стадии в самом начале сборки — базовый образ меняется очень редко в процессе разработки, поэтому стадия from выполняется перед before_install. Для всех остальных директив docker.* выделена стадия docker_instructions, которая выполняется после setup.

## Syntax

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
