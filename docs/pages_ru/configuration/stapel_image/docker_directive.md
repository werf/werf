---
title: Добавление Docker-инструкций
sidebar: documentation
permalink: documentation/configuration/stapel_image/docker_directive.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
summary: |
  <a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vTZB0BLxL7mRUFxkrOMaj310CQgb5D5H_V0gXe7QYsTu3kKkdwchg--A1EoEP2CtKbO8pp2qARfeoOK/pub?w=2031&amp;h=144" data-featherlight="image">
    <img src="https://docs.google.com/drawings/d/e/2PACX-1vTZB0BLxL7mRUFxkrOMaj310CQgb5D5H_V0gXe7QYsTu3kKkdwchg--A1EoEP2CtKbO8pp2qARfeoOK/pub?w=1016&amp;h=72">
  </a>

    <div class="language-yaml highlighter-rouge"><div class="highlight"><pre class="highlight"><code><span class="na">docker</span><span class="pi">:</span>
    <span class="na">VOLUME</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;volume&gt;</span>
    <span class="na">EXPOSE</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;expose&gt;</span>
    <span class="na">ENV</span><span class="pi">:</span>
      <span class="s">&lt;env_name&gt;</span><span class="pi">:</span> <span class="s">&lt;env_value&gt;</span>
    <span class="na">LABEL</span><span class="pi">:</span>
      <span class="s">&lt;label_name&gt;</span><span class="pi">:</span> <span class="s">&lt;label_value&gt;</span>
    <span class="na">ENTRYPOINT</span><span class="pi">:</span> <span class="s">&lt;entrypoint&gt;</span>
    <span class="na">CMD</span><span class="pi">:</span> <span class="s">&lt;cmd&gt;</span>
    <span class="na">WORKDIR</span><span class="pi">:</span> <span class="s">&lt;workdir&gt;</span>
    <span class="na">USER</span><span class="pi">:</span> <span class="s">&lt;user&gt;</span>
    <span class="na">HEALTHCHECK</span><span class="pi">:</span> <span class="s">&lt;healthcheck&gt;</span></code></pre></div></div>
---

Инструкции в [Dockerfile](https://docs.docker.com/engine/reference/builder/) можно условно разделить на две группы: сборочные инструкции и инструкции, которые влияют на manifest Docker-образа. 
Так как Werf сборщик использует свой синтаксис для описания сборки, поддерживаются только следующие Dockerfile-инструкции второй группы:

* `USER` — пользователь и группа, которые необходимо использовать при запуске контейнера ([подробнее](https://docs.docker.com/engine/reference/builder/#user))
* `WORKDIR` — рабочая директория, используемая при запуске контейнера ([подробнее](https://docs.docker.com/engine/reference/builder/#workdir))
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
  CMD: ['python', './index.py']
  EXPOSE: '5000'
  ENV:
    TERM: xterm
    LC_ALL: en_US.UTF-8
```

Указанные в конфигурации Docker-инструкции применяются на последней стадии конвейера стадий, стадии `docker_instructions`. 
Поэтому указание Docker-инструкций в `werf.yaml` никак не влияет на сам процесс сборки, а только добавляет данные к уже собранному образу.

Если вам требуются определённые переменные окружения во время сборки (например, `TERM`), то вам необходимо использовать [базовый образ]({{ site.baseurl }}/documentation/configuration/stapel_image/base_image.html), в котором эти переменные окружения установлен или экспортировать их в [_пользовательской стадии_]({{ site.baseurl }}/documentation/configuration/stapel_image/assembly_instructions.html#what-is-user-stages).
