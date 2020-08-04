---
title: Подключение зависимостей
sidebar: applications_guide
guide_code: gitlab_nodejs
permalink: documentation/guides/applications_guide/gitlab_nodejs/030_dependencies.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- werf.yaml
{% endfilesused %}

В этой главе мы настроим в нашем базовом приложении работу с зависимостями. Важно корректно вписать зависимости в [стадии сборки](https://ru.werf.io/documentation/reference/stages_and_images.html), что позволит не тратить время на пересборку зависимостей тогда, когда зависимости не изменились.

{% offtopic title="Что за стадии?" %}
Werf подразумевает, что лучшей практикой будет разделить сборочный процесс на этапы, каждый с четкими функциями и своим назначением. Каждый такой этап соответствует промежуточному образу, подобно слоям в Docker. В werf такой этап называется стадией, и конечный образ в итоге состоит из набора собранных стадий. Все стадии хранятся в хранилище стадий, которое можно рассматривать как кэш сборки приложения, хотя по сути это скорее часть контекста сборки.

Стадии — это этапы сборочного процесса, кирпичи, из которых в итоге собирается конечный образ. Стадия собирается из группы сборочных инструкций, указанных в конфигурации. Причем группировка этих инструкций не случайна, имеет определенную логику и учитывает условия и правила сборки. С каждой стадией связан конкретный Docker-образ. Подробнее о том, какие стадии для чего предполагаются можно посмотреть в [документации](https://ru.werf.io/documentation/reference/stages_and_images.html).

Werf предлагает использовать для стадий следующую стратегию:

*   использовать стадию beforeInstall для инсталляции системных пакетов;
*   использовать стадию install для инсталляции системных зависимостей и зависимостей приложения;
*   использовать стадию beforeSetup для настройки системных параметров и установки приложения;
*   использовать стадию setup для настройки приложения.

Подробно про стадии описано в [документации](https://ru.werf.io/documentation/configuration/stapel_image/assembly_instructions.html).

Одно из основных преимуществ использования стадий в том, что мы можем не перезапускать нашу сборку с нуля, а перезапускать её только с той стадии, которая зависит от изменений в определенных файлах.
{% endofftopic %}

В nodejs в качестве менеджера зависимостей используется npm. Пропишем его использование в файле `werf.yaml` и затем оптимизируем его использование.

## Подключение менеджера зависимостей

Пропишем команду `npm ci` в нужные стадии сборки в `werf.yaml`

{% snippetcut name="werf.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/werf.yaml" %}
{% raw %}
```yaml
shell:
  install:
  - cd /app && npm сi
```
{% endraw %}
{% endsnippetcut %}

Однако, если оставить всё так — стадия `install` не будет запускаться при изменении lock-файла `package.json`. Подобная зависимость пользовательской стадии от изменений [указывается с помощью параметра git.stageDependencies](https://ru.werf.io/documentation/configuration/stapel_image/assembly_instructions.html#%D0%B7%D0%B0%D0%B2%D0%B8%D1%81%D0%B8%D0%BC%D0%BE%D1%81%D1%82%D1%8C-%D0%BE%D1%82-%D0%B8%D0%B7%D0%BC%D0%B5%D0%BD%D0%B5%D0%BD%D0%B8%D0%B9-%D0%B2-git-%D1%80%D0%B5%D0%BF%D0%BE%D0%B7%D0%B8%D1%82%D0%BE%D1%80%D0%B8%D0%B8):

{% snippetcut name="werf.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/030-deps/werf.yaml" %}
{% raw %}
```yaml
git:
- add: /
  to: /app
  stageDependencies:
    install:
    - package.json
```
{% endraw %}
{% endsnippetcut %}

При изменении файла `package.json` стадия `install` будет запущена заново.

## Оптимизация сборки

Чтобы не хранить не нужные кэши пакетного менеджера в образе, можно при сборке смонтировать директорию, в которой будет хранится наш кэш.

Для того, чтобы оптимизировать работу с этим кешом при сборке, мы добавим специальную конструкцию в `werf.yaml`:

{% snippetcut name="werf.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/030-deps/werf.yaml" %}
{% raw %}
```yaml
mount:
- from: build_dir
  to: /var/cache/apt
```
{% endraw %}
{% endsnippetcut %}

При каждом запуске билда, эта директория будет монтироваться с сервера, где запускается билд, и не будет очищаться между билдами. Таким образом кэш будет сохраняться между сборками.

<div>
    <a href="040_assets.html" class="nav-btn">Далее: Генерируем и раздаем ассеты</a>
</div>
