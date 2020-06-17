---
title: Базовые настройки
sidebar: applications-guide
permalink: documentation/guides/applications-guide/template/020-basic.html
layout: guide
---



В первой главе мы покажем поэтапную сборку и деплой приложения без задействования внешних ресурсов таких как база данных и сборку ассетов.

Наше приложение будет состоять из одного docker образа собранного с помощью werf. Его единственной задачей будет вывод сообщения “hello world” по http.

В этом образе будет работать один основной процесс gunicorn, который запустит приложение через wsgi.

Управлять маршрутизацией запросов к приложению будет управлять Ingress в kubernetes кластере.

Мы реализуем два стенда: production и staging. В рамках hello world приложения мы предполагаем, что разработка ведётся локально, на вашем компьютере.

_В ближайшее время werf реализует удобные инструменты для локальной разработки, следите за обновлениями._


## Локальная сборка

Для того чтобы werf смогла начать работу с нашим приложением - необходимо в корне нашего репозитория создать файл werf.yaml в которым будут описаны инструкции по сборке. Для начала соберем образ локально не загружая его в registry чтобы разобраться с синтаксисом сборки.

С помощью werf можно собирать образы с используя Dockerfile или используя синтаксис, описанный в документации werf (мы называем этот синтаксис и движок, который этот синтаксис обрабатывает, stapel). Для лучшего погружения - соберем наш образ с помощью stapel.
Прежде всего нам необходимо собрать docker image с нашим приложением внутри. 

Клонируем наши исходники любым удобным способом. В нашем случае это:


```
git clone git@gitlab-example.com:article/tools.git
```

После, в корне склоненного проекта, создаём файл `werf.yaml`. Данный файл будет отвечать за сборку вашего приложения и он обязательно должен находиться в корне проекта. Исходный код находится в отдельной директории _django_, в данном случае это сделано просто для удобства, чтобы подразделить исходный код проекта от части связанной со сборкой.

Итак, начнём с самой главной секции нашего werf.yaml файла, которая должна присутствовать в нём **всегда**. Называется она [meta config section](https://werf.io/documentation/configuration/introduction.html#meta-config-section) и содержит всего два параметра.

werf.yaml:
```yaml
project: tools
configVersion: 1
```

**_project_** - поле, задающее имя для проекта, которым мы определяем связь всех docker images собираемых в данном проекте. Данное имя по умолчанию используется в имени helm релиза и имени namespace в которое будет выкатываться наше приложение. Данное имя не рекомендуется изменять (или подходить к таким изменениям с должным уровнем ответственности) так как после изменений уже имеющиеся ресурсы, которые выкачаны в кластер, не будут переименованы.

**_configVersion_** - в данном случае определяет версию синтаксиса используемую в `werf.yaml`.

После мы сразу переходим к следующей секции конфигурации, которая и будет для нас основной секцией для сборки - [image config section](https://werf.io/documentation/configuration/introduction.html#image-config-section). И чтобы werf понял что мы к ней перешли разделяем секции с помощью тройной черты.


```yaml
project: tools
configVersion: 1
---
image: django
from: python:3.6-stretch
```

**_image_** - поле задающее имя нашего docker image, с которым он будет запушен в registry. Должно быть уникально в рамках одного werf-файла.

**_from _** - задает имя базового образа который мы будем использовать при сборке. Задаем мы его точно так же, как бы мы это сделали в dockerfile. В примере используется {{imagename}} {{объяснение почкему такой}} 

Теперь встает вопрос о том как нам добавить исходный код приложения внутрь нашего docker image. И для этого мы можем использовать Git! И нам даже не придётся устанавливать его внутрь docker image.

**_git_**, на наш взгляд это самый правильный способ добавления ваших исходников внутрь docker image, хотя существуют и другие. Его преимущество в том что он именно клонирует, и в дальнейшем накатывает коммитами изменения в тот исходный код что мы добавили внутрь нашего docker image, а не просто копирует файлы. Вскоре мы узнаем зачем это нужно.

```yaml
project: tools
configVersion: 1
---
image: django
from: python:3.6-stretch
git:
- add: /
  to: /app
```

Werf подразумевает что ваша сборка будет происходить внутри директории склонированного git репозитория. Потому мы списком можем указывать директории и файлы относительно корня репозитория которые нам нужно добавить внутрь image.

`add: /` - та директория которую мы хотим добавить внутрь docker image, мы указываем, что это весь наш репозиторий

`to: /app` - то куда мы клонируем наш репозиторий внутри docker image. Важно заметить что директорию назначения werf создаст сам.

 Есть возможность даже добавлять внешние репозитории внутрь проекта не прибегая к предварительному клонированию, как это сделать можно узнать [тут](https://werf.io/documentation/configuration/stapel_image/git_directive.html), но мы не рекомендуем такой подход.

Следующим этапом необходимо описать правила сборки для приложения. Werf позволяет кэшировать сборку образа подобно слоям в docker, только с явным набором инструкций необходимых данном кэше. Этот сборочный этап - называется стадия. Мы рассмотрим более подробно возможности стадий в с следующих главах.

Для текущего приложения опишем 2 стадии в которых сначала устанавливаем необходимые зависимости для возможности сборки приложения а потом - непосредственно собираем приложение.

Команды описывающие сборку можно описывать в ansible формате или shell командами.
Добавим в `werf.yaml` следующий блок используя ansible синтаксис:

```yaml
ansible:
  install:
  - name: Install requirements
    apt:
      name:
      - locales
      update_cache: yes
  - name: Set timezone
    timezone:
      name: "Etc/UTC"
  - name: Generate locale
    locale_gen:
      name: en_US.UTF-8
      state: present
  setup:
  - name: Install python requirements
    pip:
      requirements: /usr/src/app/requirements.txt
      executable: pip3.6
```

Полный список поддерживаемых модулей ansible в werf можно найти [тут](https://werf.io/documentation/configuration/stapel_image/assembly_instructions.html#supported-modules).

Не забыв [установить werf](https://werf.io/documentation/guides/installation.html) локально, запускаем сборку с помощью [werf build](https://werf.io/documentation/cli/main/build.html)!

```bash
$  werf build --stages-storage :local
```

![alt_text](images/-0.gif "image_tooltip")

Вот и всё, наша сборка успешно завершилась. К слову если сборка падает и вы хотите изнутри контейнера её подебажить вручную, то вы можете добавить в команду сборки флаги:

```yaml
--introspect-before-error
```

или

```yaml
--introspect-error
```

Которые при падении сборки на одном из шагов автоматически откроют вам shell в контейнер, перед исполнением проблемной инструкции или после.

В конце werf отдал информацию о готовом image:

![alt_text](images/-1.png "image_tooltip")

Теперь его можно запустить локально используя image_id просто с помощью docker.
Либо вместо этого использовать [werf run](https://werf.io/documentation/cli/main/run.html):


```bash
werf run --stages-storage :local --docker-options="-d -p 8080:8080 --restart=always" -- python manage.py runserver
```

Первая часть команды очень похожа на build, а во второй мы задаем [параметры](https://docs.docker.com/engine/reference/run/) docker и через двойную черту команду с которой хотим запустить наш image.

Небольшое пояснение про `--stages-storage :local `который мы использовали и при сборке и при запуске приложения. Данный параметр указывает на то где werf хранить стадии сборки. На момент написания статьи это возможно только локально, но в ближайшее время появится возможность сохранять их в registry.

Теперь наше приложение доступно локально на порту 8080:

![alt_text](images/-2.png "image_tooltip")

На этом часть с локальным использованием werf мы завершаем и переходим к той части для которой werf создавался, использовании его в CI.

## Построение CI-процесса

После того как мы закончили со сборкой, которую можно производить локально, мы приступаем к базовой настройке CI/CD на базе Gitlab.

Начнем с того что добавим нашу сборку в CI с помощью .gitlab-ci.yml, который находится внутри корня проекта. Нюансы настройки CI в Gitlab можно найти [тут](https://docs.gitlab.com/ee/ci/).

Мы предлагаем простой флоу, который мы называем [fast and furious](https://docs.google.com/document/d/1a8VgQXQ6v7Ht6EJYwV2l4ozyMhy9TaytaQuA9Pt2AbI/edit#). Такой флоу позволит вам осуществлять быструю доставку ваших изменений в production согласно методологии GitOps и будут содержать два окружения, production и stage.

На стадии сборки мы будем собирать образ с помощью werf и загружать образ в registry, а затем на стадии деплоя собрать инструкции для kubernetes, чтобы он скачивал нужные образы и запускал их.

### Сборка в Gitlab CI

Для того, чтобы настроить CI-процесс создадим .gitlab-ci.yaml в корне репозитория.

Инициализируем werf перед запуском основной команды. Это необходимо делать перед каждым использованием werf поэтому мы вынесли в секцию `before_script`
Такой сложный путь с использованием multiwerf нужен для того, чтобы вам не надо было думать про обновление верфи и установке новых версий — вы просто указываете, что используете, например, use 1.1 stable и пребываете в уверенности, что у вас актуальная версия с закрытыми issues.

```yaml
before_script:
  - type multiwerf && source <(multiwerf use 1.1 stable)
  - type werf && source <(werf ci-env gitlab --verbose)
```

`werf ci-env gitlab --verbose` - готовит наш werf для работы в Gitlab, выставляя для этого все необходимые переменные.
Пример переменных автоматически выставляемых этой командой:

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


Многие из этих переменных интуитивно понятны, и содержат базовую информацию о том где находится проект, где находится его registry, информацию о коммитах. \
Подробную информацию о конфигурации ci-env можно найти [тут](https://werf.io/documentation/reference/plugging_into_cicd/overview.html). От себя лишь хочется добавить, что если вы используете совместно с Gitlab внешний registry (harbor,Docker Registry,Quay etc.), то в команду билда и пуша нужно добавлять его полный адрес (включая путь внутри registry), как это сделать можно узнать [тут](https://werf.io/documentation/cli/main/build_and_publish.html). И так же не забыть первой командой выполнить [docker login](https://docs.docker.com/engine/reference/commandline/login/).

В рамках статьи нам хватит значений выставляемых по умолчанию.

Переменная [WERF_STAGES_STORAGE](https://ru.werf.io/documentation/reference/stages_and_images.html#%D1%85%D1%80%D0%B0%D0%BD%D0%B8%D0%BB%D0%B8%D1%89%D0%B5-%D1%81%D1%82%D0%B0%D0%B4%D0%B8%D0%B9) указывает где werf сохраняет свой кэш (стадии сборки) У werf есть опция распределенной сборки, про которую вы можете прочитать в нашей статье, в текущем примере мы сделаем по-простому и сделаем сборку на одном узле в один момент времени.


```yaml
variables:
    WERF_STAGES_STORAGE: ":local"
```
Дело в том что werf хранит стадии сборки раздельно, как раз для того чтобы мы могли не пересобирать весь образ, а только отдельные его части.

Плюс стадий в том, что они имеют собственный тэг, который представляет собой хэш содержимого нашего образа. Тем самым позволяя полностью избегать не нужных пересборок наших образов. Если вы собираете ваше приложение в разных ветках, и исходный код в них различается только конфигами которые используются для генерации статики на последней стадии. То при сборке образа одинаковые стадии пересобираться не будут, будут использованы уже собранные стадии из соседней ветки. Тем самым мы резко снижаем время доставки кода.

Основная команда на текущий момент - это werf build-and-publish, которая запускает сборку и публикацию в registry на gitlab runner с тегом werf для любой ветки. Путь до registry и другие параметры беруться верфью автоматически их переменных окружения gitlab ci.

```yaml
Build:
  stage: build
  script:
    - werf build-and-publish
  tags:
    - werf
```

Если вы всё правильно сделали и корректно настроен registry и gitlab ci — вы увидите собранный образ в registry. При использовании registry от gitlab — собранный образ можно увидеть через веб-интерфейс гитлаба.

Следующие параметры тем кто работал с гитлаб уже должны быть знакомы.

**_tags_** - нужен для того чтобы выбрать наш раннер, на который мы навесили этот тэг. В данном случае наш gitlab-runner в Gitlab имеет тэг werf

```yaml
  tags:
    - werf
```

Теперь мы можем запушить наши изменения и увидеть что наша стадия успешно выполнилась.

![alt_text](images/-3.png "image_tooltip")


Лог в Gitlab будет выглядеть так же как и при локальной сборке, за исключением того что в конце мы увидим как werf пушит наш docker image в registry.

```
207 │ ┌ Publishing image {{node}} by stages-signature tag c905b748cb9647a03476893941837bf79910ab09e ...
208 │ ├ Info
209 │ │   images-repo: registry.gitlab-example.com/{{chat/node}}
210 │ │        image: registry.gitlab-example.com/{{chat/node}}:c905b748cb9647a03476893941 ↵
211 │ │   837bf79910ab09ef5878037592a45d
212 │ └ Publishing image {{node}} by stages-signature tag c905b748cb9647a0347689394 ... (14.90 seconds)
213 └ ⛵ image {{node}} (73.44 seconds)
214 Running time 73.47 seconds
218 Job succeeded
```
По умолчанию у werf выставлена стартегия тэгирования docker images `stages-signature`, которая тэгирует ваши docker images на основе контента, который они содержат. Это называется [content based tagging](https://werf.io/documentation/reference/publish_process.html#content-based-tagging). Суть его в том что тэг является хэшсуммой всех стадий сборки. Такой подход избавляет нас от лишних пересборок.
### Деплой в Kubernetes

Werf использует встроенный Helm для применения конфигурации в Kubernetes. Для описания объектов Kubernetes werf использует конфигурационные файлы Helm: шаблоны и файлы с параметрами (например, values.yaml). Помимо этого, werf поддерживает дополнительные файлы, такие как файлы c секретами и с секретными значениями (например secret-values.yaml), а также дополнительные Go-шаблоны для интеграции собранных образов.

Werf (по аналогии с helm) берет yaml шаблоны, которые описывают объекты Kubernetes, и генерирует из них общий манифест. Манифест отдается API Kubernetes, который на его основе внесет все необходимые изменения в кластер. Werf отслеживает как Kubernetes вносит изменения и сигнализирует о результатах в реальном времени. Все это благодаря встроенной в werf библиотеке [kubedog](https://github.com/flant/kubedog).

Внутри Werf доступны команды для работы с Helm, например можно проверить как сгенерируется общий манифест в результате работы werf с шаблонами:

```bash
$ werf helm render
```

Аналогично, доступны команды [helm list](https://werf.io/documentation/cli/management/helm/list.html) и другие.

#### Общее про хельм-конфиги

На сегодняшний день [Helm](https://helm.sh/) один из самых удобных способов которым вы можете описать свой deploy в Kubernetes. Кроме возможности установки готовых чартов с приложениями прямиком из репозитория, где вы можете введя одну команду, развернуть себе готовый Redis, Postgres, Rabbitmq прямиком в Kubernetes, вы также можете использовать Helm для разработки собственных чартов с удобным синтаксисом для шаблонизации выката ваших приложений.

Потому для werf это был очевидный выбор использовать такую технологию.

Мы не будем вдаваться в подробности разработки yaml манифестов с помощью Helm для Kubernetes. Осветим лишь отдельные её части, которые касаются данного приложения и werf в целом. Если у вас есть вопросы о том как именно описываются объекты Kubernetes, советуем посетить страницы документации по Kubernetes с его [концептами](https://kubernetes.io/ru/docs/concepts/) и страницы документации по разработке [шаблонов](https://helm.sh/docs/chart_template_guide/) в Helm.

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

![alt_text](images/-4.png "image_tooltip")

#### Описание приложения в хельме

Для работы нашего приложения в среде Kubernetes понадобится описать сущности Deployment, Service, завернуть трафик на приложение, донастроив роутинг в кластере с помощью сущности Ingress. И не забыть создать отдельную сущность Secret, которая позволит нашему kubernetes пулить собранные образа из registry.

##### Запуск контейнера

Начнем с описания deployment.yaml

<details><summary>deployment.yaml</summary>
<p>

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}-django
spec:
  revisionHistoryLimit: 3
  strategy:
    type: RollingUpdate
  replicas: 1
  selector:
    matchLabels:
      app: {{ .Chart.Name }}-django
  template:
    metadata:
      labels:
        app: {{ .Chart.Name }}-django
    spec:
      imagePullSecrets:
      - name: "registrysecret"
      containers:
      - name: {{ .Chart.Name }}-django
{{ tuple "django" . | include "werf_container_image" | indent 8 }}
        workingDir: /app
        command: ['gunicorn', 'tools.wsgi:application', '--bind', '0.0.0.0:80', '--access-logfile', '-', '--log-level', 'debug']
        ports:
        - containerPort: 80
          protocol: TCP
        env:
{{ tuple "django" . | include "werf_container_env" | indent 8 }}
```
</p>
</details>


Коснусь только шаблонизированных параметров. Значение остальных параметров можно найти в документации [Kubernetes](https://kubernetes.io/docs/concepts/).

`{{ .Chart.Name }}` - значение для данного параметра берётся из файла werf.yaml из поля **_project_**


werf.yaml:

```yaml
project: tools
configVersion: 1
```
Далее мы указываем имя сектрета в котором мы будем хранить данные для подключение к нашему registry, где хранятся наши образа.

```yaml
      imagePullSecrets:
      - name: "registrysecret"
```
О том как его создать мы опишем в конце главы.



Шаблон ниже отвечает за то чтобы вставить информацию касающуюся местонахождения нашего doсker image в registry, чтобы kubernetes знал откуда его скачать. А также политику пула этого образа.

```yaml
{{ tuple "django" . | include "werf_container_image" | indent 8 }}
```
 И в итоге эта строка будет заменена helm’ом на это:


```yaml
   image: registry.gitlab-example.com/tools/django:6e3af42b741da90f2bc674e5646a87ad6b81d14c531cc89ef4450585   
   imagePullPolicy: IfNotPresent
```

Замену производит сам werf из собственных переменных. Изменять эту конструкцию нужно только в двух местах:
1. Рядом в первой части “django”  -  это название вашего docker image, которые мы указывали в werf.yaml в поле **image**, когда описывали сборку.

2. Intent 8 - параметр указывает какое количество пробелов вставить перед блоком, делаем мы это чтобы не нарушить синтаксис yaml, где пробелы(отступы) играют важную разделительную роль.  \
При разработке особенно важно учитывать что yaml не воспринимает табуляцию **только пробелы**!

```yaml
{{ tuple "django" . | include "werf_container_env" | indent 8 }}
```
Одна из самых главных строк, отвечает непосредственно за то какую команду запустить при запуске приложения.

```yaml
        ports:
        - containerPort: 80
          protocol: TCP
```
Блок отвечающий за то какие порты необходимо сделать доступными снаружи контейнера, и по какому протоколу.

```yaml
{{ tuple "django" . | include "werf_container_env" | indent 8 }}
```
Этот шаблон позволяет werf работать с переменными.
Его назначение подробно описано в следующей главе.


Теперь, как мы и обещали перейдем к созданию сущности Secret, которая будет содержать доступы до images registry.
Вы можете использовать команду kubectl, из кластера или у себя на личном компьютере даже если он не имеет доступ к кластеру (он нам не понадобится).
Вы можете запустить следующую команду:
```bash
kubectl create secret docker-registry regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email> --dry-run=true -o yaml
```
В команде вы указываете данные пользователя для подключения и затем получаете такой вывод:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: registrysecret
  creationTimestamp: null
type: kubernetes.io/dockerconfigjson

data:
  .dockerconfigjson: eyJhdXRocyI6eyJyZWdpc3RyeS5leGFtcGxlLmNvbSI6eyJ1c2VybmFtZSI6InVzZXIiLCJwYXNzd29yZCI6InF3ZXJ0eSIsImVtYWlsIjoiZXhhbXBsZUBnbWFpbC5ydSIsImF1dGgiOiJkWE5sY2pweGQyVnlkSGs9In19fQ==

```
Команда сформировала готовый секрет и отдала его вам, зашифровав данные в base64. Сработало это благодаря флагам `--dry-run=true` и `-o yaml`, первый флаг говорит о том что мы хотим сымитировать создание сущности в кластере без доступа к нему, а второй о том что мы хотим видеть наши данные в формате `yaml`

Теперь вам осталось только создать отдельный файл Secret.yaml и положить в него содержимое которое выдала вам команда, предварительно удалив строку `creationTimestamp: null`.

P.S. Настоятельно не рекомендуем хранить данные подключения в сыром виде в котором нам выдала команда, о том каким образом можно зашифровать данные с помощью werf будет показано в главе [Секретные переменные](####секретные-переменные).

##### Переменные окружения

Для корректной работы нашего приложения ему нужно узнать переменные окружения.
По умолчанию Django запускает и без них, но это не позводит нам конфигурировать приложение на лету.
Поэтому в настройках мы изменим некоторые параметры, и будем брать их из переменных, например, `DEBUG` и `SECRET_KEY`.


И эти переменные можно параметризовать с помощью файла `values.yaml`.

Так например, мы пробросим значение переменной DEBUG в наш контейнер из `values.yaml`

```yaml
app:
  debug:
    stage: "1"
    production: "0"
```
И теперь добавляем переменную в наш Deployment.
```yaml
          - name: DEBUG
            value: {{ pluck .Values.global.env .Values.app.debug | first | default .Values.app.debug._default | quote }}
```
Конструкция указывает на то что в зависимости от значения .Values.global.env мы будем подставлять первое совпадающее значение из .Values.app.debug

Werf устанавливает значение .Values.global.env в зависимости от названия окружения указанного в .gitlab-ci.yml в стадии деплоя.

Теперь перейдем к описанию шаблона из предыдущей главы:


```yaml
        env:
{{ tuple "django" . | include "werf_container_env" | indent 8 }}
```
Werf закрывает ряд вопросов, связанных с перевыкатом контейнеров с помощью конструкции  [werf_container_env](https://ru.werf.io/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_env). Она возвращает блок с переменной окружения DOCKER_IMAGE_ID контейнера пода. Значение переменной будет установлено только если .Values.global.werf.is_branch=true, т.к. в этом случае Docker-образ для соответствующего имени и тега может быть обновлен, а имя и тег останутся неизменными. Значение переменной DOCKER_IMAGE_ID содержит новый ID Docker-образа, что вынуждает Kubernetes обновить объект.

Важно учесть что данный параметр не подставляет ничего при использовании стратегии тэгирования `stages-signature`, но мы настоятельно рекомендуем добавить его внутрь манифеста для удобства интеграции будущих обновлений werf.

Аналогично можно пробросить секретные переменные (пароли и т.п.) и у Верфи есть специальный механизм для этого. Но к этому вопросу мы вернёмся позже.


##### Логгирование

При запуске приложения в kubernetes необходимо логи отправлять в stdout и stderr - это необходимо для простого сбора логов например через `filebeat`, а так же чтобы не разростались docker образы запущенных приложений. По умолчанию Django не логирует ошибки и критичные логи, а отправляет на email администратора.

Мы предлагаем:
1. Писать все логи в stdout контейнера и чтобы оттуда их собирал сторонний сервис.

Чтобы все логи отправлялись в stdout, его нужно сконфигурировать в коде приложения. Добавьте в файл настроек словарь:

```yaml
LOGGING = {
    'version': 1,
    'disable_existing_loggers': False,
    'handlers': {
        'console': {
            'class': 'logging.StreamHandler',
        },
    },
    'root': {
        'handlers': ['console'],
        'level': 'WARNING',
    },
}
```

[Подбробно про логирование](https://docs.djangoproject.com/en/3.0/topics/logging/)

2. Ограничить их количество в stdout с помощью настройки для Docker в /etc/docker/daemon.json

```json
{
        "log-driver": "json-file",
        "log-opts": {
                "max-file": "5",
                "max-size": "10m"
        }
}
```
В общей сложности конструкция выше понятна, но если вы хотите разобрать её подробнее вы можете обратиться к официальной [документации](https://docs.docker.com/config/containers/logging/configure/).

##### Направление трафика на приложение

Нам надо будет пробить порт у пода, сервиса и настроить Ingress, который выступает у нас в качестве балансера.

Если вы мало работали с Kubernetes — эта часть может вызвать у вас много проблем. Большинство тех, кто начинает работать с Kubernetes по невнимательности допускают ошибки при конфигурировании labels и затем занимаются долгой и мучительной отладкой.

{{TODO: тут бы дать какую-то подсказку, как человеку пройти через это и не поседеть, если у него рядом нет ментора, который ткнёт ему в опечатку}}

###### Проброс портов

Для того чтобы мы смогли общаться с нашим приложением извне необходимо привязать к нашему deployment объект Service.

В наш service.yaml нужно добавить:

```yaml
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Chart.Name }}-django
spec:
  selector:
    app: {{ .Chart.Name }}-django
  clusterIP: None
  ports:
  - name: http
    port: 80
    protocol: TCP
```
Обязательно нужно указывать порты, на котором будет слушать наше приложение внутри контейнера. И в Service, как указано выше и в Deployment:

```yaml
        ports:
        - containerPort: 80
          protocol: TCP
```

Сама же привязка к deployment происходит с помощью блока **selector:**


```yaml
  selector:
    app: {{ .Chart.Name }}-django
```


Внутри селектора у нас указан лэйбл `app: {{ .Chart.Name }}-django` он должен полностью совпадать с блоком `labels` в Deployment который мы описывали в главах выше:



```yaml
  template:
    metadata:
      labels:
        app: {{ .Chart.Name }}-django
```


Иначе Kubernetes не поймет на какой именно под или совокупность подов Service указывать. Это важно еще и из-за того что ip адреса подов попадают в DNS Kubernetes под именем сервиса, что позволяет нам обращаться к поду с нашим приложения просто по имени сервиса.

Полная запись для пода в нашем случае будет выглядеть так:
`tools-django.stage.svc.cluster.local` и расшифровывается так - `имя_сервиса.имя_неймспейса.svc.cluster.local` - неизменная часть это стандартный корневой домен Kubernetes.

Интересно то что поды находящиеся внутри одного неймспейса могут обращаться друг к другу просто по имени сервиса.

Подробнее о том как работать с сервисами можно узнать в [документации](connect-applications-service).


###### Роутинг на Ingress
Теперь мы можем передать nginx ingress имя сервиса на который нужно проксировать запросы извне. 
<details><summary>ingress.yaml</summary>
<p>

```yaml
---
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: nginx
  name: {{ .Chart.Name }}-django
spec:
  rules:
  - host: {{ .Values.global.ci_url }}
    http:
      paths:
      - backend:
          serviceName: {{ .Chart.Name }}-django
          servicePort: 80
        path: /
```
</p>
</details>

Настройка роутинга происходит непосредственно в блоке `rules:`, где мы можем описать правила по которму трафик будет попадать в наше приложение.

`- host: {{ .Values.global.ci_url }}` - в данном поле мы описываем тот домен на который конечный пользователь будет обращаться чтобы попасть в наше приложение. Можно сказать что это точка входа в наше приложение.

`paths:` - отвечает за настройку путей внутри нашего домена. И принимает в себя список из конфигураций этих путей.
Далее мы прямо описываем что все запросы попадающие на корень `path: /`, мы отправляем на backend, которым выступает наш сервис:
```yaml
      - backend:
          serviceName: {{ .Chart.Name }}-django
          servicePort: 80
        path: /
```
Имя сервиса и его порт должны полностью совпадать с теми что мы описывали в сущности Service.
Удобство в том что описаний таких бэкендов может быть множество. И вы можете на одном домене по разным путям направлять трафик в разные приложения. Как это делать будет описано в последующих главах.
```

Обратите внимание на параметр {{ .Values.global.ci_url }}. Данный параметр передается из файла .gitlab-ci.yml

```yaml
.base_deploy:
  script:
    - werf deploy
      --set "global.ci_url=example.com"
```

Подобным образом, можно передавать и другие необходимые переменные.


#### Секретные переменные

Мы уже рассказывали о том как использовать обычные переменные в нашем СI забирая их напрямую из values.yaml. Суть работы с секретными переменными абсолютно та же, единственное что в репозитории они будут храниться в зашифрованном виде.

Потому для хранения в репозитории паролей, файлов сертификатов и т.п., рекомендуется использовать подсистему работы с секретами werf.

Идея заключается в том, что конфиденциальные данные должны храниться в репозитории вместе с приложением, и должны оставаться независимыми от какого-либо конкретного сервера.


Для этого в werf существует инструмент [helm secret](https://werf.io/documentation/reference/deploy_process/working_with_secrets.html). Чтобы воспользоваться шифрованием нам сначала нужно создать ключ, сделать это можно так: 

```bash
$ werf helm secret generate-secret-key
ad747845284fea7135dca84bde9cff8e
$ export WERF_SECRET_KEY=ad747845284fea7135dca84bde9cff8e
```

После того как мы сгенерировали ключ, добавим его в переменные окружения у себя локально.

Секретные данные мы можем добавить создав рядом с values.yaml файл secret-values.yaml

Теперь использовав команду:


```bash
$ werf helm secret values edit ./helm/secret-values.yaml
```


Откроется текстовый редактор по-умолчанию, где мы сможем добавить наши секретные данные как обычно:


```yaml
app:
  s3:
    access_key:
      _default: bNGXXCF1GF
    secret_key:
      _default: zpThy4kGeqMNSuF2gyw48cOKJMvZqtrTswAQ
```


После того как вы закроете редактор, werf зашифрует их и secret-values.yaml будет выглядеть так:

И вы сможете добавить их в переменные окружения в Deployment точно так же как делали это с обычными переменными. Главное это не забыть добавить ваш WERF_SECRET_KEY в переменные репозитория гитлаба, найти их можно тут Settings -> CI/CD -> Variables. Настройки репозитория доступны только участникам репозитория с ролью выше Administrator, потому никто кроме доверенных лиц не сможет получить наш ключ. А werf при деплое нашего приложения сможет спокойно получить ключ для расшифровки наших переменных.

#### Деплой в Gitlab CI

Теперь мы наконец приступаем к описанию стадии выката. Потому продолжаем нашу работу в gitlab-ci.yml.

Мы уже решили, что у нас будет два окружения, потому под каждое из них мы должны описать свою стадию, но в общей сложности они будут отличаться только параметрами, потому мы напишем для них небольшой шаблон:

```yaml
.base_deploy: &base_deploy
  script:
    - werf deploy --stages-storage :local 
      --set "global.ci_url=$(cut -d / -f 3 <<< $CI_ENVIRONMENT_URL)"
  dependencies:
    - Build
  tags:
    - werf
```

Скрипт стадий выката отличается от сборки всего одной командой:

```yaml
    - werf deploy --stages-storage :local
      --set "global.ci_url=$(cut -d / -f 3 <<< $CI_ENVIRONMENT_URL)"
```

И тут назревает вполне логичный вопрос.

Как werf понимает куда нужно будет деплоить и каким образом? На это есть два ответа.

Первый из них вы уже видели и заключается он в команде `werf ci-env` которая берёт нужные переменные прямиком из pipeline Gitlab - и в данном случае ту что касается названия окружения.

А второй это описание стадий выката в нашем gitlab-ci.yml:

```yaml
Deploy to Stage:
  extends: .base_deploy
  stage: deploy
  environment:
    name: stage
    url: https://stage.example.com
  only:
    - merge_requests
  when: manual

Deploy to Production:
  extends: .base_deploy
  stage: deploy
  environment:
    name: production
    url: http://example.com
  only:
    - master
```

Описание деплоя содержит в себе немного. Скрипт, указание принадлежности к стадии **deploy**, которую мы описывали в начале gitlab-ci.yml, и **dependencies** что означает что стадия не может быть запущена без успешного завершения стадии **Build**. Также мы указали с помощью **only**, ветку _master_, что означает что стадия будет доступна только из этой ветки. **environment** указали потому что werf необходимо понимать в каком окружении он работает. В дальнейшем мы покажем, как создать CI для нескольких окружений. Остальные параметры вам уже известны.

И что не мало важно **url** указанный прямо в стадии. 
1. Это добавляет в MR и pipeline дополнительную кнопку по которой мы можем сразу попасть в наше приложение. Что добавляет удобства.
2. С помощью конструкции `--set "global.ci_url=$(cut -d / -f 3 <<< $CI_ENVIRONMENT_URL)"` мы добавляем адрес в глобальные переменные проекта и затем можем например использовать его динамически в качестве главного домена в нашей сущности Ingress:
```yaml
      - host: {{ .Values.global.ci_url }}
```
По умолчанию деплой будет происходить в namespace состоящий из имени проекта задаваемого в `werf.yaml` и имени окружения задаваемого в `.gitlab-ci.yml` куда мы деплоим наше приложение.

Ну а теперь достаточно создать Merge Request и нам будет доступна кнопка Deploy to Stage.

![alt_text](images/-6.png "image_tooltip")

Посмотреть статус выполнения pipeline можно в интерфейсе gitlab **CI / CD - Pipelines**

![alt_text](images/-7.png "image_tooltip")


Список всех окружений - доступен в меню **Operations - Environments**

![alt_text](images/-8.png "image_tooltip")

Из этого меню - можно так же быстро открыть приложение в браузере.

{{И тут в итоге должна быть картинка как аппка задеплоилась и объяснение картинки}}

