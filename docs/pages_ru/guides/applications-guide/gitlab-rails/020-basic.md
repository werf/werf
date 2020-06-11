---
title: Базовые настройки
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-rails/020-basic.html
author: alexey.chazov <alexey.chazov@flant.com>
layout: guide
toc: false
author_team: "bravo"
author_name: "alexey.chazov"
ci: "gitlab"
language: "ruby"
framework: "rails"
is_compiled: 0
package_managers_possible:
 - bundler
package_managers_chosen: "bundler"
unit_tests_possible:
 - Rspec
unit_tests_chosen: "Rspec"
assets_generator_possible:
 - webpack
 - gulp
assets_generator_chosen: "webpack"
---

В этой главе мы возьмём приложение, которое будет выводить сообщение "hello world" по http и опубликуем его в kubernetes с помощью Werf. Сперва мы разберёмся со сборкой и добьёмся того, чтобы образ оказался в Registry, затем — разберёмся с деплоем собранного приложения в Kubernetes, и, наконец, организуем CI/CD-процесс силами Gitlab CI. 

Наше приложение будет состоять из одного docker образа собранного с помощью werf. В этом образе будет работать один основной процесс который запустит веб сервер для ruby. Управлять маршрутизацией запросов к приложению будет управлять Ingress в kubernetes кластере. Мы реализуем два стенда: [production](https://ru.werf.io/documentation/reference/ci_cd_workflows_overview.html#production) и [staging](https://ru.werf.io/documentation/reference/ci_cd_workflows_overview.html#staging).

![](/images/applications-guide/gitlab-rails/020-basic-process-overview.png)

<a name="building" />

## Сборка

В этой главе мы научимся писать конфигурацию сборки, отлаживать её в случае ошибок и загружать полученный образ в Registry.

{% offtopic title="Локальная разработка или на сервере?" %}

TODO: расписать проблематику где вести разработку: на локальной тачке или на сервере.

На локальной тачке может не быть ресурсов для того, чтобы гонять что билды, что разворачивать кубернетесы потом. Поэтому хочется разворачивать на серваке помощнее. Можно и так, конечно.

Но тащемта похуй

В ближайшее время werf реализует удобные инструменты для локальной разработки, где будет подниматься миникуб.

{% endofftopic %}

Возьмём исходные коды приложения из git:

```bash
$ git clone http://....... TODO
```

Для того чтобы werf смог собрать docker-образ с приложением - необходимо в корне нашего репозитория создать файл `werf.yaml` в которым будут описаны инструкции по сборке.

{% offtopic title="Варианты синтаксиса werf.yaml" %}

TODO: вариантов синтаксиса несколько вот раз два три подробнее читать там, а мы будем рассматривать Shell

{% endofftopic %}

Начнём werf.yaml с обязательной [**секции мета-информации**]({{ site.baseurl }}/documentation/configuration/introduction.html#секция-мета-информации):

{% snippetcut name="werf.yaml" url="gitlab-rails-files/examples/example_1/werf.yaml" %}
```yaml
project: example-1
configVersion: 1
```
{% endsnippetcut %}

**_project_** - поле, задающее имя для проекта, которым мы определяем связь всех docker images собираемых в данном проекте. Данное имя по умолчанию используется в имени helm релиза и имени namespace в которое будет выкатываться наше приложение. Данное имя не рекомендуется изменять (или подходить к таким изменениям с должным уровнем ответственности) так как после изменений уже имеющиеся ресурсы, которые выкачаны в кластер, не будут переименованы.

**_configVersion_** - в данном случае определяет версию синтаксиса используемую в `werf.yaml`.

После мы сразу переходим к следующей секции конфигурации, которая и будет для нас основной секцией для сборки - [**image config section**](https://ru.werf.io/v1.1-alpha/documentation/configuration/introduction.html#%D1%81%D0%B5%D0%BA%D1%86%D0%B8%D1%8F-%D0%BE%D0%B1%D1%80%D0%B0%D0%B7%D0%B0).

{% snippetcut name="werf.yaml" url="gitlab-rails-files/examples/example_1/werf.yaml" %}
```yaml
project: example-1
configVersion: 1
---
image: rails
from: ruby:2.7.1
```
{% endsnippetcut %}

В строке `image: rails` дано название для образа, который соберёт werf. Это имя мы впоследствии будем указывать при запуске контейнера. Строка `from: ruby:2.7.1` определяет, что будет взято за основу, мы берем официальный публичный образ с нужной нам версией ruby.

{% offtopic title="Что делать, если образов и других констант станет много" %}

TODO: кратко написать про шаблоны Go и дать ссылку на https://ru.werf.io/v1.1-alpha/documentation/configuration/introduction.html#%D1%88%D0%B0%D0%B1%D0%BB%D0%BE%D0%BD%D1%8B-go

{% endofftopic %}

Добавим исходный код нашего приложения в контейнер с помощью [**директивы git**](https://ru.werf.io/v1.1-alpha/documentation/configuration/stapel_image/git_directive.html)

{% snippetcut name="werf.yaml" url="gitlab-rails-files/examples/example_1/werf.yaml" %}
```yaml
project: example-1
configVersion: 1
---
image: rails
from: ruby:2.7.1
git:
- add: /
  to: /app
```
{% endsnippetcut %}

При использовании директивы `git`, werf использует репозиторий проекта, в котором расположен `werf.yaml`. Добавляемые директории и файлы проекта указываются списком относительно корня репозитория.

`add: /` - та директория которую мы хотим добавить внутрь docker image, мы указываем, что это весь наш репозиторий

`to: /app` - то куда мы клонируем наш репозиторий внутри docker image. Важно заметить что директорию назначения werf создаст сам.

Как такового репозитория в образе не создаётся, директория .git отсутствует. На одном из первых этапов сборки все исходники пакуются в архив (с учётом всех пользовательских параметров) и распаковываются внутри контейнера. При последующих сборках [накладываются патчи с разницей](https://ru.werf.io/v1.1-alpha/documentation/configuration/stapel_image/git_directive.html#подробнее-про-gitarchive-gitcache-gitlatestpatch).

{% offtopic title="Можно ли использовать насколько репозиториев?" %}

Директива git также позволяет добавлять исходники из удалённых git-репозиториев, используя параметр `url` (детали и особенности можно почитать в [документации]({{ site.baseurl }}/documentation/configuration/stapel_image/git_directive.html#работа-с-удаленными-репозиториями)).

{% endofftopic %}

{% offtopic title="Реальная практика" %}

TODO: В реальной практике нужно: добавление файлов по маске (include/exclude), изменение владельца (owner/group). Надо пару слов об этом и ссылку.

{% endofftopic %}

Следующим этапом необходимо описать **правила [сборки для приложения](https://ru.werf.io/v1.1-alpha/documentation/configuration/stapel_image/assembly_instructions.html)**. Werf позволяет кэшировать сборку образа подобно слоям в docker, но с более явной конфигурацией. Этот механизм называется [стадиями](https://ru.werf.io/v1.1-alpha/documentation/configuration/stapel_image/assembly_instructions.html#пользовательские-стадии). Для текущего приложения опишем 2 стадии в которых сначала устанавливаем пакеты, а потом - работаем с исходными кодами приложения.

{% offtopic title="Что нужно знать про стадии" %}

TODO: Стадии — это очень важный инструмент, который резко уменьшает время сборки приложения. Вплотную мы разберёмся с работой со стадиями, когда будем разбирать вопрос зависимостей и ассетов. Пока что — хватит интуитивного понимания происходящего.

{% endofftopic %}

Добавим в `werf.yaml` следующий блок используя shell синтаксис:

{% snippetcut name="werf.yaml" url="#" %}
```yaml
shell:
  beforeInstall:
  - gem install bundler
  install:
  - bundle config set without 'development test' && bundle install
```
{% endsnippetcut %}

{% offtopic title="Shell-синтаксис vs Ansible-синтаксис" %}

TODO: Про то, что есть два варианта шелл или ансибл. Какие-то ссылки на объяснение.

TODO: И то, что вот скрипт выше можно было написать вот так в ансибл синтаксисе.

TODO: какая-то ссылка на поддерживаемые модули

Полный список поддерживаемых модулей ansible в werf можно найти [тут](https://werf.io/documentation/configuration/stapel_image/assembly_instructions.html#supported-modules).

{% endofftopic %}

Чтобы при запуске приложения по умолчанию использовалась дириктория `/app` - воспользуемся **[указанием Docker-инструкций](https://ru.werf.io/v1.1-alpha/documentation/configuration/stapel_image/docker_directive.html)**:

{% snippetcut name="werf.yaml" url="#" %}
```yaml
docker:
  WORKDIR: /app
```
{% endsnippetcut %}

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
│ ┌ Building stage rails/dockerInstructions
│ ├ Info
│ │     repository: werf-stages-storage/example-1
│ │       image_id: 2743bc56bbf7
│ │        created: 2020-05-26T22:44:26.0159982Z
│ │            tag: 7e691385166fc7283f859e35d0c9b9f1f6dc2ea7a61cb94e96f8a08c-1590533066068
│ │           diff: 0 B
│ │   instructions: WORKDIR /app
│ └ Building stage rails/dockerInstructions (0.82 seconds)
└ ⛵ image rails (239.56 seconds)
```

{% offtopic title="Что делать, если что-то пошло не так?" %}

TODO: Werf предоставляет удобные способы отладки. Если сборка падает - есть introspect https://ru.werf.io/v1.1-alpha/documentation/reference/development_and_debug/stage_introspection.html

```bash
$  werf build --stages-storage :local --introspect-before-error
```

или

```bash
$  werf build --stages-storage :local --introspect-error
```

Которые при падении сборки на одном из шагов автоматически откроют вам shell в контейнер, перед исполнением проблемной инструкции или после.

TODO: и написать или найти в доке верфи вообще сам концепт работы-доработки с проброской в контейнер.

Потому что даже если успешно сборка закончилась — не факт, что всё правильно отработало, интроспект может быть полезен.

{% endofftopic %}

В конце werf отдал информацию о готовом image:

```
werf-stages-storage/example-1:7e691385166fc7283f859e35d0c9b9f1f6dc2ea7a61cb94e96f8a08c-1590533066068
```

Запустим собранный образ с помощью [werf run](https://werf.io/documentation/cli/main/run.html):

```bash
$ werf run --stages-storage :local --docker-options="-d -p 3000:3000 --restart=always" -- bash -c "RAILS_MASTER_KEY=0f4b75104a5b3b7ea9c98cc644ec0fa4 RAILS_ENV=production bundle exec rails server -b 0.0.0.0"
```

Первая часть команды очень похожа на build, а во второй мы задаем [параметры docker](https://docs.docker.com/engine/reference/run/) и через двойную черту команду с которой хотим запустить наш image.

Теперь наше приложение доступно локально на порту 3000:

![alt_text](/images/applications-guide/gitlab-rails/2.png "image_tooltip")

Как только мы убедились в том, что всё корректно — мы должны **загрузить образ в Registry**. Сборка с последующей загрузкой в Registry делается [командой `build-and-publish`](https://ru.werf.io/documentation/cli/main/build_and_publish.html). Когда werf запускается внутри CI-процесса — werf узнаёт реквизиты для доступа к Registry [из переменных окружения](https://docs.gitlab.com/ee/ci/variables/predefined_variables.html).

Если мы запускаем Werf вне Gitlab CI — нам нужно:

* Вручную подключиться к gitlab registry [с помощью `docker login`](https://docs.docker.com/engine/reference/commandline/login/)
* Установить переменную окружения `WERF_IMAGES_REPO` с путём до Registry (вида `registry.mydomain.com/myproject`)
* Выполнить сборку и загрузку в Registry: `werf build-and-publish`

Если вы всё правильно сделали — вы увидите собранный образ в registry. При использовании registry от gitlab — собранный образ можно увидеть через веб-интерфейс GitLab.

<a name="config-iac" />

## Конфигурирование инфраструктуры в виде кода

Для того, чтобы приложение заработало в Kubernetes — необходимо описать инфраструктуру приложения как код, т.е. в нашем случае объекты kubernetes: Pod, Service и Ingress.

Конфигурацию для Kubernetes нужно шаблонизировать. Один из популярных инструментов для такой шаблонизации — это Helm, и движок Helm-а встроен в Werf. Помимо этого, werf предоставляет возможности работы с секретными значениями, а также дополнительные Go-шаблоны для интеграции собранных образов.

В этой главе мы научимся описывать helm-шаблоны, используя возможности werf, а также освоим встроенные инструменты отладки. 

<a name="iac-configs-making" />

### Составление конфигов инфраструктуры

{% filesused title="Файлы, упомянутые в главе" %}
- [.helm/templates/deployment.yaml](.helm/templates/deployment.yaml)
- [.helm/templates/ingress.yaml](.helm/templates/ingress.yaml)
- [.helm/templates/service.yaml](.helm/templates/service.yaml)
- [.helm/values.yaml](.helm/values.yaml)
- [.helm/secret-values.yaml](.helm/secret-values.yaml)
{% endfilesused %}


На сегодняшний день [Helm](https://helm.sh/) один из самых удобных способов которым вы можете описать свой deploy в Kubernetes. Кроме возможности установки готовых чартов с приложениями прямиком из репозитория, где вы можете введя одну команду, развернуть себе готовый Redis, Postgres, Rabbitmq прямиком в Kubernetes, вы также можете использовать Helm для разработки собственных чартов с удобным синтаксисом для шаблонизации выката ваших приложений.

Потому для werf это был очевидный выбор использовать такую технологию.

{% offtopic title="Что делать, если вы не работали с Helm?" %}

Мы не будем вдаваться в подробности [разработки yaml манифестов с помощью Helm для Kubernetes](https://habr.com/ru/company/flant/blog/423239/). Осветим лишь отдельные её части, которые касаются данного приложения и werf в целом. Если у вас есть вопросы о том как именно описываются объекты Kubernetes, советуем посетить страницы документации по Kubernetes с его [концептами](https://kubernetes.io/ru/docs/concepts/) и страницы документации по разработке [шаблонов](https://helm.sh/docs/chart_template_guide/) в Helm.

Работа с Helm и конфигурацией для Kubernetes может быть очень сложной первое время из-за нелепых мелочей, таких, как опечатки или пропущенные пробелы. Если вы только начали осваивать эти технологии — постарайтесь найти ментора, который поможет вам преодолеть эти сложности и посмотрит на ваши исходники сторонним взглядом. 

В случае затруднений, пожалуйста убедитесь, что вы

- Понимаете, как работает indent
- Понимаете, что такое конструкция tuple
- Понимаете, как Helm работает с хэш-массивами
- Очень внимательно следите за пробелами в Yaml

TODO:  ^^ вот это выше надо причесать, дать ссылки

{% endofftopic %}

Для работы нашего приложения в среде Kubernetes понадобится описать сущности Deployment (который породит в кластере Pod), Service, завернуть трафик на приложение, донастроив роутинг в кластере с помощью сущности Ingress. И не забыть создать отдельную сущность Secret, которая позволит нашему kubernetes пулить собранные образа из registry.

#### Создание Pod-а

{% filesused title="Файлы, упомянутые в главе" %}
- [.helm/templates/deployment.yaml](.helm/templates/deployment.yaml)
- [.helm/values.yaml](.helm/values.yaml)
{% endfilesused %}

Для того, чтобы в кластере появился Pod с нашим приложением — мы пропишем объект Deployment. У создаваемого Pod будет один контейнер `rails`. Укажем, **как этот контейнер будет запускаться**:

{% snippetcut name="deployment.yaml" url="gitlab-rails-files/examples/example_1/.helm/templates/deployment.yaml#L17" %}
{% raw %}
```yaml
      containers:
      - name: rails
        command: ["bundle", "exec", "rails", "server", "-b", "0.0.0.0"]
{{ tuple "rails" . | include "werf_container_image" | indent 8 }}
```
{% endraw %}
{% endsnippetcut %}

Обратите внимание на [вызов `werf_container_image`](https://ru.werf.io/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_image). Данная функция генерирует ключи `image` и `imagePullPolicy` со значениями, необходимыми для соответствующего контейнера пода и это позволяет гарантировать перевыкат контейнера тогда, когда это нужно.

{% offtopic title="А в чём проблема?" %}
Kubernetes не знает ничего об изменении контейнера — он действует на основании описания объектов и сам выкачивает образы из Registry. Поэтому Kubernetes-у нужно в явном виде сообщать, что делать.

Werf складывает собранные образы в Registry с разными именами, в зависимости от выбранной стратегии тегирования и деплоя — подробнее это мы разберём в главе про CI. И, как следствие, в описание контейнера нужно пробрасывать правильный путь до образа, а также дополнительные аннотации, связанные со стратегией деплоя. 

Подробнее - можно посмотреть в [документации](https://ru.werf.io/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_image).
{% endofftopic %}

Для корректной работы нашего приложения ему нужно узнать **переменные окружения**.

Для Ruby on Rails это, например, `RAILS_ENV` в которой указывается окружение. Так же нужно добавить переменную `RAILS_MASTER_KEY` которая расшифровывает секретные значения.

{% snippetcut name="deployment.yaml" url="gitlab-rails-files/examples/example_1/.helm/templates/deployment.yaml#L21" %}
{% raw %}
```yaml
      env:
      - name: RAILS_MASTER_KEY
        value: 9eeddad83cebe240f55ae06ccdd95f8e
      - name: RAILS_ENV
        value: production
{{ tuple "rails" . | include "werf_container_env" | indent 8 }}
```
{% endraw %}
{% endsnippetcut %}

Мы задали значение для `RAILS_MASTER_KEY` в явном виде — и это абсолютно не безопасный путь для хранения таких критичных данных. Мы разберём более правильный путь ниже, в главе "Разное поведение в разных окружениях".

Обратите также внимание на [функцию `werf_container_env`](https://ru.werf.io/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_env) — с помощью неё Werf вставляет в описание объекта служебные переменые окружения.

{% offtopic title="А как динамически подставлять в переменные окружения нужные значения?" %}

<a name="helm-values-yaml" /> Helm — шаблонизатор, и он поддерживает множество инструментов для подстановки значений. Один из центральных способов — подставлять значения из файла `values.yaml`. Наша конструкция могла бы иметь вид 

{% snippetcut name="deployment.yaml" url="#" %}
{% raw %}
```yaml
      env:
      - name: RAILS_MASTER_KEY
        value: {{ .Values.rails.master_key }}
```
{% endraw %}
{% endsnippetcut %}

или даже более сложный, для того, чтобы значение основывалось на текущем окружении

{% snippetcut name="deployment.yaml" url="#" %}
```yaml
      env:
      - name: RAILS_MASTER_KEY
        value: {% raw %}{{ pluck .Values.global.env .Values.rails.master_key | first | default .Values.rails.master_key._default }}{% endraw %}
```
{% endsnippetcut %}

{% snippetcut name="values.yaml" url="#" %}
```yaml
rails:
  master_key:
    _default: 9eeddad83cebe240f55ae06ccdd95f8e
    production: 684e5cb73034052dc89e3055691b7ac4
    testing: 189af8ca60b04e529140ec114175f098
```
{% endsnippetcut %}

TODO: вот эти варианты надо оформить и куда-то положить, чтобы была валидная ссылка

{% endofftopic %}


При запуске приложения в kubernetes необходимо **логи отправлять в stdout и stderr** - это необходимо для простого сбора логов например через `filebeat`, а так же чтобы не разрастались docker образы запущенных приложений.

Для того чтобы логи приложения отправлялись в stdout нам необходимо будет добавить переменную окружения `RAILS_LOG_TO_STDOUT="true"` согласно [изменениям](https://github.com/rails/rails/pull/23734) в rails framework.

{% snippetcut name="deployment.yaml" url="#" %}
```yaml
      - name: RAILS_LOG_TO_STDOUT
        value: "true"
```
{% endsnippetcut %}

#### Доступность Pod-а

{% filesused title="Файлы, упомянутые в главе" %}
- [.helm/templates/deployment.yaml](.helm/templates/deployment.yaml)
- [.helm/templates/service.yaml](.helm/templates/service.yaml)
- [.helm/values.yaml](.helm/values.yaml)
{% endfilesused %}

Для того чтобы запросы извне попали к нам в приложение нужно открыть порт у Pod-а, создать объект Service и привязать его к Pod-у и создать объект Ingress.

{% offtopic title="Что за объект Ingress и как он связан с балансером?" %}
Возможна коллизия терминов:

* Есть Ingress — в смысле [NGINX Ingress Controller](https://github.com/kubernetes/ingress-nginx), который работает в кластере и принимает входящие извне запросы
* И есть [объект Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/), который фактически описывает настройки для Ingress Controller

В статьях и бытовой речи оба этих термина зачастую называют "Ingress", так что нужно догадываться по контексту.
{% endofftopic %}

Наше приложение работает на стандартном порту `3000` — **откроем порт Pod-у**:

{% snippetcut name="deployment.yaml" url="#" %}
```yaml
        ports:
        - containerPort: 3000
          name: http
          protocol: TCP
```
{% endsnippetcut %}

Затем, **пропишем Service**, чтобы к Pod-у могли обращаться другие приложения кластера. 

{% snippetcut name="service.yaml" url="gitlab-rails-files/examples/example_1/.helm/templates/service.yaml" %}
{% raw %}
```yaml
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Chart.Name }}
spec:
  selector:
    service: {{ .Chart.Name }}
  ports:
  - name: http
    port: 3000
    protocol: TCP
```
{% endraw %}
{% endsnippetcut %}

Обратите внимание на поле `selector` у Service: он должен совпадать с аналогичным полем у Deployment и ошибки в этой части — самая частая проблема с настройкой маршрута до приложения.

{% snippetcut name="deployment.yaml" url="gitlab-rails-files/examples/example_1/.helm/templates/deployment.yaml" %}
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
TODO: написать, что надо курлануть с любого пода по имени сервиса или пингануть и вот буквально абзац об этом.
НО ПРИ НАПИСАНИИ ЭТОГО АБЗАЦА НАДО ОБЯЗАТЕЛЬНО УЧЕСТЬ, что мы в этом повествовании до непосредственно применения конфигов в кубернетес придём сильно позже.
То есть надо писать что-то типа "мы будем деплоиться в кубы позже. Но если после деплоя не заведётся — вернитесь сюда и проведите проверку такую-то. В случае таких проблем — это вот оно."
{% endofftopic %}

После этого можно настраивать **роутинг на Ingress**. Укажем, запросы на какой домен и путь по какому протоколу направлять в какой сервис и на какой порт.

{% snippetcut name="ingress.yaml" url="gitlab-rails-files/examples/example_1/.helm/templates/ingress.yaml" %}
{% raw %}
```yaml
  rules:
  - host: myrailsapp.io
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

Некоторые настройки хочется видеть разными в разных окружениях. К примеру, домен, на котором будет открываться приложение должен быть либо staging.myrailsapp.io, либо myrailsapp.io — смотря куда мы задеплоились.

В werf есть для этого есть три механики:

1. Подстановка значений из `values.yaml` по аналогии с Helm
2. Проброска значений через аттрибут `--set` при работе в CLI-режиме, по аналогии с Helm
3. Подстановка секретных значений из `secret-values.yaml`

**Вариант с `values.yaml`** рассматривался ранее [в главе "Создание Pod-а"](#helm-values-yaml).

Второй вариант подразумевает **задание переменных через CLI** `werf deploy --set "global.ci_url=example.com"`, которое затем будет доступно в yaml-ах в виде {% raw %}`{{ .Values.global.ci_url }}`{% endraw %}.

Этот вариант удобен для проброски, например, имени домена для каждого окружения

{% snippetcut name="ingress.yaml" url="gitlab-rails-files/examples/example_1/.helm/templates/ingress.yaml" %}
{% raw %}
```yaml
  rules:
  - host: {{ .Values.global.ci_url }}
```
{% endraw %}
{% endsnippetcut %}

<a name="secret-values-yaml" />Отдельная проблема — **хранение и задание секретных переменных**, например, учётных данных аутентификации для сторонних сервисов, API-ключей и т.п.

Так как Werf рассматривает git как единственный источник правды — корректно хранить секретные переменные там же. Чтобы делать это корректно — мы [храним данные в шифрованном виде](https://ru.werf.io/documentation/reference/deploy_process/working_with_secrets.html). Подстановка значений из этого файла происходит при рендере шаблона, который также запускается при деплое.

Чтобы продолжать дальше 

* [Сгенерируйте ключ](https://ru.werf.io/documentation/cli/management/helm/secret/encrypt.html) (`werf helm secret generate-secret-key`)
* Задайте ключ в переменных приложения, в текущей сессии консоли (например, `export WERF_SECRET_KEY=504a1a2b17042311681b1551aa0b8931z`)
* Пропишите полученный ключ в Variables для вашего репозитория в Gitlab (раздел `Settings` - `CI/CD`), название переменной `WERF_SECRET_KEY`

![alt_text](/images/applications-guide/gitlab-rails/5.png "image_tooltip")

После этого мы сможем задать секретную переменную `RAILS_MASTER_KEY`. Зайдите в режим редактирования секретных значений:

```bash
$ werf helm secret values edit .helm/secret-values.yaml
```

Откроется консольный текстовый редактор с данными в расшифованном виде:

{% snippetcut name="secret-values.yaml в расшифрованном виде" url="gitlab-rails-files/examples/example_1/.helm/secret-values.yaml" %}
```yaml
rails:
  master_key: 0f4b75104a5b3b7ea9c98cc644ec0fa4
```
{% endsnippetcut %}

После сохранения значения в файле зашифруются и примут примерно такой вид:

{% snippetcut name="secret-values.yaml в зашифрованном виде" url="gitlab-rails-files/examples/example_1/.helm/secret-values.yaml" %}
```yaml
rails:
  master_key: 100083f330adfb9e13fff74c9ab71b93ed77704aca3a0c607679336e099d48977d6565b314c06b8ad7aefc9d8d90629e92d851b573a89915ff036239de129d722ef5
```
{% endsnippetcut %}

<a name="iac-debug-deploy" />

### Отладка конфигов инфраструктуры и деплой в Kubernetes

После того, как написана основная часть конфигов — хочется проверить корректность конфигов и задеплоить их в kubernetes. Для того, чтобы отрендерить конфиги инфраструктуры нужны сведения об окружении, на который будет произведён деплой, ключ для расшифровки секретных значений и т.п.

Если мы запускаем Werf вне Gitlab CI — нам нужно сделать несколько операций вручную прежде чем Werf сможет рендерить конфиги и деплоить в Kubernetes.

* Вручную подключиться к gitlab registry [с помощью `docker login`](https://docs.docker.com/engine/reference/commandline/login/) (если ранее это не сделано)
* Установить переменную окружения `WERF_IMAGES_REPO` с путём до Registry (вида `registry.mydomain.com/myproject`)
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

{% offtopic title="Что делать, если что-то пошло не так?" %}

{% endofftopic %}

Проверить что приложение задеплоилось в кластер можно с помощью kubectl, вы увидите что-то вида:

```bash
$ kubectl get namespace
NAME                          STATUS   AGE
default                       Active   161d
example-1-production          Active   4m44s
example-1-stage               Active   3h2m

$ kubectl -n example-1-stage get po
NAME                        READY   STATUS    RESTARTS   AGE
example-1-9f6bd769f-rm8nz   1/1     Running   0          6m12s

$ kubectl -n example-1-stage get ingress
NAME        HOSTS                                           ADDRESS   PORTS   AGE
example-1   stage.myrailsapp.io                                       80      6m18s
```

А также вы должны увидеть ваш сервис через браузер.

TODO: ОБЯЗАТЕЛЬНО нужно показать как оно работает в браузере. "Задеплоено" — это не критерий работы.

<a name="ci" />

## Построение CI-процесса

После того, как мы разобрались, как делать сборку и деплой "вручную" — пора автоматизировать процесс.

Мы предлагаем простой флоу, который мы называем [fast and furious](https://docs.google.com/document/d/1a8VgQXQ6v7Ht6EJYwV2l4ozyMhy9TaytaQuA9Pt2AbI/edit#). Такой флоу позволит вам осуществлять быструю доставку ваших изменений в production согласно методологии GitOps и будут содержать два окружения, production и stage.

Начнем с того что добавим нашу сборку в CI с помощью `.gitlab-ci.yml`, который находится внутри корня проекта и опишем там заготовки для всех стадий и общий код, обеспечивающий работу werf.

{% snippetcut name=".gitlab-ci.yaml" url="#" %}
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

Deploy to stage:
  script:
    - echo "todo deploy to stage"
  environment:
    name: stage
    url: http://stage.myrailsapp.io
  only:
    - merge_requests
  when: manual

Deploy to production:
  script:
    - echo "todo deploy to production"
  environment:
    name: production
    url: http://myrailsapp.io
  only:
    - master
```
{% endsnippetcut %}

TODO: ^^^ чёто мне кажется это не fast&furious! Надо проверить и пофиксить

{% offtopic title="Зачем используется multiwerf?" %}
Такой сложный путь с использованием multiwerf нужен для того, чтобы вам не надо было думать про обновление верфи и установке новых версий — вы просто указываете, что используете, например, use 1.1 stable и пребываете в уверенности, что у вас актуальная версия с закрытыми issues.
{% endofftopic %}

{% offtopic title="Что за werf ci-env gitlab?" %}
Практически все опции werf можно задавать переменными окружения. Команда ci-env проставляет предопределенные значения опций, которые будут использоваться всеми командами werf в shell-сессии, на основе переменных окружений CI:

```bash
### DOCKER CONFIG
export DOCKER_CONFIG="/tmp/werf-docker-config-832705503"
### STAGES_STORAGE
export WERF_STAGES_STORAGE="registry.gitlab-example.com/chat/stages"
### IMAGES REPO
export WERF_IMAGES_REPO="registry.gitlab-example.com/chat"
export WERF_IMAGES_REPO_IMPLEMENTATION="gitlab"
### TAGGING
export WERF_TAG_BY_STAGES_SIGNATURE="true"
### DEPLOY
# export WERF_ENV=""
export WERF_ADD_ANNOTATION_PROJECT_GIT="project.werf.io/git=https://lab.gitlab-example.com/chat"
export WERF_ADD_ANNOTATION_CI_COMMIT="ci.werf.io/commit=61368705db8652555bd96e68aadfd2ac423ba263"
export WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL="gitlab.ci.werf.io/pipeline-url=https://lab.gitlab-example.com/chat/pipelines/71340"
export WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL="gitlab.ci.werf.io/job-url=https://lab.gitlab-example.com/chat/-/jobs/184837"
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

Пропишем подробнее в стадии сборки уже знакомую команду сборки:

{% snippetcut name=".gitlab-ci.yaml" url="#" %}
```yaml
Build:
  stage: build
  script:
    - werf build-and-publish
```
{% endsnippetcut %}

TODO: здесь и в следующих CI-стейджах разобраться с `--stages-storage :local`

Теперь при коммите кода в gitlab будет происходить сборка

![alt_text](/images/applications-guide/gitlab-rails/3.png "image_tooltip")

<a name="ci-deploy" />

### Деплой в Gitlab CI

Опишем деплой приложения в Kubernetes. Деплой будет осуществляться на два стенда: staging и production.

Выкат на два стенда отличается только параметрами, поэтому воспользуемся шаблонами. Опишем базовый деплой, который потом будем кастомизировать под стенды:

{% snippetcut name=".gitlab-ci.yml" url="gitlab-rails-files/examples/example_1/.gitlab-ci.yml#L21" %}
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

{% snippetcut name=".gitlab-ci.yml" url="gitlab-rails-files/examples/example_1/.gitlab-ci.yml#L21" %}
 ```yaml
 Deploy to Stage:
   extends: .base_deploy
   stage: deploy
   environment:
     name: stage
```
{% endsnippetcut %}

TODO: а в нашем случае точно тут должен быть stage? До этого мы называли его staging. Надо разобраться.

Аналогичным образом — настраиваем production окружение.

TODO: то, что написано ниже надо проверить на соответствие FAST&FURIOUS

После описания стадий выката при создании Merge Request и будет доступна кнопка Deploy to Stage.

![alt_text](/images/applications-guide/gitlab-rails/6.png "image_tooltip")

Посмотреть статус выполнения pipeline можно в интерфейсе gitlab **CI / CD - Pipelines**

![alt_text](/images/applications-guide/gitlab-rails/7.png "image_tooltip")

TODO: наверное надо написать ещё главу про очистку образов!
