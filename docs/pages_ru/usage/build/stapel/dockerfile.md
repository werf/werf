---
title: Использование Docker-инструкций
permalink: usage/build/stapel/dockerfile.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: docker
---

Инструкции [в Dockerfile](https://docs.docker.com/engine/reference/builder/) можно условно разделить на две группы: сборочные инструкции и инструкции, которые влияют на manifest Docker-образа.
Так как werf-сборщик использует свой синтаксис для описания сборки, поддерживаются только следующие Dockerfile-инструкции второй группы:

* `USER` — имя пользователя (или UID) и опционально пользовательская группа (или GID) ([подробнее](https://docs.docker.com/engine/reference/builder/#user));
* `WORKDIR` — рабочая директория ([подробнее](https://docs.docker.com/engine/reference/builder/#workdir));
* `VOLUME` — точка монтирования ([подробнее](https://docs.docker.com/engine/reference/builder/#volume));
* `ENV` — переменные окружения ([подробнее](https://docs.docker.com/engine/reference/builder/#env));
* `LABEL` — метаданные ([подробнее](https://docs.docker.com/engine/reference/builder/#label));
* `EXPOSE` — описание сетевых портов, которые будут прослушиваться в запущенном контейнере ([подробнее](https://docs.docker.com/engine/reference/builder/#expose));
* `ENTRYPOINT` — команда по умолчанию, которая будет выполнена при запуске контейнера ([подробнее](https://docs.docker.com/engine/reference/builder/#entrypoint));
* `CMD` — аргументы по умолчанию для `ENTRYPOINT` ([подробнее](https://docs.docker.com/engine/reference/builder/#cmd));
* `HEALTHCHECK` — инструкции, которые Docker может использовать для проверки работоспособности запущенного контейнера ([подробнее](https://docs.docker.com/engine/reference/builder/#healthcheck)).

Эти инструкции могут быть указаны с помощью директивы `docker` в конфигурации.

Пример:

```yaml
docker:
  WORKDIR: /app
  CMD: ["python", "./index.py"]
  EXPOSE: '5000'
  ENV:
    TERM: xterm
    LC_ALL: en_US.UTF-8
```

Указанные в конфигурации Docker-инструкции применяются на последней стадии конвейера стадий, стадии `docker_instructions`.
Поэтому указание Docker-инструкций в `werf.yaml` никак не влияет на сам процесс сборки, а только добавляет данные к уже собранному образу.

Если вам требуются определённые переменные окружения во время сборки (например, `TERM`), то вам необходимо использовать [базовый образ]({{ "usage/build/stapel/base.html" | true_relative_url }}), в котором эти переменные окружения установлены, или экспортировать их [в _пользовательской стадии_]({{ "usage/build/stapel/instructions.html#пользовательские-стадии" | true_relative_url }}).
