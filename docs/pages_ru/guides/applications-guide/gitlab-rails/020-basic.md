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

Как только мы убедились в том, что всё корректно — мы должны **загрузить образ в Registry**. Сборка с последующей загрузкой в Registry делается командой

```bash
$ werf build-and-publish
```

Путь до registry и другие параметры берутся автоматически их переменных окружения.

Если вы всё правильно сделали и корректно настроен registry и gitlab ci — вы увидите собранный образ в registry. При использовании registry от gitlab — собранный образ можно увидеть через веб-интерфейс GitLab.

## Конфигурирование инфраструктуры в виде кода

Для того, чтобы приложение заработало в Kubernetes — необходимо описать инфраструктуру приложения как код, т.е. в нашем случае объекты kubernetes: Pod, Service и Ingress.

Конфигурацию для Kubernetes нужно шаблонизировать. Один из популярных инструментов для такой шаблонизации — это Helm, и движок Helm-а встроен в Werf. Помимо этого, werf предоставляет возможности работы с секретными значениями, а также дополнительные Go-шаблоны для интеграции собранных образов.

В этой главе мы научимся описывать helm-шаблоны, используя возможности werf, а также освоим встроенные инструменты отладки. 

### Составление конфигов

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

Мы не будем вдаваться в подробности разработки yaml манифестов с помощью Helm для Kubernetes. Осветим лишь отдельные её части, которые касаются данного приложения и werf в целом. Если у вас есть вопросы о том как именно описываются объекты Kubernetes, советуем посетить страницы документации по Kubernetes с его [концептами](https://kubernetes.io/ru/docs/concepts/) и страницы документации по разработке [шаблонов](https://helm.sh/docs/chart_template_guide/) в Helm.

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
```yaml
      containers:
      - name: rails
        command: ["bundle", "exec", "rails", "server", "-b", "0.0.0.0"]
{% raw %}{{ tuple "rails" . | include "werf_container_image" | indent 8 }}{% endraw %}
```
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
```yaml
      env:
      - name: RAILS_MASTER_KEY
        value: 9eeddad83cebe240f55ae06ccdd95f8e
      - name: RAILS_ENV
        value: production
{% raw %}{{ tuple "rails" . | include "werf_container_env" | indent 8 }}{% endraw %}
```
{% endsnippetcut %}

Мы задали значение для `RAILS_MASTER_KEY` в явном виде — и это абсолютно не безопасный путь для хранения таких критичных данных. Мы разберём более правильный путь ниже, в главе "Секретные переменные".

Обратите также внимание на [функцию `werf_container_env`](https://ru.werf.io/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_env) — с помощью неё Werf вставляет в описание объекта служебные переменые окружения.

{% offtopic title="А как динамически подставлять в переменные окружения нужные значения?" %}

Helm — шаблонизатор, и он поддерживает множество инструментов для подстановки значений. Один из центральных способов — подставлять значения из файла `values.yaml`. Наша конструкция могла бы иметь вид 

{% snippetcut name="deployment.yaml" url="#" %}
```yaml
      env:
      - name: RAILS_MASTER_KEY
        value: {{ .Values.rails.master_key }}
```
{% endsnippetcut %}

или даже более сложный, для того, чтобы значение основывалось на текущем окружении

{% snippetcut name="deployment.yaml" url="#" %}
```yaml
      env:
      - name: RAILS_MASTER_KEY
        value: {{ pluck .Values.global.env .Values.rails.master_key | first | default .Values.rails.master_key._default }}
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

Для того чтобы логи приложения отправлялись в stdout нам необходимо будет добавить переменную окружения `RAILS_LOG_TO_STDOUT="true" `согласно [изменениям](https://github.com/rails/rails/pull/23734) в rails framework.

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

Для того чтобы запросы извне попали к нам в приложение нужно открыть порт у пода, привязать к поду сервис и настроить Ingress, который выступает у нас в качестве балансера.

Если вы мало работали с Kubernetes — эта часть может вызвать у вас много проблем. Большинство тех, кто начинает работать с Kubernetes по невнимательности допускают ошибки при конфигурировании labels и затем занимаются долгой и мучительной отладкой.




При запуске - приложение работает на стандартном порту `3000`
Описание порта который будет доступен через сервис.


```yaml
        ports:
        - containerPort: 3000
          name: http
          protocol: TCP
```

В файле `.helm/templates/service.yaml` описан сервис через который будет доступно наше приложение.

Порт и сервис нам очень нужны и важны, так как именно на них мы будем отправлять запросы, когда будем конфигурировать ingrss.

[.helm/templates/service.yaml](gitlab-rails-files/examples/example_1/.helm/templates/service.yaml)
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

При описании сервиса - нужно правильно указать селектор деплоймента иначе запросы не будут приходить на приложение.


#### Прописываем роутинг на Ingress


{{ TODO вводная и человеческое описание того, что происходит}}

Создадим ingress `.helm/templates/ingress.yaml `для нашего `kubernetes.io/ingress.class.`

_Тут есть коллизия терминов “ингресс” - в смысле приложение-балансировщик, которое работает в кластере и принимает входящие извне запросы и “Ингресс” в смысле объект, которую мы скармливаем кубернетесу, чтобы он настроил ингресс-приложение. Смиритесь, разбирайтесь по контексту._

[.helm/templates/ingress.yaml](gitlab-rails-files/examples/example_1/.helm/templates/ingress.yaml)
```yaml
  rules:
  - host: {{ .Values.global.ci_url }}
    http:
      paths:
      - path: /
        backend:
          serviceName: {{ .Chart.Name }}
          servicePort: 3000
```

Обратите внимание на параметр {{ .Values.global.ci_url }}. Данный параметр передается из файла .gitlab-ci.yml

[.gitlab-ci.yml](gitlab-rails-files/examples/example_1/.gitlab-ci.yml#L21)
```yaml
.base_deploy:
  script:
    - werf deploy
      --set "global.ci_url=example.com"
```

Описание файла `.gitlab-ci.yml` мы рассмотрим позже

Подобным образом, можно передавать и другие необходимые переменные.


#### Секретные переменные

Ruby on rails позволяет шифровать переменные. Это разрешает сохранять любые учетные данные аутентификации для сторонних сервисов напрямую в репозиторий, зашифрованный с помощью ключа в файле `config/master.key` или переменной среды `RAILS_MASTER_KEY`

Для хранения в репозитории паролей, файлов сертификатов и т.п., рекомендуется использовать подсистему работы с секретами werf.

Идея заключается в том, что конфиденциальные данные должны храниться в репозитории вместе с приложением в зашифрованном виде, и должны оставаться независимыми от какого-либо конкретного сервера.

Для работы нашего приложения потребуется `RAILS_MASTER_KEY` с помощью которого происходит шифрование секретов и оно необходимо в production окружении. Данную переменную окружения мы уже указали в деплойменете нашего приложения в строке 24. Все секретные данные нужно описывать в зашифрованом файле `.helm/secret-values.yaml`.


Подстановка значений из этого файла происходит при рендере шаблона, который также запускается при деплое.

Для создания зашифрованных данных нам необходимо сгенерировать ключ.

```bash
$ werf helm secret generate-secret-key
504a1a2b17042311681b1551aa0b8931
```

После генерации ключа - необходимо указать его в переменных окружения или в файле в корне нашего приложения

```bash
$ export WERF_SECRET_KEY=504a1a2b17042311681b1551aa0b8931
OR
$ echo 504a1a2b17042311681b1551aa0b8931 > .werf_secret_key
```

Теперь можем редактировать наши секретные переменные:

```bash
$ werf helm secret values edit .helm/secret-values.yaml
```

Откроется текстовый редактор который указан в `$EDITOR` или vim по умолчанию.
После шифрования наш файл будет выглядеть следующем образом:

[.helm/secret-values.yaml](gitlab-rails-files/examples/example_1/.helm/secret-values.yaml)
```yaml
rails:
  master_key: 100083f330adfb9e13fff74c9ab71b93ed77704aca3a0c607679336e099d48977d6565b314c06b8ad7aefc9d8d90629e92d851b573a89915ff036239de129d722ef5
```

Для декодирования секретных переменных необходимо добавить переменную `WERF_SECRET_KEY` в Variables  в для репозитория Settings - CI / CD. Гитлаб пробросит эту переменную в раннер и когда мы в нашем gitlab ci вызываем верфь деплой — werf увидит это значение и производит расшифровку секретных переменных.

![alt_text](/images/applications-guide/gitlab-rails/5.png "image_tooltip")



### Отладка конфигов



Нам понадобятся следующие файлы со структурой каталогов:


```
.helm (здесь мы будем описывать деплой)
├── templates (объекты kubernetes в виде шаблонов)
│   ├── deployment.yaml (основное приложение)
│   ├── ingress.yaml (описание для ingress)
│   └── service.yaml (сервис для приложения)
├── secret-values.yaml (файл с секретными переменными)
└── values.yaml (файл с переменными для параметризации шаблонов)
```

Подробнее читайте в [нашей статье](https://habr.com/ru/company/flant/blog/423239/) из серии про Helm.




```bash
$ werf helm render
```





##############################################################################################################
##############################################################################################################
##############################################################################################################
##############################################################################################################

## Выкат

TODO: мы тут берём и выкатываем с локальной тачки на прод в кубернетесы

{% offtopic title="Как вообще работает деплой" %}

Werf (по аналогии с helm) берет yaml шаблоны, которые описывают объекты Kubernetes, и генерирует из них общий манифест. Манифест отдается API Kubernetes, который на его основе внесет все необходимые изменения в кластер. Werf отслеживает как Kubernetes вносит изменения и сигнализирует о результатах в реальном времени. Все это благодаря встроенной в werf библиотеке [kubedog](https://github.com/flant/kubedog).

{% endofftopic %}



## Построение CI-процесса

После того как мы закончили со сборкой, которую можно производить локально, мы приступаем к базовой настройке CI/CD на базе Gitlab.

Начнем с того что добавим нашу сборку в CI с помощью .gitlab-ci.yml, который находится внутри корня проекта. Нюансы настройки CI в Gitlab можно найти [тут](https://docs.gitlab.com/ee/ci/).

Мы предлагаем простой флоу, который мы называем [fast and furious](https://docs.google.com/document/d/1a8VgQXQ6v7Ht6EJYwV2l4ozyMhy9TaytaQuA9Pt2AbI/edit#). Такой флоу позволит вам осуществлять быструю доставку ваших изменений в production согласно методологии GitOps и будут содержать два окружения, production и stage.

На стадии сборки мы будем собирать образ с помощью werf и загружать образ в registry, а затем на стадии деплоя собрать инструкции для kubernetes, чтобы он скачивал нужные образы и запускал их.

### Сборка в Gitlab CI

Для того, чтобы настроить CI-процесс создадим .gitlab-ci.yaml в корне репозитория.

Инициализируем werf перед запуском основной команды. Это необходимо делать перед каждым использованием werf поэтому мы вынесли в секцию `before_script`
Такой сложный путь с использованием multiwerf нужен для того, чтобы вам не надо было думать про обновление верфи и установке новых версий — вы просто указываете, что используете, например, use 1.1 stable и пребываете в уверенности, что у вас актуальная версия с закрытыми issues.

[.gitlab-ci.yml](gitlab-rails-files/examples/example_1/.gitlab-ci.yml#L8)
```yaml
before_script:
  - type multiwerf && source <(multiwerf use 1.1 stable)
  - type werf && source <(werf ci-env gitlab --verbose)
```

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


Многие из этих переменных интуитивно понятны, и содержат базовую информацию о том где находится проект, где находится его registry, информацию о коммитах. В рамках статьи нам хватит значений выставляемых по умолчанию.

Подробную информацию о конфигурации ci-env можно найти [тут](https://werf.io/documentation/reference/plugging_into_cicd/overview.html). Если вы используете GitLab CI/CD совместно с внешним docker registry (harbor, Docker Registry, Quay etc.), то в команду билда и пуша нужно добавлять его полный адрес (включая путь внутри registry), как это сделать можно узнать [тут](https://werf.io/documentation/cli/main/build_and_publish.html). И так же не забыть первой командой выполнить [docker login](https://docs.docker.com/engine/reference/commandline/login/).

Переменная [WERF_STAGES_STORAGE](https://ru.werf.io/documentation/reference/stages_and_images.html#%D1%85%D1%80%D0%B0%D0%BD%D0%B8%D0%BB%D0%B8%D1%89%D0%B5-%D1%81%D1%82%D0%B0%D0%B4%D0%B8%D0%B9) указывает где werf сохраняет стадии сборки (сборочный кэш). По умолчанию для GitLab CI/CD werf хранит стадии в docker registry, привязанном к GitLab CI/CD, по адресу `CI_REGISTRY_IMAGE/stages`. Используется распределённый режим работы werf, почитать о котором подробнее можно в статьях про [устройство стадий](https://ru.werf.io/documentation/reference/stages_and_images.html#%D1%85%D1%80%D0%B0%D0%BD%D0%B8%D0%BB%D0%B8%D1%89%D0%B5-%D1%81%D1%82%D0%B0%D0%B4%D0%B8%D0%B9) и [инструкции по переключению в распределённый режим](https://ru.werf.io/documentation/guides/switch_to_distributed_mode.html).

Дело в том что werf хранит стадии сборки раздельно, как раз для того чтобы мы могли не пересобирать весь образ, а только отдельные его части.

Плюс стадий в том, что они имеют собственный тэг, который представляет собой хэш, зависящий от содержимого нашего образа и истории изменений в git. Это позволяет избегать ненужных пересборок наших образов, но при этом обеспечивается изоляцию кеша между ветками. Подробнее об этом [читайте в руководстве](https://ru.werf.io/documentation/reference/stages_and_images.html#%D1%81%D1%82%D0%B0%D0%B4%D0%B8%D0%B8).

Следующие параметры тем кто работал с гитлаб уже должны быть знакомы.

**_tags_** - нужен для того чтобы выбрать наш раннер, на который мы навесили этот тэг. В данном случае наш gitlab-runner в Gitlab имеет тэг werf

```yaml
  tags:
    - werf
```

Теперь мы можем запушить наши изменения и увидеть что наша стадия успешно выполнилась.

![alt_text](/images/applications-guide/gitlab-rails/3.png "image_tooltip")


Лог в Gitlab будет выглядеть так же как и при локальной сборке, за исключением того что в конце мы увидим как werf пушит наш docker image в registry.

```
 │ ┌ Publishing image rails by stages-signature tag 72588513f1acf65c6436ab8d7e344964efd1fcf3 ...
 │ ├ Info
 │ │   images-repo: registry.example.com/alexey.chazov/example_1/rails
 │ │         image: registry.example.com/alexey.chazov/example_1/rails:72588513f1acf65c6436ab8d7 ↵
 │ │   e344964efd1fcf3ccefc07974decede
 │ └ Publishing image rails by stages-signature tag 72588513f1acf65c6436ab8d ... (13.34 seconds)
 └ ⛵ image rails (69.68 seconds)
```

### Деплой в Kubernetes


#### Деплой в Gitlab CI

Опишем деплой приложения в Kubernetes. Деплой будет осуществляться на два стенда: staging и production.

Выкат на два стенда отличается только параметрами, поэтому воспользуемся шаблонами. Опишем базовый деплой, который потом будем кастомизировать под стенды:

[.gitlab-ci.yml](gitlab-rails-files/examples/example_1/.gitlab-ci.yml#L21)
```yaml
.base_deploy: &base_deploy
  script:
    - werf deploy --stages-storage :local
  dependencies:
    - Build
  tags:
    - werf
```

Выкат, например, на Staging, будет выглядеть так:

 ```yaml
 Deploy to Stage:
   extends: .base_deploy
   stage: deploy
   environment:
     name: stage
   except:
     - schedules
   only:
     - merge_requests
   when: manual
```

Нет необходимости пробрасывать переменные окружения, создаваемые GitLab CI — этим занимается Werf. Достаточно только указать название стенда

```yaml
environment:
     name: stage
```

_Обратите внимание: домены каждого из стендов указываются в helm-шаблонах._

_Остальные настройки подробно описывать не будем, разобраться в них можно с [помощью документации Gitlab](https://docs.gitlab.com/ce/ci/yaml/)_

После описания стадий выката при создании Merge Request и будет доступна кнопка Deploy to Stage.

![alt_text](/images/applications-guide/gitlab-rails/6.png "image_tooltip")

Посмотреть статус выполнения pipeline можно в интерфейсе gitlab **CI / CD - Pipelines**

![alt_text](/images/applications-guide/gitlab-rails/7.png "image_tooltip")


Список всех окружений - доступен в меню **Operations - Environments**

![alt_text](/images/applications-guide/gitlab-rails/8.png "image_tooltip")

Из этого меню - можно так же быстро открыть приложение в браузере.

При переходе по ссылке - должно октрыться приложение.

![alt_text](/images/applications-guide/gitlab-rails/9.png "image_tooltip")

Проверить что приложение задеплоилось в кластер можно с помощью kubectl

```
root@kube-master ~ # kubectl get namespace
NAME                          STATUS   AGE
default                       Active   161d
example-1-production          Active   4m44s
example-1-stage               Active   3h2m

root@kube-master ~ # kubectl -n example-1-stage get po
NAME                        READY   STATUS    RESTARTS   AGE
example-1-9f6bd769f-rm8nz   1/1     Running   0          6m12s

root@kube-master ~ # kubectl -n example-1-stage get ingress
NAME        HOSTS                                           ADDRESS   PORTS   AGE
example-1   example-1-stage.kube.example.com                          80      6m18s
```
