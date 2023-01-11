---
title: Использование Docker-инструкций
permalink: usage/build_draft/stapel/dockerfile.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: docker
---

Инструкции в [Dockerfile](https://docs.docker.com/engine/reference/builder/) можно условно разделить на две группы: сборочные инструкции и инструкции, которые влияют на manifest Docker-образа. 
Так как werf сборщик использует свой синтаксис для описания сборки, поддерживаются только следующие Dockerfile-инструкции второй группы:

* `USER` — имя пользователя (или UID) и опционально пользовательская группа (или GID) ([подробнее](https://docs.docker.com/engine/reference/builder/#user))
* `WORKDIR` — рабочая директория ([подробнее](https://docs.docker.com/engine/reference/builder/#workdir))
* `VOLUME` — точка монтирования ([подробнее](https://docs.docker.com/engine/reference/builder/#volume))
* `ENV` — переменные окружения ([подробнее](https://docs.docker.com/engine/reference/builder/#env))
* `LABEL` — метаданные ([подробнее](https://docs.docker.com/engine/reference/builder/#label))
* `EXPOSE` — описание сетевых портов, которые будут прослушиваться в запущенном контейнере ([подробнее](https://docs.docker.com/engine/reference/builder/#expose))
* `ENTRYPOINT` — команда по умолчанию, которая будет выполнена при запуске контейнера ([подробнее](https://docs.docker.com/engine/reference/builder/#entrypoint))
* `CMD` — аргументы по умолчанию для `ENTRYPOINT` ([подробнее](https://docs.docker.com/engine/reference/builder/#cmd))
* `HEALTHCHECK` — инструкции, которые Docker может использовать для проверки работоспособности запущенного контейнера ([подробнее](https://docs.docker.com/engine/reference/builder/#healthcheck))

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

Если вам требуются определённые переменные окружения во время сборки (например, `TERM`), то вам необходимо использовать [базовый образ]({{ "usage/build_draft/stapel/base.html" | true_relative_url }}), в котором эти переменные окружения установлен или экспортировать их в [_пользовательской стадии_]({{ "usage/build/building_images_with_stapel/assembly_instructions.html#пользовательские-стадии" | true_relative_url }}).

##Как Stapel-сборщик работает с CMD и ENTRYPOINT

Для сборки стадии werf запускает контейнер со служебными значениями `CMD` и `ENTRYPOINT`, а затем заменяет их значениями [базового образа]({{ "advanced/building_images_with_stapel/base_image.html" | true_relative_url }}). Если в базовом образе эти значения не установлены, werf сбрасывает их следующим образом:
- `[]` для `CMD`;
- `[""]` для `ENTRYPOINT`.

Также werf сбрасывает (использует специальные пустые значения) значение `ENTRYPOINT` базового образа, если указано значение `CMD` в конфигурации (`docker.CMD`).

В противном случае поведение werf аналогично [поведению Docker](https://docs.docker.com/engine/reference/builder/#understand-how-cmd-and-entrypoint-interact).
