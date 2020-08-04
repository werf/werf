---
title: Сборка
sidebar: applications_guide
guide_code: gitlab_nodejs
permalink: documentation/guides/applications_guide/gitlab_nodejs/020_basic/10_build.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- werf.yaml
{% endfilesused %}

В этой главе мы научимся писать конфигурацию сборки, отлаживать её в случае ошибок и загружать полученный образ в Registry.

{% offtopic title="Локальная разработка или на сервере?" %}
Существует 2 варианта ведения разработки: локально на компьютере либо разработка на сервере. У обоих вариантов есть свои плюсы и минусы.

Локальная разработка подойдет вам если у вас достаточно мощный компьютер и для вас важна скорость разработки. В остальных случаях - лучше выбирать разработку на сервере.

Разберём плюсы и минусы каждого варианта подробнее:

**Локальная разработка**

Основной плюс этого метода - скорость разработки и проверки изменений. В зависимости от используемого языка программирования, результат правок можно увидеть уже через несколько секунд.

К минусам можно отнести необходимость в более мощном железе, чтобы мы могли запускать приложение, а также то, что локальное окружение будет значительно отличаться от production.

_В скором времени werf предоставит разработчикам запускать локальное окружение, идентичное production._

**Разработка на сервере**

Главный плюс этого метода - однотипность production с остальными окружениями. При возникновении проблем со сборкой мы узнаем об этом моментально. Отбрасывается необходимость в дополнительных ресурсах для локального запуска.

Один из минусов это отзывчивость. На процесс от пуша кода до появления результата может потребоваться несколько минут.
{% endofftopic %}

Возьмём исходный код приложения из git:

```bash
git clone git@github.com:werf/demos.git
```

Примеры из данного гайда находятся в папке `/applications-guide/gitlab-nodejs`, начнём с примера `020-basic`.

Для того чтобы werf смог собрать docker-образ с приложением - необходимо в корне нашего репозитория создать файл `werf.yaml` в которым будут описаны инструкции по сборке.

{% offtopic title="Варианты синтаксиса werf.yaml" %}

Существует 2 варианта синтаксиса при использовании werf: Ansible и Shell.


**Ansible**

Werf поддерживает почти все модули из ansible, поэтому если у вас имеются ansible плейбуки для сборки, их можно будет очень легко адаптировать под werf.

Пример установки curl:
{% raw %}
```yaml
- name: "Install additional packages"
apt:
    state: present
    update_cache: yes
    pkg:
    - curl
```
{% endraw %}
Полный список поддерживаемых модулей ansible в werf можно найти [тут](https://werf.io/documentation/configuration/stapel_image/assembly_instructions.html#supported-modules).


**Shell**

Werf поддерживает использование обычных shell команд, будто мы запускаем обычный bash скрипт.

Пример установки curl:
{% raw %}
```yaml
shell:
  beforeInstall:
  - apt-get update
  - apt-get install -y curl
```
{% endraw %}
Для простоты и удобства, мы будем рассматривать Shell, но для своего проекта, вы можете выбрать любой из предложенных вариантов.

Прочитать подробнее про виды синтаксисов вы можете тут:
[**syntax section**](https://ru.werf.io/documentation/guides/advanced_build/first_application.html#%D1%88%D0%B0%D0%B3-1-%D0%BA%D0%BE%D0%BD%D1%84%D0%B8%D0%B3%D1%83%D1%80%D0%B0%D1%86%D0%B8%D1%8F-werfyaml).


{% endofftopic %}

Начнём werf.yaml с обязательной [**секции мета-информации**]({{ site.baseurl }}/documentation/configuration/introduction.html#секция-мета-информации):

{% snippetcut name="werf.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/werf.yaml" %}
```yaml
project: werf-guided-project
configVersion: 1
```
{% endsnippetcut %}

**_project_** - поле, задающее имя для проекта, которым мы определяем связь всех docker images собираемых в данном проекте. Данное имя по умолчанию используется в имени helm релиза и имени namespace в которое будет выкатываться наше приложение. Данное имя не рекомендуется изменять (или подходить к таким изменениям с должным уровнем ответственности) так как после изменений уже имеющиеся ресурсы, которые выкачены кластер, не будут переименованы.

**_configVersion_** - в данном случае определяет версию синтаксиса используемую в `werf.yaml`.(На данный момент мы всегда используем 1)


После мы сразу переходим к следующей секции конфигурации, которая и будет для нас основной секцией для сборки - [**image config section**](https://ru.werf.io/v1.1-alpha/documentation/configuration/introduction.html#%D1%81%D0%B5%D0%BA%D1%86%D0%B8%D1%8F-%D0%BE%D0%B1%D1%80%D0%B0%D0%B7%D0%B0).

{% snippetcut name="werf.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/werf.yaml" %}
```yaml
project: werf-guided-project
configVersion: 1
---
image: basicapp
from: node:14-stretch
```
{% endsnippetcut %}

В строке `image: basicapp` дано название для образа, который соберёт werf. Это имя мы впоследствии будем указывать при запуске контейнера. Строка `from: node:14-stretch` определяет, что будет взято за основу, мы берем официальный публичный образ с нужной нам версией ruby.

{% offtopic title="Что делать, если образов и других констант станет много" %}

Werf поддерживает [**Go templates**](https://ru.werf.io/v1.1-alpha/documentation/configuration/introduction.html#%D1%88%D0%B0%D0%B1%D0%BB%D0%BE%D0%BD%D1%8B-go), так что вы с легкостью можете заводить переменные и выписывать в них константы и часто использующиеся образы.

Сделаем 2 образа, используя один базовый образ `golang:1.11-alpine`

Пример:
{% raw %}

```yaml
{{ $base_image := "golang:1.11-alpine" }}

project: my-project
configVersion: 1
---

image: gogomonia
from: {{ $base_image }}
---
image: saml-authenticator
from: {{ $base_image }}
```
{% endraw %}

Подробнее почитать про Go шаблоны в werf можно тут: [**werf go templates**](https://ru.werf.io/v1.1-alpha/documentation/configuration/introduction.html#%D1%88%D0%B0%D0%B1%D0%BB%D0%BE%D0%BD%D1%8B-go).

{% endofftopic %}

Добавим исходный код нашего приложения в контейнер с помощью [**директивы git**](https://ru.werf.io/v1.1-alpha/documentation/configuration/stapel_image/git_directive.html)

{% snippetcut name="werf.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/werf.yaml" %}
{% raw %}
```yaml
project: werf-guided-project
configVersion: 1
---
image: basicapp
from: node:14-stretch
git:
- add: /
  to: /app
```
{% endraw %}
{% endsnippetcut %}

При использовании директивы `git`, werf использует репозиторий проекта, в котором расположен `werf.yaml`. Добавляемые директории и файлы проекта указываются списком относительно корня репозитория.

`add: /` - та директория которую мы хотим добавить внутрь docker image, мы указываем, что это корень, т.е. мы хотим склонировать внутрь docker image весь наш репозиторий.

`to: /app` - то куда мы клонируем наш репозиторий внутри docker image. Важно заметить что директорию назначения werf создаст сам.

Как такового репозитория в образе не создаётся, директория .git отсутствует. На одном из первых этапов сборки все исходники пакуются в архив (с учётом всех пользовательских параметров) и распаковываются внутри контейнера. При последующих сборках [накладываются патчи с разницей](https://ru.werf.io/v1.1-alpha/documentation/configuration/stapel_image/git_directive.html#подробнее-про-gitarchive-gitcache-gitlatestpatch).

{% offtopic title="Можно ли использовать насколько репозиториев?" %}

Директива git также позволяет добавлять код из удалённых git-репозиториев, используя параметр `url`

Пример:
{% raw %}
```yaml
git:
- add: /src
  to: /app
- url: https://github.com/ariya/phantomjs
  add: /
  to: /src/phantomjs
```
{% endraw %}
Детали и особенности можно почитать в [документации]({{ site.baseurl }}/documentation/configuration/stapel_image/git_directive.html#работа-с-удаленными-репозиториями).

{% endofftopic %}

{% offtopic title="Реальная практика" %}

В реальной практике нужно добавлять файлы **фильтруя по названию или пути**.

К примеру, в данном варианте мы добавляем все файлы `.php` и `.js` из папки `/src`, исключая файлы с суффиксом `-dev.` и `-test.` в имени файла.
{% raw %}
```yaml
git:
- add: /src
  to: /app
  includePaths:
  - '**/*.php'
  - '**/*.js'
  excludePaths:
  - '**/*-dev.*'
  - '**/*-test.*'
```
{% endraw %}
В Linux системах важно не забывать про установку **владельца файлов**.

Например, вот так мы изменяем владельца файла `index.php` на `www-data`.

```yaml
git:
- add: /src/index.php
  to: /app/index.php
  owner: www-data
```

Подробнее о всех опциях директивы git можно прочитать в  [документации]({{ site.baseurl }}/documentation/configuration/stapel_image/git_directive.html#Изменение-владельца).


{% endofftopic %}

Следующим этапом необходимо описать **правила [сборки для приложения](https://ru.werf.io/v1.1-alpha/documentation/configuration/stapel_image/assembly_instructions.html)**. Werf позволяет кэшировать сборку образа подобно слоям в docker, но с более явной конфигурацией. Этот механизм называется [стадиями](https://ru.werf.io/v1.1-alpha/documentation/configuration/stapel_image/assembly_instructions.html#пользовательские-стадии). Для текущего приложения опишем 2 стадии в которых сначала устанавливаем пакеты, а потом - работаем с исходными кодами приложения.

{% offtopic title="Что нужно знать про стадии" %}

Стадии — это очень важный, и в то же время непростой инструмент, который резко уменьшает время сборки приложения. Вплотную мы разберёмся с работой со стадиями, когда будем разбирать вопрос зависимостей и ассетов.

Пока нам нужно знать несколько вещей:
При правильном использовании стадий мы сильно уменьшаем время сборки приложения, на стадии `install` мы устанавливаем системные пакеты, неодходимое для работы программы, а на стадии `setup` мы работаем с исходным кодом приложения.

{% endofftopic %}

Добавим в `werf.yaml` следующий блок используя shell синтаксис:

{% snippetcut name="werf.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/werf.yaml" %}
```yaml
shell:
  beforeInstall:
  - apt update
  - apt install -y tzdata locales
  install:
  - cd /app && npm сi
```
{% endsnippetcut %}

Чтобы при запуске приложения по умолчанию использовалась директория `/app` - воспользуемся **[указанием Docker-инструкций](https://ru.werf.io/v1.1-alpha/documentation/configuration/stapel_image/docker_directive.html)**:

{% snippetcut name="werf.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/werf.yaml" %}
```yaml
docker:
  WORKDIR: /app
```
{% endsnippetcut %}

```yaml
  stageDependencies:
    install:
    - package.json
```
Данная конструкция отвечает за отслеживание изменений в файле `package.json` и пересборки стадии `install` в случае нахождения таковых.

Когда `werf.yaml` готов (или кажется таковым) — пробуем запустить сборку:

```bash
$  werf build --stages-storage :local
```

{% offtopic title="Что делать, если что-то пойдёт не так?" %}

Werf предоставляет удобные способы отладки.

Если во время сборки что-то пошло не так — вы можете мгновенно оказаться внутри процесса сборки и изучить, что конкретно пошло не так. Для этого есть механизм [интроспекции стадий](https://ru.werf.io/v1.1-alpha/documentation/reference/development_and_debug/stage_introspection.html).
Если сборка запущена с флагом `--introspect-before-error` — вы окажетесь в сборочном контейнере перед тем, как сработала ошибка.

Например, если ошибка произошла на команде `jekyll build`. Если вы воспользуетесь интроспекцией — то вы окажетесь внутри контейнера прямо перед тем, как должна была выполняться эта команда.
А значит, выполнив её самостоятельно — вы можете увидеть сообщение об ошибке, и посмотреть на состояние других файлов в контейнере на этот момент, чтобы понять, что в сценарии сборки пошло не так.

{% endofftopic %}

{% offtopic title="Что за stages-storage?" %}
Стадии, которые используются для ускорения сборки, должны где-то храниться. Werf подразумевает, что сборка может производиться на нескольких раннерах — а значит кэш сборочных стадий нужно хранить в каком-то едином хранилище.

`stages-storage` [позволяет настраивать](https://ru.werf.io/documentation/reference/stages_and_images.html#%D1%85%D1%80%D0%B0%D0%BD%D0%B8%D0%BB%D0%B8%D1%89%D0%B5-%D1%81%D1%82%D0%B0%D0%B4%D0%B8%D0%B9), где будет храниться кэш сборочных стадий: на локальном сервере или в registry.
{% endofftopic %}

Если всё написано правильно, то наша сборка завершится примерно так:

```
...
│ ┌ Building stage basicapp/beforeInstall
│ ├ Info
│ │     repository: werf-stages-storage/werf-guided-project
│ │       image_id: 2743bc56bbf7
│ │        created: 2020-05-26T22:44:26.0159982Z
│ │            tag: 7e691385166fc7283f859e35d0c9b9f1f6dc2ea7a61cb94e96f8a08c-1590533066068
│ │           diff: 0 B
│ │   instructions: WORKDIR /app
│ └ Building stage basicapp/beforeInstall (0.82 seconds)
└ ⛵ image basicapp (239.56 seconds)
```

В конце werf отдал информацию о готовом repository и о tag для этого образа для дальнейшего использования.

```
werf-stages-storage/werf-guided-project:981ece3acc63d57d5ab07f45fd0c0c477088649523822c2bff033df4-1594911680806
```

Запустим собранный образ с помощью [werf run](https://werf.io/documentation/cli/main/run.html):

```bash
$ werf run --stages-storage :local --docker-options="-d -p 3000:3000 --restart=always" -- node /app/src/js/index.js
```

Первая часть команды очень похожа на build, а во второй мы задаем [параметры docker](https://docs.docker.com/engine/reference/run/) и через двойную черту команду с которой хотим запустить наш image.

Теперь наше приложение доступно локально на порту 3000:

![](/documentation/guides/applications_guide/images/020-hello-world-in-browser.png)

Как только мы убедились в том, что всё корректно — мы должны **загрузить образ в Registry**. Сборка с последующей загрузкой в Registry делается [командой `build-and-publish`](https://ru.werf.io/documentation/cli/main/build_and_publish.html). Когда werf запускается внутри CI-процесса — werf узнаёт реквизиты для доступа к Registry [из переменных окружения](https://docs.gitlab.com/ee/ci/variables/predefined_variables.html).

Если мы запускаем Werf вне Gitlab CI — нам нужно:

* Вручную подключиться к gitlab registry [с помощью `docker login`](https://docs.docker.com/engine/reference/commandline/login/)
* Установить переменную окружения `WERF_IMAGES_REPO` с путём до Registry (вида `registry.mydomain.io/myproject`)
* Выполнить сборку и загрузку в Registry: `werf build-and-publish`

Если вы всё правильно сделали — вы увидите собранный образ в registry. При использовании registry от gitlab — собранный образ можно увидеть через веб-интерфейс GitLab.

<div>
    <a href="20_iac.html" class="nav-btn">Далее: Конфигурирование инфраструктуры в виде кода</a>
</div>

