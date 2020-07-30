---
title: Базовые настройки
sidebar: applications_guide
permalink: documentation/guides/applications_guide/gitlab_nodejs/020_basic.html
layout: guide
toc: false
---

В этой главе мы возьмём приложение, которое будет выводить сообщение "hello world" по http и опубликуем его в kubernetes с помощью Werf. Сперва мы разберёмся со сборкой и добьёмся того, чтобы образ оказался в Registry, затем — разберёмся с деплоем собранного приложения в Kubernetes, и, наконец, организуем CI/CD-процесс силами Gitlab CI.

Наше приложение будет состоять из одного docker образа собранного с помощью werf. В этом образе будет работать один основной процесс, который запустит Node. Управлять маршрутизацией запросов к приложению будет Ingress в Kubernetes кластере. Мы реализуем два стенда: [production](https://ru.werf.io/documentation/reference/ci_cd_workflows_overview.html#production) и [staging](https://ru.werf.io/documentation/reference/ci_cd_workflows_overview.html#staging).

<a name="building" />

## Сборка

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

{% offtopic title="Что делать, если что-то пошло не так?" %}

Werf предоставляет удобные способы отладки.

Если во время сборки что-то пошло не так, то можно воспользоваться [интроспекцией стадий](https://ru.werf.io/v1.1-alpha/documentation/reference/development_and_debug/stage_introspection.html), которая при падении сборки на каком либо из шагов вы автоматически окажетесь внутри собираемого вами контейнера, и сможете наглядно понять его текущее состояние и разобраться с ошибкой.

{% endofftopic %}

В конце werf отдал информацию о готовом `repository` и о `tag` для этого образа для дальнейшего использования.

```
werf-stages-storage/werf-guided-project:981ece3acc63d57d5ab07f45fd0c0c477088649523822c2bff033df4-1594911680806
```

Запустим собранный образ с помощью [werf run](https://werf.io/documentation/cli/main/run.html):

```bash
$ werf run --stages-storage :local --docker-options="-d -p 3000:3000 --restart=always" -- node /app/src/js/index.js
```

Первая часть команды очень похожа на build, а во второй мы задаем [параметры docker](https://docs.docker.com/engine/reference/run/) и через двойную черту команду с которой хотим запустить наш image.

Теперь наше приложение доступно локально на порту 3000:

![](/images/applications_guide/images/020-hello-world-in-browser.png)

Как только мы убедились в том, что всё корректно — мы должны **загрузить образ в Registry**. Сборка с последующей загрузкой в Registry делается [командой `build-and-publish`](https://ru.werf.io/documentation/cli/main/build_and_publish.html). Когда werf запускается внутри CI-процесса — werf узнаёт реквизиты для доступа к Registry [из переменных окружения](https://docs.gitlab.com/ee/ci/variables/predefined_variables.html).

Если мы запускаем Werf вне Gitlab CI — нам нужно:

* Вручную подключиться к gitlab registry [с помощью `docker login`](https://docs.docker.com/engine/reference/commandline/login/)
* Установить переменную окружения `WERF_IMAGES_REPO` с путём до Registry (вида `registry.mydomain.io/myproject`)
* Выполнить сборку и загрузку в Registry: `werf build-and-publish`

Если вы всё правильно сделали — вы увидите собранный образ в registry. При использовании registry от gitlab — собранный образ можно увидеть через веб-интерфейс GitLab.

<a name="config-iac" />

## Конфигурирование инфраструктуры в виде кода

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/templates/ingress.yaml
- .helm/templates/service.yaml
- .helm/values.yaml
- .helm/secret-values.yaml
{% endfilesused %}

Для того, чтобы приложение заработало в Kubernetes — необходимо описать инфраструктуру приложения как код, т.е. в нашем случае объекты kubernetes: Pod, Service и Ingress.

Конфигурацию для Kubernetes нужно шаблонизировать. Один из популярных инструментов для такой шаблонизации — это Helm, и движок Helm-а встроен в Werf. Помимо этого, werf предоставляет возможности работы с секретными значениями, а также дополнительные Go-шаблоны для интеграции собранных образов.

В этой главе мы научимся описывать helm-шаблоны, используя возможности werf, а также освоим встроенные инструменты отладки.

<a name="iac-configs-making" />

### Составление конфигов инфраструктуры

На сегодняшний день [Helm](https://helm.sh/) один из самых удобных способов которым вы можете описать свой deploy в Kubernetes. Кроме возможности установки готовых чартов с приложениями прямиком из репозитория, где вы можете введя одну команду, развернуть себе готовый Redis, Postgres, Rabbitmq прямиком в Kubernetes, вы также можете использовать Helm для разработки собственных чартов с удобным синтаксисом для шаблонизации выката ваших приложений.

Потому для werf это был очевидный выбор использовать такую технологию.

{% offtopic title="Что делать, если вы не работали с Helm?" %}

Мы не будем вдаваться в подробности [разработки yaml манифестов с помощью Helm для Kubernetes](https://habr.com/ru/company/flant/blog/423239/). Осветим лишь отдельные её части, которые касаются данного приложения и werf в целом. Если у вас есть вопросы о том как именно описываются объекты Kubernetes, советуем посетить страницы документации по Kubernetes с его [концептами](https://kubernetes.io/ru/docs/concepts/) и страницы документации по разработке [шаблонов](https://helm.sh/docs/chart_template_guide/) в Helm.

Работа с Helm и конфигурацией для Kubernetes может быть очень сложной первое время из-за нелепых мелочей, таких, как опечатки или пропущенные пробелы. Если вы только начали осваивать эти технологии — постарайтесь найти ментора, который поможет вам преодолеть эти сложности и посмотрит на ваши исходники сторонним взглядом.

В случае затруднений, пожалуйста убедитесь, что вы:

- Понимаете, как работает [indent](https://helm.sh/docs/chart_template_guide/function_list/#indent)
- Понимаете, что такое конструкция [tuple](https://helm.sh/docs/chart_template_guide/control_structures/)
- Понимаете, как Helm работает с хэш-массивами
- Очень внимательно следите за пробелами в Yaml
{% endofftopic %}

Для работы нашего приложения в среде Kubernetes понадобится описать сущности Deployment (который породит в кластере Pod), Service, направить трафик на приложение, донастроив роутинг в кластере с помощью сущности Ingress. И не забыть создать отдельную сущность Secret, которая позволит нашему Kubernetes скачивать собранные образа из registry.

#### Создание Pod-а

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/values.yaml
{% endfilesused %}

Для того, чтобы в кластере появился Pod с нашим приложением — мы пропишем объект Deployment. У создаваемого Pod будет один контейнер `basicapp`. Укажем, **как этот контейнер будет запускаться**:

{% snippetcut name="deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.helm/templates/deployment.yaml" %}
{% raw %}
```yaml
      containers:
      - name: basicapp
        command: ["node","/app/app.js"]
{{ tuple "basicapp" . | include "werf_container_image" | indent 8 }}
```
{% endraw %}
{% endsnippetcut %}

Обратите внимание на вызов [`werf_container_image`](https://ru.werf.io/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_image). Данная функция генерирует ключи `image` и `imagePullPolicy` со значениями, необходимыми для соответствующего контейнера пода и это позволяет гарантировать перевыкат контейнера тогда, когда это нужно.

{% offtopic title="А в чём проблема?" %}
Kubernetes не знает ничего об изменении контейнера — он действует на основании описания объектов и сам выкачивает образы из Registry. Поэтому Kubernetes-у нужно в явном виде сообщать, что делать.

Werf складывает собранные образы в Registry с разными именами, в зависимости от выбранной стратегии тегирования и деплоя — подробнее это мы разберём в главе про CI. И, как следствие, в описание контейнера нужно пробрасывать правильный путь до образа, а также дополнительные аннотации, связанные со стратегией деплоя.

Подробнее - можно посмотреть в [документации](https://ru.werf.io/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_image).
{% endofftopic %}

Для корректной работы нашего приложения ему нужно узнать **переменные окружения**.

Для NodeJS это, например, `DEBUG`

{% snippetcut name="deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.helm/templates/deployment.yaml" %}
{% raw %}
```yaml
      env:
      - name: "DEBUG"
        value: "True"
{{ tuple "basicapp" . | include "werf_container_env" | indent 8 }}
```
{% endraw %}
{% endsnippetcut %}

Обратите также внимание на функцию [`werf_container_env`](https://ru.werf.io/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_env) — с помощью неё Werf вставляет в описание объекта служебные переменые окружения.

<a name="helm-values-yaml" />

{% offtopic title="А как динамически подставлять в переменные окружения нужные значения?" %}

Helm — шаблонизатор, и он поддерживает множество инструментов для подстановки значений. Один из центральных способов — подставлять значения из файла `values.yaml`. Наша конструкция могла бы иметь вид

{% snippetcut name="deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.helm/templates/deployment.yaml" %}
{% raw %}
```yaml
      env:
      - name: DEBUG_2
        value: {{ .Values.app.isDebug }}
```
{% endraw %}
{% endsnippetcut %}

или даже более сложный, для того, чтобы значение основывалось на текущем окружении:

{% snippetcut name="deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.helm/templates/deployment.yaml" %}
{% raw %}
```yaml
      env:
      - name: DEBUG_3
        value: {{ pluck .Values.global.env .Values.app.isDebug | first | default .Values.app.isDebug._default }}
```
{% endraw %}
{% endsnippetcut %}

{% snippetcut name="values.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.helm/values.yaml" %}
```yaml
app:
  isDebug:
    _default: "true"
    production: "false"
    testing: "true"
```
{% endsnippetcut %}

{% endofftopic %}


При запуске приложения в Kubernetes **логи необходимо отправлять в stdout и stderr** - это нужно для простого сбора логов например через `filebeat`, а так же чтобы не разрастались docker образы запущенных приложений.

Для того чтобы логи приложения отправлялись в stdout нам необходимо просто использовать такую конструкцию в исходном коде:

```js
console.log("I will goto the STDOUT");
console.error("I will goto the STDERR");
```
Либо если вы хотите получать логи об HTTP запросах, то можно использовать [morgan](https://www.npmjs.com/package/morgan).
```js
var morgan = require("morgan");
app.use(morgan("combined"));
```
#### Доступность Pod-а

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/templates/service.yaml
- .helm/templates/ingress.yaml
{% endfilesused %}

Для того чтобы запросы извне попали к нам в приложение нужно открыть порт у Pod-а, создать объект Service и привязать его к Pod-у, и создать объект Ingress.

{% offtopic title="Что за объект Ingress и как он связан с балансером?" %}
Возможна коллизия терминов:

* Есть Ingress — в смысле [NGINX Ingress Controller](https://github.com/kubernetes/ingress-nginx), который работает в кластере и принимает входящие извне запросы
* И есть [объект Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/), который фактически описывает настройки для NGINX Ingress Controller

В статьях и бытовой речи оба этих термина зачастую называют "Ingress", так что нужно догадываться по контексту.
{% endofftopic %}

Наше приложение работает на стандартном порту `3000` — **откроем порт Pod-у**:

{% snippetcut name="deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.helm/templates/deployment.yaml" %}
```yaml
        ports:
        - containerPort: 3000
          name: http
          protocol: TCP
```
{% endsnippetcut %}

Затем, **пропишем Service**, чтобы к Pod-у могли обращаться другие приложения кластера.

{% snippetcut name="service.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.helm/templates/service.yaml" %}
{% raw %}
```yaml
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Chart.Name }}
spec:
  selector:
    app: {{ .Chart.Name }}
  ports:
  - name: http
    port: 3000
    protocol: TCP
```
{% endraw %}
{% endsnippetcut %}

Обратите внимание на поле `selector` у Service: он должен совпадать с аналогичным полем у Deployment и ошибки в этой части — самая частая проблема с настройкой маршрута до приложения.

{% snippetcut name="deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.helm/templates/deployment.yaml" %}
{% raw %}
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}
spec:
  selector:
    matchLabels:
      app: {{ .Chart.Name }}
```
{% endraw %}
{% endsnippetcut %}

{% offtopic title="Как убедиться, что выше всё сделано правильно?" %}

Мы будем деплоить наше приложение в Kubernetes позже, но если после деплоя у вас возникли проблемы в этом месте, то вернитесь сюда и проведите проверку описанную ниже.

Попробуйте получить `endpoint` сервиса в нужном вам окружении.
Если в нем будет фигурировать ip пода, значит вы все правильно настроили. А если нет, то проверьте еще раз совпадают ли у вас поля selector в сервисе и деплойменте.
Название эндпоинта совпадает с названием сервиса.

Пример команды:

{% raw %}
`kubectl -n <название окружения> get ep {{ .Chart.Name }}`
{% endraw %}

{% endofftopic %}

После этого можно настраивать **роутинг на Ingress**. Укажем, на какой домен и путь, в какой сервис и на какой порт направлять запросы.

{% snippetcut name="ingress.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.helm/templates/ingress.yaml" %}
{% raw %}
```yaml
  rules:
  - host: mydomain.io
    http:
      paths:
      - path: /
        backend:
          serviceName: {{ .Chart.Name }}
          servicePort: 3000
```
{% endraw %}
{% endsnippetcut %}

#### Разное поведение в разных окружениях

Некоторые настройки хочется видеть разными в разных окружениях. К примеру, домен, на котором будет открываться приложение должен быть либо staging.mydomain.io, либо mydomain.io — смотря куда мы задеплоились.

В werf для этого существует три механики:

1. Подстановка значений из `values.yaml` по аналогии с Helm
2. Проброс значений через аттрибут `--set` при работе в CLI-режиме, по аналогии с Helm
3. Подстановка секретных значений из `secret-values.yaml`

**Вариант с `values.yaml`** рассматривался ранее в главе ["Создание Pod-а"](#helm-values-yaml).

Второй вариант подразумевает **задание переменных через CLI** `werf deploy --set "global.ci_url=mydomain.io"`, которое затем будет доступно в yaml-ах в виде {% raw %}`{{ .Values.global.ci_url }}`{% endraw %}.

Этот вариант удобен для проброски, например, имени домена для каждого окружения

{% snippetcut name="ingress.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.helm/templates/ingress.yaml" %}
{% raw %}
```yaml
  rules:
  - host: {{ .Values.global.ci_url }}
```
{% endraw %}
{% endsnippetcut %}

<a name="secret-values-yaml" />Отдельная проблема — **хранение и задание секретных переменных**, например, учётных данных аутентификации для сторонних сервисов, API-ключей и т.п.

Так как Werf рассматривает git как единственный источник правды — правильно хранить секретные переменные там же. Чтобы делать это корректно — мы [храним данные в шифрованном виде](https://ru.werf.io/documentation/reference/deploy_process/working_with_secrets.html). Подстановка значений из этого файла происходит при рендере шаблона, который также запускается при деплое.

Чтобы продолжать дальше

* [Сгенерируйте ключ](https://ru.werf.io/documentation/cli/management/helm/secret/generate_secret_key.html) (`werf helm secret generate-secret-key`)
* Задайте ключ в переменных приложения, в текущей сессии консоли (например, `export WERF_SECRET_KEY=504a1a2b17042311681b1551aa0b8931z`)
* Пропишите полученный ключ в Variables для вашего репозитория в Gitlab (раздел `Settings` - `CI/CD`), название переменной `WERF_SECRET_KEY`

![](/images/applications_guide/images/020-werf-secret-key-in-gitlab.png)

После этого мы сможем задать секретные переменные `access_key` и `secret_key`, например для работы с S3. Зайдите в режим редактирования секретных значений:

```bash
$ werf helm secret values edit .helm/secret-values.yaml
```

Откроется консольный текстовый редактор с данными в расшифованном виде:

{% snippetcut name="secret-values.yaml в расшифрованном виде" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.helm/secret-values.yaml" %}
```yaml
app:
  s3:
    access_key:
      _default: bNGXXCF1GF
    secret_key:
      _default: zpThy4kGeqMNSuF2gyw48cOKJMvZqtrTswAQ
```
{% endsnippetcut %}

После сохранения значения в файле зашифруются и примут примерно такой вид:

{% snippetcut name="secret-values.yaml в зашифрованном виде" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.helm/secret-values.yaml" %}
```yaml
app:
  s3:
    access_key:
      _default: 100063ead9354e0259b4cd224821cc4c57adfbd63f9d10cb1e828207f294a958d0e9
    secret_key:
      _default: 1000307fe833d63f467597d10e00ef7b321badd5d02b2f9823b0babfa7265c03e0967e434c1a496274825885e6d3645764fe097e3f463cbde27b4552b9376a78ba5f
```
{% endsnippetcut %}

<a name="iac-debug-deploy" />

### Отладка конфигов инфраструктуры и деплой в Kubernetes

После того, как написана основная часть конфигов — хочется проверить корректность конфигов и задеплоить их в Kubernetes. Для того, чтобы отрендерить конфиги инфраструктуры нужны сведения об окружении, на которое будет произведён деплой, ключ для расшифровки секретных значений и т.п.

Если мы запускаем Werf вне Gitlab CI — нам нужно сделать несколько операций вручную прежде чем Werf сможет рендерить конфиги и деплоить в Kubernetes.

* Вручную подключиться к gitlab registry с помощью [`docker login`](https://docs.docker.com/engine/reference/commandline/login/) (если ранее это не сделано)
* Установить переменную окружения `WERF_IMAGES_REPO` с путём до Registry (вида `registry.mydomain.io/myproject`)
* Установить переменную окружения `WERF_SECRET_KEY` со значением, [сгенерированным ранее в главе "Разное поведение в разных окружениях"](#secret-values-yaml)
* Установить переменную окружения `WERF_ENV` с названием окружения, в которое будет осуществляться деплой. Вопроса разных окружений мы коснёмся подробнее, когда будем строить CI-процесс, сейчас — просто установим значение `staging`

Если вы всё правильно сделали, то вы у вас корректно будут отрабатывать команды [`werf helm render`](https://ru.werf.io/documentation/cli/management/helm/render.html) и [`werf deploy`](https://ru.werf.io/documentation/cli/main/deploy.html)

{% offtopic title="Как вообще работает деплой" %}

Werf (по аналогии с helm) берет yaml шаблоны, которые описывают объекты Kubernetes, и генерирует из них общий манифест. Манифест отдается API Kubernetes, который на его основе внесет все необходимые изменения в кластер. Werf отслеживает как Kubernetes вносит изменения и сигнализирует о результатах в реальном времени. Все это благодаря встроенной в werf библиотеке [kubedog](https://github.com/flant/kubedog). Уже сам Kubernetes занимается выкачиванием нужных образов из Registry и запуском их на нужных серверах с указанными настройками.

{% endofftopic %}

Запустите деплой и дождитесь успешного завершения

```bash
werf deploy --stages-storage :local
```

Проверить что приложение задеплоилось в кластер можно с помощью kubectl, вы увидите что-то вида:

```bash
$ kubectl get namespace
NAME                                 STATUS               AGE
default                              Active               161d
werf-guided-project-production       Active               4m44s
werf-guided-project-staging          Active               3h2m
```

```bash
$ kubectl -n example-1-staging get po
NAME                                 READY                STATUS   RESTARTS  AGE
werf-guided-project-9f6bd769f-rm8nz  1/1                  Running  0         6m12s
```

```bash
$ kubectl -n example-1-staging get ingress
NAME                                 HOSTS                ADDRESS  PORTS     AGE
werf-guided-project                  staging.mydomain.io           80        6m18s
```

А также вы должны увидеть ваш сервис через браузер.

<a name="ci" />

## Построение CI-процесса

{% filesused title="Файлы, упомянутые в главе" %}
- .gitlab-ci.yaml
{% endfilesused %}

После того, как мы разобрались, как делать сборку и деплой "вручную" — пора автоматизировать процесс.

Мы предлагаем простой флоу, который мы называем [fast and furious](https://ru.werf.io/documentation/reference/ci_cd_workflows_overview.html#1-fast-and-furious). Такой флоу позволит вам осуществлять быструю доставку ваших изменений в production и будут содержать два окружения, production и staging.

Начнем с того что добавим нашу сборку в CI с помощью `.gitlab-ci.yml`, который находится в корне проекта и опишем там заготовки для всех стадий и общий код, обеспечивающий работу werf.

{% snippetcut name=".gitlab-ci.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.gitlab-ci.yml" %}
```yaml
before_script:
  - type multiwerf && source <(multiwerf use 1.1 stable)
  - type werf && source <(werf ci-env gitlab --verbose)

Build:
  stage: build
  script:
    - echo "todo build"
  tags:
    - werf

Deploy to staging:
  script:
    - echo "todo deploy to staging"
  environment:
    name: staging
    url: http://staging.mydomain.io
  only:
    - merge_requests
  when: manual

Deploy to production:
  script:
    - echo "todo deploy to production"
  environment:
    name: production
    url: http://mydomain.io
  only:
    - master
```
{% endsnippetcut %}

{% offtopic title="Зачем используется multiwerf?" %}
Такой сложный путь с использованием multiwerf нужен для того, чтобы вам не надо было думать про обновление werf и об установке новых версий — вы просто указываете, что используете, например, use 1.1 stable и пребываете в уверенности, что у вас актуальная версия.
{% endofftopic %}

{% offtopic title="Что за werf ci-env gitlab?" %}
Практически все опции werf можно задавать переменными окружения. Команда ci-env проставляет предопределенные значения, которые будут использоваться всеми командами werf в shell-сессии, на основе переменных окружений CI:

```bash
### DOCKER CONFIG
export DOCKER_CONFIG="/tmp/werf-docker-config-832705503"
### STAGES_STORAGE
export WERF_STAGES_STORAGE="registry.mydomain.io/werf-guided-project/stages"
### IMAGES REPO
export WERF_IMAGES_REPO="registry.mydomain.io/werf-guided-project"
export WERF_IMAGES_REPO_IMPLEMENTATION="gitlab"
### TAGGING
export WERF_TAG_BY_STAGES_SIGNATURE="true"
### DEPLOY
# export WERF_ENV=""
export WERF_ADD_ANNOTATION_PROJECT_GIT="project.werf.io/git=https://lab.mydomain.io/werf-guided-project"
export WERF_ADD_ANNOTATION_CI_COMMIT="ci.werf.io/commit=61368705db8652555bd96e68aadfd2ac423ba263"
export WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL="gitlab.ci.werf.io/pipeline-url=https://lab.mydomain.io/werf-guided-project/pipelines/71340"
export WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL="gitlab.ci.werf.io/job-url=https://lab.mydomain.io/werf-guided-project/-/jobs/184837"
### IMAGE CLEANUP POLICIES
export WERF_GIT_TAG_STRATEGY_LIMIT="10"
export WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS="30"
export WERF_GIT_COMMIT_STRATEGY_LIMIT="50"
export WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS="30"
export WERF_STAGES_SIGNATURE_STRATEGY_LIMIT="-1"
export WERF_STAGES_SIGNATURE_STRATEGY_EXPIRY_DAYS="-1"
### OTHER
export WERF_LOG_COLOR_MODE="on"
export WERF_LOG_PROJECT_DIR="1"
export WERF_ENABLE_PROCESS_EXTERMINATOR="1"
export WERF_LOG_TERMINAL_WIDTH="95"
```

Многие из этих переменных интуитивно понятны, и содержат базовую информацию о том, где находится проект, где находится его registry, информацию о коммитах. В рамках статьи нам хватит значений выставляемых по умолчанию.

Подробную информацию о конфигурации ci-env можно найти [тут](https://werf.io/documentation/reference/plugging_into_cicd/overview.html). Если вы используете GitLab CI/CD совместно с внешним docker registry (harbor, Docker Registry, Quay etc.), то в команду билда и пуша нужно добавлять его полный адрес (включая путь внутри registry), как это сделать можно узнать [тут](https://werf.io/documentation/cli/main/build_and_publish.html). И так же не забыть первой командой выполнить [docker login](https://docs.docker.com/engine/reference/commandline/login/).
{% endofftopic %}

<a name="ci-building" />

### Сборка в Gitlab CI

Пропишем подробнее в стадии сборки уже знакомую нам команду:

{% snippetcut name=".gitlab-ci.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.gitlab-ci.yml" %}
```yaml
Build:
  stage: build
  script:
    - werf build-and-publish
```
{% endsnippetcut %}

Теперь при коммите кода в gitlab будет происходить сборка

![](/images/applications_guide/images/020-gitlab-pipeline.png)

<a name="ci-deploy" />

### Деплой в Gitlab CI

Опишем деплой приложения в Kubernetes. Деплой будет осуществляться на два стенда: staging и production.

Выкат на два стенда отличается только параметрами, поэтому воспользуемся шаблонами. Опишем базовый деплой, который потом будем кастомизировать под стенды:

{% snippetcut name=".gitlab-ci.yml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.gitlab-ci.yml" %}
```yaml
.base_deploy: &base_deploy
  script:
    - werf deploy
  dependencies:
    - Build
  tags:
    - werf
```
{% endsnippetcut %}

В результате мы можем делать деплой, например, на staging с использованием базового шаблона:

{% snippetcut name=".gitlab-ci.yml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.gitlab-ci.yml" %}
 ```yaml
 Deploy to Staging:
   extends: .base_deploy
   stage: deploy
   environment:
     name: staging
```
{% endsnippetcut %}

Аналогичным образом — настраиваем production окружение.

После описания стадий выката при создании Merge Request и будет доступна кнопка Deploy to Staging.

![](/images/applications_guide/images/020-gitlab-mr-details.png)

Посмотреть статус выполнения pipeline можно в интерфейсе gitlab **CI / CD - Pipelines**

![](/images/applications_guide/images/020-pipelines-list.png)


### Очистка образов

При активной разработке в скором времени образы начнут занимать все больше и больше места в нашем registry. Если ничего не предпринять, то registry может разрастись до неприлично больших размеров. Для решения этой проблемы в werf реализован эффективный многоуровневый алгоритм очистки образов.

В werf реализована возможность автоматической очистки Docker registry, которая работает согласно определенных правил — политик очистки. Политики очистки определяют, какие образы можно удалять, а какие — нет.

Существует 3 основные политики очистки:

## По веткам

werf удаляет образ из Docker registry в случае отсутствия в git-репозитории соответствующей ветки. Образ никогда не удаляется, пока существует соответствующая ветка в git-репозитории.

## По коммитам

werf удаляет образ из Docker registry в случае отсутствия в git-репозитории соответствующего коммита. Для остальных образов применяется следующая политика:

Оставлять образы в Docker registry не старше указанного количества дней с момента их публикации, либо оставлять в Docker registry не более чем указанного количества образов.

## По тегам

werf удаляет образ из Docker registry в случае отсутствия в git-репозитории соответствующего тега. Для остальных образов применяется политика как и в случае с коммитами описанная выше.


При очистке по политикам никогда не удаляется в Docker registry образ, пока в кластере Kubernetes существует объект использующий такой образ. Другими словами, если вы запустили что-то в вашем кластере Kubernetes, то используемые образы ни при каких условиях не будут удалены.


Подробно можно прочитать об очистке [тут](https://ru.werf.io/documentation/reference/cleaning_process.html).

Чтобы автоматизировать очистку, добавим новую стадию cleanup в `gitlab-ci.yml`:


{% snippetcut name=".gitlab-ci.yml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/020-basic/.gitlab-ci.yml" %}
 ```yaml
Cleanup:
  stage: cleanup
  script:
    - type multiwerf && source <(multiwerf use 1.1 stable)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - docker login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_IMAGES_REPO}
    - werf cleanup --stages-storage :local
  only:
    - schedules
```
{% endsnippetcut %}

После этого добавим новую задачу в планировщик, чтобы gitlab автоматически запускал cleanup (раздел `CI/CD` - `Schedules`).

![](https://puu.sh/G0lRx/75cda37bd7.png)

В настройках указываем необходимое нам время запуска и ветку. Например, 4 утра.

![](https://puu.sh/G0lUR/7e3269e724.png)

После этого каждый день в 4 утра будет запускаться автоматическая очистка образов.


<div>
    <a href="030_dependencies.html" class="nav-btn">Далее: Подключение зависимостей</a>
</div>
