---
title: Гайд по использованию Java Springboot + GitLab + Werf
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-java-springboot.html
author: Евгений Ермонин <evgeny.ermonin@flant.com>
layout: guide
toc: false
author_team: "charlie"
author_name: "Евгений Ермонин"
ci: "gitlab"
language: "java"
framework: "springboot"
is_compiled: 1
package_managers_possible:
 - maven
 - gradle
package_managers_chosen: "maven"
unit_tests_possible:
 - junit
 - testng
unit_tests_chosen: "junit"
assets_generator_possible:
 - webpack
 - maven frontend plugin
assets_generator_chosen: "webpack"
---

<a name="preparing" />

# Подготовка

Рассмотрим разные способы которые помогут собрать Java-приложение на примере springboot и запустить его в kubernetes кластере.

Предполагается что читатель имеет базовые знания в разработке на Java а также немного знаком с Gitlab CI и примитивами kubernetes, либо готов во всём этом разобраться самостоятельно. Мы постараемся предоставить все ссылки на необходимые ресурсы, если потребуется приобрести какие то новые знания.  

Собирать приложения будем с помощью werf. Данный инструмент работает в Linux MacOS и Windows, инструкция по [установке](https://ru.werf.io/documentation/guides/installation.html) находится на официальном [сайте](https://ru.werf.io/). В качестве примера - также приложим Docker файлы.

Для иллюстрации действий в данной статье - создан репозиторий с исходным кодом, в котором находятся несколько простых приложений. Мы постараемся подготовить примеры чтобы они запускались на вашем стенде и постараемся подсказать, как отлаживать возможные проблемы при вашей самостоятельной работе.

<a name="preparing-app" />

## Подготовка приложения

Наилучшим образом приложения будут работать в Kubernetes - если они соответствуют [12 факторам heroku](https://12factor.net/). Благодаря этому - у нас в kubernetes работают stateless приложения, которые не зависят от среды. Это важно, так как кластер может самостоятельно переносить приложения с одного узла на другой, заниматься масштабированием и т.п. — и мы не указываем, где конкретно запускать приложение, а лишь формируем правила, на основании которого кластер принимает свои собственные решения.

Договоримся что наши приложения соответствуют этим требованиям. На хабре уже было описание данного подхода, вы можете почитать про него например [тут](https://12factor.net/).

<a name="preparing-env" />

## Подготовка среды

Для того, чтобы пройти по этому гайду, необходимо, чтобы

*   У вас был работающий и настроенный Kubernetes кластер
*   Код приложения находился в Gitlab
*   Был настроен Gitlab CI, подняты и подключены к нему раннеры

Для пользователя под которым будет производиться запуск runner-а - нужно установить multiwerf - данная утилита позволяет переключаться между версиями werf и автоматически обновлять его. Инструкция по установке - доступна по [ссылке](https://ru.werf.io/documentation/guides/installation.html#installing-multiwerf).

Для автоматического выбора актуальной версии werf в канале stable, релиз 1.1 выполним следующую  команду:

```
. $(multiwerf use 1.1 stable --as-file)
```

Перед деплоем нашего приложения необходимо убедиться что наша инфраструктура готова к тому чтобы использовать werf. Используя [инструкцию](https://ru.werf.io/documentation/guides/gitlab_ci_cd_integration.html#%D0%BD%D0%B0%D1%81%D1%82%D1%80%D0%BE%D0%B9%D0%BA%D0%B0-runner) по подготовке к использованию Werf в Gitlab CI, вам нужно убедиться что все следующие пункты выполнены:

*   Развернут отдельный сервер с сетевой доступностью до мастер ноды Kubernetes.
*   На данном сервере установлен gitlab-runner.
*   Gitlab-runner подключен к нашему Gitlab с тегом werf в режиме shell executor. 
*   Ранеры включены и активны для репозитория с нашим приложением.
*   Для пользователя, которого использует gitlab-runner и под которым запускается сборка и деплой, установлен kubectl и добавлен конфигурационный файл для подключения к kubernetes.
*   Для gitlab включен и настроен gitlab registry
*   Gitlab-runner имеет доступ к API kubernetes и запускается по тегу werf  

<a href="hello-world" />

# Hello World приложение

В первой главе мы покажем поэтапную сборку и деплой приложения без задействования внешних ресурсов таких как база данных и сборку ассетов.

Наше приложение будет состоять из одного docker образа собранного с помощью werf. Его единственной задачей будет вывод сообщения “hello world” по http.

В нашем случае будет работать процесс java, исполняющий собранный jar отдающий hello world по http.

Управлять маршрутизацией запросов к приложению будет Ingress в kubernetes кластере.

Мы реализуем два стенда: production и staging. В рамках hello world приложения мы предполагаем, что разработка ведётся локально, на вашем компьютере.

_В ближайшее время werf реализует удобные инструменты для локальной разработки, следите за обновлениями._

<a href="hello-world-local" />

## Локальная сборка

Поскольку собирать мы будем spring-фреймворк - для скачивания шаблона приложения перейдем на start.spring.io. Для простоты оставляем все поля как есть, справа добавляем в dependencies только "Spring Web" и нажмем generate. Разархивируем полученный архив - получим готовую структуру папок и нужные нам файлы для того чтобы описать простейшее приложение.
tree:

```
├── HELP.md
├── mvnw
├── mvnw.cmd
├── pom.xml
└── src
    ├── main
    │   ├── java
    │   │   └── com
    │   │       └── example
    │   │           └── demo
    │   │               └── DemoApplication.java
    │   └── resources
    │       └── application.properties
    └── test
        └── java
            └── com
                └── example
                    └── demo
                        └── DemoApplicationTests.java

12 directories, 7 files
```
[Можно посмотреть в репозитории](gitlab-java-springboot-files/00-demo/)


pom.xml у нас сгенерирован автоматически, в нем правки не нужны.
application.properties на данном этапе так же оставим пустым.
А вот в DemoApplication.java чуть допишем код, чтобы приложение по http отвечало Hello World:

```java
...
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;


@SpringBootApplication
@RestController
public class DemoApplication {

        @RequestMapping("/")
        public String home() {
                return "Hello World";
        }
...
```
[DemoApplication.java](gitlab-java-springboot-files/00-demo/src/main/java/com/example/demo/DemoApplication.java)

И инициализируем здесь git-репозиторий (чуть ниже будет рассказано зачем)

```bash
git init
git add .
git commit -m 'initial commit'
```

Для того чтобы werf смогла начать работу с нашим приложением - необходимо в корне нашего репозитория создать файл werf.yaml в которым будут описаны инструкции по сборке. Для начала соберем образ локально не загружая его в registry чтобы разобраться с синтаксисом сборки.

С помощью werf можно собирать образы с используя Dockerfile или используя синтаксис, описанный в документации werf (мы называем этот синтаксис и движок, который этот синтаксис обрабатывает, stapel). Для лучшего погружения - соберем наш образ с помощью stapel.

Итак, начнём с самой главной секции нашего werf.yaml файла, которая должна присутствовать в нём **всегда**. Называется она [meta config section](https://werf.io/documentation/configuration/introduction.html#meta-config-section) и содержит всего два параметра.

```yaml
project: spring
configVersion: 1
```
[werf.yaml](gitlab-java-springboot-files/00-demo/werf.yaml:1-2)

**_project_** - поле, задающее имя для проекта, которым мы определяем связь всех docker images собираемых в данном проекте. Данное имя по умолчанию используется в имени helm релиза и имени namespace в которое будет выкатываться наше приложение. Данное имя не рекомендуется изменять (или подходить к таким изменениям с должным уровнем ответственности) так как после изменений уже имеющиеся ресурсы, которые выкачаны в кластер, не будут переименованы.

**_configVersion_** - в данном случае определяет версию синтаксиса используемую в `werf.yaml`.

После мы сразу переходим к следующей секции конфигурации, которая и будет для нас основной секцией для сборки - [image config section](https://werf.io/documentation/configuration/introduction.html#image-config-section). И чтобы werf понял что мы к ней перешли разделяем секции с помощью тройной черты.


```yaml
project: spring
configVersion: 1
---
image: hello
from: maven:3-jdk-8
```

[werf.yaml](gitlab-java-springboot-files/00-demo/werf.yaml:1-5)


**_image_** задает короткое имя собираемого docker-образа. Должно быть уникально в рамках одного werf-файла.

**_from _** - аналогичная секция с обычным dockerfile. В примере spring используется `openjdk:8-jdk-alpine, `но он хорош для запуска, мы же воспользуемся образом в котором уже предустановлены все что необходимо maven для сборки - `maven:3-jdk-8.`

Теперь встает вопрос о том как нам добавить исходный код приложения внутрь нашего docker image. И для этого мы можем использовать Git! И нам даже не придётся устанавливать его внутрь docker image.

**_git_**, на наш взгляд это самый правильный способ добавления ваших исходников внутрь docker image, хотя существуют и другие. Его преимущество в том что он именно клонирует, и в дальнейшем накатывает коммитами изменения в тот исходный код что мы добавили внутрь нашего docker image, а не просто копирует файлы. Вскоре мы узнаем зачем это нужно.

```yaml
project: spring
configVersion: 1
---
image: hello
from: maven:3-jdk-8
git:
- add: /
  to: /app
```

[werf.yaml](gitlab-java-springboot-files/00-demo/werf.yaml:1-8)

Werf подразумевает что ваша сборка будет происходить внутри директории склонированного git репозитория. Потому мы списком можем указывать директории и файлы относительно корня репозитория которые нам нужно добавить внутрь image.

`add: /` - та директория которую мы хотим добавить внутрь docker image, мы указываем, что это весь наш репозиторий

`to: /app` - то куда мы клонируем наш репозиторий внутри docker image. Важно заметить что директорию назначения werf создаст сам.

 Есть возможность даже добавлять внешние репозитории внутрь проекта не прибегая к предварительному клонированию, как это сделать можно узнать [тут](https://werf.io/documentation/configuration/stapel_image/git_directive.html), но мы не рекомендуем такой подход.

Приступим к описанию того как происходит сама сборка.
Сейчас доступно 2 вида сборщика - shell и ansible. Первый аналогичен директиве RUN в dockerfile. Его удобнее использовать для быстрого получения результата с минимальными затратами времени на изучение. ansible более молодой инструмент и требующий несколько большего времени на изучение, но он позволяет получить более прогнозируемый результат вследствии декларативности. 
Пользовательские стадии - их всего 4 before install, install, before setup, setup и детально мы к ним вернемся в разделе управления зависимостями. Подробнее о них можно почитать в [документации](https://werf.io/documentation/configuration/stapel_image/assembly_instructions.html#usage-of-user-stages)

Однако, чтобы запускать jar его нужно предварительно собрать. Предлагается сделать это локально, мы же соберем jar так же используя werf и ansible-сборшик. Поскольку все системные зависимости для сборки удовлетворены - мы используем образ с maven и всеми зависимостями- опишем сборку приложения в стадии setup:

```yaml
project: spring
configVersion: 1
---
image: hello
from: maven:3-jdk-8
git:
- add: /
  to: /app
ansible:
  setup:
  - name: Build jar
    shell: |
      mvn -B -f pom.xml package dependency:resolve
    args:
      chdir: /app
      executable: /bin/bash
```

Уже сейчас можем запустить сборку и получить docker-образ с лежащим внутри jar.

Полный список поддерживаемых модулей ansible в werf можно найти [тут](https://werf.io/documentation/configuration/stapel_image/assembly_instructions.html#supported-modules).

Не забыв [установить werf](https://werf.io/documentation/guides/installation.html) локально, запускаем сборку с помощью [werf build](https://werf.io/documentation/cli/main/build.html)!

```bash
$  werf build --stages-storage :local
```

![werf build](gitlab-java-springboot-files/images/00-demo-build-1.gif "werf build")

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

![werf image](gitlab-java-springboot-files/images/00-demo-build-1-result.png "werf_result")

Теперь его можно запустить локально используя image_id просто с помощью docker.
Либо вместо этого использовать [werf run](https://werf.io/documentation/cli/main/run.html):


```bash
werf run --stages-storage :local --docker-options="-d -p 8080:8080 --restart=always" -- java -jar /app/target/demo-1.0.jar
```

Первая часть команды очень похожа на build, а во второй мы задаем [параметры](https://docs.docker.com/engine/reference/run/) docker и через двойную черту команду с которой хотим запустить наш image.

Небольшое пояснение про `--stages-storage :local `который мы использовали и при сборке и при запуске приложения. Данный параметр указывает на то где werf хранить стадии сборки. На момент написания статьи это возможно только локально, но в ближайшее время появится возможность сохранять их в registry.

Теперь наше приложение доступно локально на порту 8080:

![app running](gitlab-java-springboot-files/images/00-demo-browser.png "app running")

На этом часть с локальным использованием werf мы завершаем и переходим к той части для которой werf создавался, использовании его в CI.

<a name="hello-world-ci" />

## Построение CI-процесса

После того как мы закончили со сборкой, которую можно производить локально, мы приступаем к базовой настройке CI/CD на базе Gitlab.

Начнем с того что добавим нашу сборку в CI с помощью .gitlab-ci.yml, который находится внутри корня проекта. Нюансы настройки CI в Gitlab можно найти [тут](https://docs.gitlab.com/ee/ci/).

Мы предлагаем простой флоу, который мы называем [fast and furious](https://docs.google.com/document/d/1a8VgQXQ6v7Ht6EJYwV2l4ozyMhy9TaytaQuA9Pt2AbI/edit#). Такой флоу позволит вам осуществлять быструю доставку ваших изменений в production согласно методологии GitOps и будут содержать два окружения, production и stage.

На стадии сборки мы будем собирать образ с помощью werf и загружать образ в registry, а затем на стадии деплоя собрать инструкции для kubernetes, чтобы он скачивал нужные образы и запускал их.

<a name="hello-world-ci-build" />

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

![gitlab build success](gitlab-java-springboot-files/images/00-demo-gitlab-build-succes.png "gitlab build success")


Лог в Gitlab будет выглядеть так же как и при локальной сборке, за исключением того что в конце мы увидим как werf пушит наш docker image в registry.

```
│ ┌ Publishing image hello by stages-signature tag 3a35e5ff158514066abadb0012e2fe0f19551902fa0355064aeb4cf7
│ ├ Info
│ │   images-repo: registry.example.com/demo/hello
│ │         image: registry.example.com/demo/hello:3a35e5ff158514066abadb0012e2fe0f19551902fa0355064aeb4cf7
│ └ Publishing image hello by stages-signature tag 3a35e5ff158514066abadb0012e2fe0f19551902fa0355064aeb4cf7 (27.23 seconds)
└ ⛵ image hello (27.32 seconds)

```

<a name="hello-world-ci-deploy" />

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

Для того чтобы запустить наше собранное приложение нужно описать с какой командой будет запускаться контейнер, который мы пушнули в registry. В нашем случае это `java -jar jarfile.jar`. Возможно еще какие то опции java при запуске (debug, лимиты по памяти, к примеру). Так же нужно указать helm-у в каком контейнере все это запускать и где взять образ.
Так же проблемой является то, что при "пустых" коммитах или нерелевантных коммитах (пустой коммит, к примеру, или мы поправили только README), сборка образа не произойдет (этот момент мы чуть позже разберем, когда будем разбираться как оптимизировать сборку), но при деплое все равно будет произведен перезапуск контейнера, так как поменялся его tag.
В нашем же случае, werf умеет немного магии.

Команда запуска в deployment.yaml будет выглядеть так:

```
       command:
       - java
       - -jar
       - /app/target/demo-1.0.jar $JAVA_OPT
```

[deployment.yaml](gitlab-java-springboot-files/00-demo/.helm/templates/10-deployment.yaml:19-22)

Решение же остальных вышеозвученных проблем с image взвалим на werf и её механизм content-based-tagging. О том что это и зачем написана отдельная [статья](https://habr.com/ru/company/flant/blog/495112/). 
Очень вкратце - werf на основании содержимого образа и истории формирования коммита в git делает вывод - поменялся ли image и стоит ли выполнять перезапуск контейнера с ним. А так же проставляет политику скачивания образа в зависимости собирался образ из тега или из бранча. В случае с branch будет проставлен Always, так как имя образа может не поменяться, а вот его содержимое - да.
Для использования этого функционала воспользуемся следующей конструкцией, которую пропишем в deployment:

```yaml
{{ tuple "hello" . | include "werf_container_image" | indent 8 }}
```

[deployment.yaml](gitlab-java-springboot-files/00-demo/.helm/templates/10-deployment.yaml:18)

Стоит так же почитать пулную [документацию](https://ru.werf.io/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_image) по werf_container_image.
Не считая еще одной похоже констуркции, о которой пойдет речь в следующей главе, это обычный deployment, описание параметров которого можно найти в официальной документации по kubernetes.

##### Переменные окружения

Самой частой проблемой связанной с запуском приложения в кубернетес является не использование переменных окружения. Зачастую большая часть переменных, которые требуются приложению прописаны в стандатные для фреймворка места (в нашем случае - application.properties) как есть. Из-за этого появляется соблазн делать образ для каждого окружения отдельно. Мы же рекомендуем подход, что должен собираться один образ и для стейдж и для прод окружения. И для того чтобы это работало необходимо использовать в коде переменные окружения. А сейчас нам нужно эти переменные пробрасывать в контейнер. Для Java это могут быть такие переменные как, например, данные о подключении к БД - хосты, пользователи БД, пароли для различных внутренних и внешних сервисов.

Эти переменные мы параметризируем с помощью файла `values.yaml`.

Например вот так мы опишем что в production окружении мы будем использовать домен example.com, а во всех остальных (_default) - stage.example.com. Так же опишем, что в production нужно запускать приложение без опции --debug.

```yaml
---
app:
  java_opt:
    _default: "--debug"
    production: ""
  url:
    _default: stage.example.com
    prodiction: example.com
```

[values.yaml](gitlab-java-springboot-files/00-demo/.helm/values.yaml)

Переменные окружения так же используются для того, чтобы не перевыкатывать контейнеры, которые не менялись.

Werf закрывает ряд вопросов, связанных с перевыкатом контейнеров с помощью конструкции  [werf_container_env](https://ru.werf.io/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_env). Она возвращает блок с переменной окружения DOCKER_IMAGE_ID контейнера пода. Значение переменной будет установлено только если .Values.global.werf.is_branch=true, т.к. в этом случае Docker-образ для соответствующего имени и тега может быть обновлен, а имя и тег останутся неизменными. Значение переменной DOCKER_IMAGE_ID содержит новый ID Docker-образа, что вынуждает Kubernetes обновить объект.

```yaml
{{ tuple "hello" . | include "werf_container_env" | indent 8 }}
```

[deployment.yaml](gitlab-java-springboot-files/00-demo/.helm/templates/10-deployment.yaml:31)

Аналогично можно пробросить секретные переменные (пароли и т.п.) и у Верфи есть специальный механизм для этого. Но к этому вопросу мы вернёмся позже.

##### Логгирование

Важно понимать, что все логи какие возможно приложение должно писать в stdout. Запись логов в файл будет как менее удобна в диагностике возможных проблем так и сама по себе может привести к проблеме с заканчивающимся местом на ноде kubernetes. А это уже приведет к автоматической попытке kubernetes расчистить место - под переедет на другую ноду, в результате получим рестарт пода и потерю всех логов.
Так же из stdout (посредством лога на ноде kubernetes) мы можем забирать централизованно их в аггрегатор логов, например fluent-ом. Это потребует меньше настроек, так как stdout-логи подов лежат в одном месте для всех подов.
В случае с springboot приложение уже настроено на логирование в stdout по умолчанию, однако можно их донастроить, например переопределив уровень логирования в application.properties. Подробнее - в [документации](https://docs.spring.io/spring-boot/docs/2.1.1.RELEASE/reference/html/boot-features-logging.html).

##### Направление трафика на приложение

Для того чтобы запросы извне попали к нам в приложение нужно 
* открыть порт у пода
* привязать к поду сервис 
* и настроить Ingress, который выступает у нас в качестве балансера.

Если вы мало работали с Kubernetes — эта часть может вызвать у вас много проблем. Большинство тех, кто начинает работать с Kubernetes по невнимательности допускают ошибки при конфигурировании labels и затем занимаются долгой и мучительной отладкой.

Следует внимательно отнестись к соответствию полей labels и selector, так как именно по ним происходит связь внутри кластера kubernetes.

###### Проброс портов

Для того чтобы попадать в наше запущенное приложение нужно сказать kubernetes как найти его. Для этого существует объект [service](https://kubernetes.io/docs/concepts/services-networking/service/).

В service нужно в селекторе указать лейблы подов и порты на которые описываемый сервис должен отправлять трафик.

Связь между service и подом в deployment описана в блоке spec.

```yaml
spec:
  selector:
    app: {{ .Chart.Name | quote }}
```

[service.yaml](gitlab-java-springboot-files/00-demo/.helm/templates/20-service.yaml:6-8)

Selector описывает по каким label нужно искать нужный под (labels описан в deployment выше).

```yaml
 template:
   metadata:
     labels:
       app: {{ .Chart.Name | quote }}
```

[deployment.yaml](gitlab-java-springboot-files/00-demo/.helm/templates/10-deployment.yaml:11-14)

Объект service позволяет сопоставлять любой входящий port и targetPort, например входящий 80 и порт в контейнере 8080, в нашем случае они одинаковы и указаны оба для прозрачности. Спецификация же позволяет не указывать targetPort если он совпадает с port. Можно так же указать несколько портов, если это требуется приложению. Для простоты оставим один порт.

###### Роутинг на Ingress

Мы описали сервис и теперь можем направить внешний (по отношению к кластеру kubernetes) трафик в приложение. Для этого опишем [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) - направим все запросы на корень домена в service описанный выше на порт 8080.

Для этого нужно будет указать набор правил роутинга: с какого домена, с какого пути, в какой сервис и на какой порт надо направлять трафик.

Домен будем задавать так
                       
```yaml
  - host: {{ pluck .Values.global.env .Values.app.url | first | default .Values.app.url._default }}
```

[ingress.yaml](gitlab-java-springboot-files/00-demo/.helm/templates/90-ingress.yaml:10)

Здесь описан выбор в values.yaml url для нашего приложения в зависимости от окружения. Либо будет взят url для явно указанного в values.yaml окружения (например, production), либо будет взять _default значение. Как это выглядит можно посмотреть выше в главе "Переменные окружения".

Для каждого домена могут существовать набор правил для разных путей (location в терминах nginx). В нашем примере

```yaml
      - path: /
        backend:
          serviceName: {{ .Chart.Name | quote }}
          servicePort: 8080
```

[ingress.yaml](gitlab-java-springboot-files/00-demo/.helm/templates/90-ingress.yaml:13-16)

мы направляем все что пришло на корень домена в сервис "spring" (напомню, что это в случае с werf - поле project в werf.yaml) на порт сервиса 8080. Дальше уже с трафиком разбирается service.

#### Секретные переменные

Мы уже рассказывали о том как использовать обычные переменные в нашем СI забирая их напрямую из values.yaml. Суть работы с секретными переменными абсолютно та же, единственное что в репозитории они будут храниться в зашифрованном виде.

Потому для хранения в репозитории паролей, файлов сертификатов и т.п., рекомендуется использовать подсистему работы с секретами werf.

Идея заключается в том, что конфиденциальные данные должны храниться в репозитории вместе с приложением в зашифрованном виде, и должны оставаться независимыми от какого-либо конкретного сервера.

Для того, чтобы начать пользоваться секретными переменными нужно сгенерировать секретный ключ с которым werf сможет его шифровать и расшифровывать во время деплоя.
Делается это из консоли:

```bash
$ werf helm secret generate-secret-key
4710f841e17fabcc85f976e4a665ff9e
```

Чтобы им воспользоваться нужно добавить его либо в свои переменные окружения, либо записать в файл .werf_secret_key. Во втором случае нужно обязательно добавить его в gitignore.

```bash
export WERF_SECRET_KEY=4710f841e17fabcc85f976e4a665ff9e
echo $WERF_SECRET_KEY > .werf_secret_key
```

Теперь добавим секретную перменную, например, password:

```bash
werf helm secret values edit .helm/secret-values.yaml
```

Откроется редактор по умолчанию (согласно перменной EDITOR), куда мы впишем наши парои plain-text в формате аналогичном values.yaml. Например:

```yaml
app:
  password:
    _default: my-secret-password
    production: my-super-secret-password
```

При просмотре результирующего файла увидим лишь такое:

```yaml
app:
  password:
    _default: 10006755d101c5243fc400ababd7358689a921c19ee7e96a95f0ab82d46e4424573ab50ba666fcf5ce5e5dbd2b696c7706cf
    production: 1000bcd51061ebd1b2c2990041d30783be607b3a0aec8890c098f17bc96dc43e93765219651d743c7a37fb7361c10b703c7b
```

[secret-values.yaml](gitlab-java-springboot-files/00-demo/.helm/secret-values.yaml)

И в таком виде уже безопасно хранить пароли в git.
Для того чтобы отредактировать значение нужно снова воспользоваться командой

```bash
werf helm secret values edit .helm/secret-values.yaml
```

Для дальнейшего его использования - для деплоя приложения - в рамках gitlab нужно ключ WERF_SECRET_KEY положить в gitlab variables для проекта (Settings -> CI/CD -> variables). Оттуда werf при запуске получит эту перменную и сформирует корректный helm-chart с расшифрованным паролем.

#### Деплой в Gitlab CI

Опишем деплой приложения в Kubernetes. Деплой будет осуществляться на два стенда: staging и production.

Выкат на два стенда отличается только параметрами, поэтому воспользуемся шаблонами. Опишем базовый деплой, который потом будем кастомизировать под стенды: 

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

![deploy-1](gitlab-java-springboot-files/images/00-demo-gitlab-deploy-1.png "deploy")

Посмотреть статус выполнения pipeline можно в интерфейсе gitlab **CI / CD - Pipelines**

![deploy-2](gitlab-java-springboot-files/images/00-demo-gitlab-deploy-2.png "deploy")

Список всех окружений - доступен в меню **Operations - Environments**

![deploy-3](gitlab-java-springboot-files/images/00-demo-gitlab-deploy-3.png "deploy")

Из этого меню - можно так же быстро открыть приложение в браузере.

![deploy-4](gitlab-java-springboot-files/images/00-demo-gitlab-deploy-4.png "deploy")

Здесь будет подставлен домен, который указывается в секции environment рядом с `name:`

```yaml
   environment:
     name: stage
     url: http://stage.example.com
```

К тому как использовать эту переменную в helm-чарте мы вернемся в самом конце, а пока что для простоты будем использовать .helm/values для объявления доменов.

<a name="dependencies" />

# Подключаем зависимости

Werf подразумевает, что лучшей практикой будет разделить сборочный процесс на этапы, каждый с четкими функциями и своим назначением. Каждый такой этап соответствует промежуточному образу, подобно слоям в Docker. В werf такой этап называется стадией, и конечный образ в итоге состоит из набора собранных стадий. Все стадии хранятся в хранилище стадий, которое можно рассматривать как кэш сборки приложения, хотя по сути это скорее часть контекста сборки.

Стадии — это этапы сборочного процесса, кирпичи, из которых в итоге собирается конечный образ. Стадия собирается из группы сборочных инструкций, указанных в конфигурации. Причем группировка этих инструкций не случайна, имеет определенную логику и учитывает условия и правила сборки. С каждой стадией связан конкретный Docker-образ. Подробнее о том, какие стадии для чего предполагаются можно посмотреть в [документации](https://ru.werf.io/documentation/reference/stages_and_images.html).

Werf предлагает использовать для стадий следующую стратегию:

*   использовать стадию beforeInstall для инсталляции системных пакетов;
*   использовать стадию install для инсталляции системных зависимостей и зависимостей приложения;
*   использовать стадию beforeSetup для настройки системных параметров и установки приложения;
*   использовать стадию setup для настройки приложения.

Подробно про стадии описано в [документации](https://ru.werf.io/documentation/configuration/stapel_image/assembly_instructions.html).

В Java, в частности в spring, в качестве менеджера зависимостей может использоваться maven, gradle. Мы будем, как и ранее использовать maven, но для gradle кроме самих команд, логика сборки не поменяется. Мы уже описывали ранее его использование в файле `werf.yaml` но сейчас оптимизируем его использование.

<a name="dependencies-implementation" />

## Подключение менеджера зависимостей

В maven используется pom.xml в качестве файла проекта где, помимо мета-информации, описываются зависимости. Вот так мы к нему обращались когда собирали приложение ранее:

```yaml
    shell: |
      mvn -B -f pom.xml package dependency:resolve
```

[werf.yaml](gitlab-java-springboot-files/00-demo/werf.yaml:20)

Однако, если оставить всё так — стадия `beforeInstall` не будет запускаться при изменении pom.xml и любого кода в src. Подобная зависимость пользовательской стадии от изменений [указывается с помощью параметра git.stageDependencies](https://ru.werf.io/documentation/configuration/stapel_image/assembly_instructions.html#%D0%B7%D0%B0%D0%B2%D0%B8%D1%81%D0%B8%D0%BC%D0%BE%D1%81%D1%82%D1%8C-%D0%BE%D1%82-%D0%B8%D0%B7%D0%BC%D0%B5%D0%BD%D0%B5%D0%BD%D0%B8%D0%B9-%D0%B2-git-%D1%80%D0%B5%D0%BF%D0%BE%D0%B7%D0%B8%D1%82%D0%BE%D1%80%D0%B8%D0%B8):

```yaml
git:
- add: /
  to: /app
  stageDependencies:
    setup:
    - pom.xml
    - src
```

Теперь при изменении файла pom.xml или любого из файлов в `src` стадия `setup` будет запущена заново.
Почитать о том как формируется pom.xml вручную можно [здесь](https://maven.apache.org/guides/getting-started/maven-in-five-minutes.html). В нашем случае, для spring мы снова воспользуемся веб-интерфейсом [https://start.spring.io/](https://start.spring.io/), который использовали для формирования hello-world приложения. Плюсом этого метода безусловано является его быстрота, возможность избежать глупых ошибок при написании pom.xml.

<a name="dependencies-optimization" />

## Оптимизация сборки

В случае с spring даже в пустом проекте сборщику нужно скачать приличное количество файлов, которое происходит не за нулевое время. И скачивать эти файлы раз за разом выглядит нецелесообразным, тем более что maven (и gradle) позволяет этот кеш переиспользовать при локальной сборке.
Для оптимизации скорости сборки у werf есть несколько механизмов.
Рассмотрим сначала возможность переиспользовать кеш, который скачивает maven в .m2/repository. В werf есть механизм переиспользования так называемого build_dir между сборками в одном проекте. С помощью него и будем пробрасывать кеш в сборочные контейнеры.

Для этого в werf служит директива mount:


```yaml
...
mount:
- from: build_dir
  to: /root/.m2/repository
...
```
[werf.yaml](gitlab-java-springboot-files/01-demo-optimization/werf.yaml:14-16)

Как упоминалось выше - werf кеширует успешные стадии деплоя, что существенно ускоряет сборки. А maven позволяет отдельно выполнить команды resolve зависимостей и выполнять сборку приложения. Вынесем resolve зависимостей на отдельную - предыдущую - стадию. В этом случае, при изменении только кода, а не pom.xml werf увидит слой с зависимостями не поменялся и возьмет его из кеша, а затем уже проведет сборку приложения используя измененный код. Вынесем скачивание локального репозитория в стадию beforeSetup. И настроим стадию setup на сборку в случае изменений чего-либо в папке src, где и лежит сам код:

```yaml
...
ansible:
  beforeSetup:
  - name: dependency resolve
    shell: |
      mvn -B -f pom.xml dependency:resolve
    args:
      chdir: /app
      executable: /bin/bash
  setup:
  - name: Build jar
    shell: |
      mvn -B -f pom.xml package
    args:
      chdir: /app
      executable: /bin/bash
```

[werf.yaml](gitlab-java-springboot-files/01-demo-optimization/werf.yaml:17-31)

Теперь первая сборка - без кешей - будет сравнительно долгой. Будут скачиваться репозитории maven, описанные в pom.xml, затем из кода в src будет собираться приложение. Напомню о механизмах отладки на этом этапе -

```shell
--introspect-before-error
--introspect-error
```

И в случае проблем мы сможем попасть внутрь контейнера и выполнить команды сборки вручную.

На этом этапе у нас получилось оптимизировать скорость сборки - кеш каждый раз не скачивается, стадия не перезапускается при изменении только кода. Однако базовый образ maven:3-jdk-8, который мы использовали для сборки, достаточно тяжелый, нет смысла запускать код со всеми сборочными зависимостями в kubernetes. 
Будем запускать используя достаточно легкий openjdk:8-jdk-alpine. Но нам все еще нужно собирать jar в образе с maven. Для реализации этого решения воспользуемся [артефактом](https://werf.io/documentation/configuration/stapel_artifact.html). По сути это то же самое что и image в директивах werf.yaml, только временный. Он не пушится в registry.
Переименуем image в artifact и назовем его build. Результатом его работы является собранный jar - именно его мы и импортируем в image с alpine-openjdk который опишем в этом же werf.yaml после "---", которые разделяют image. Назовем его spring и уже его пушнем в registry для последующего запуска в кластере.

```yaml
---
image: hello
from: openjdk:8-jdk-alpine
import:
- artifact: build
  add: /app/target/*.jar
  to: /app/demo.jar
  after: setup
```

[werf.yaml](gitlab-java-springboot-files/01-demo-optimization/werf.yaml:32-39)

Для импорта между image и artifact служит директива import. Из /app/target в сборочном артефакте импортируем собранный jar-файл в папку /app в image spring. Единственное что следует еще поправить - это версию собираемого jar в [pom.xml](01-demo-optimization/pom.xml:14). Пропишем её 1.0, чтобы имя итогового jar-файла получось предсказуемым - demo-1.0.jar. 

<a name="assets" />

# Генерируем и раздаем ассеты

В какой-то момент в процессе разработки вам понадобятся ассеты (т.е. картинки, css, js).

Для генерации ассетов мы будем использовать webpack.
Генерировать ассеты для java-spring-maven можно, конечно, разными способами. Например, в maven есть [плагин](https://github.com/eirslett/frontend-maven-plugin), который позволяет описать сборку ассетов "не выходя" из Java. Но там есть несколько оговорок про use-case этого плагина:

*   Не предполагается использовать как замена Node для разработчиков фронтенда. Скорее для того чтобы разработчики бекенда могли быстрее включить JS-код в свою сборку.
*   Не предполагается использование на production-окружениях.

Потому хорошим и распространенным выбором будет использовать webpack отдельно.

Интуитивно понятно, что на стадии сборки нам надо будет вызвать скрипт, который генерирует файлы, т.е. что-то надо будет дописать в `werf.yaml`. Однако, не только там — ведь какое-то приложение в production должно непосредственно отдавать статические файлы. Мы не будем отдавать файлики с помощью {{Frameworkname}}. Хочется, чтобы статику раздавал nginx. А значит надо будет внести какие-то изменения и в helm чарты.

<a name="assets-scenario" />

## Сценарий сборки ассетов

webpack - гибкий в плане реализации ассетов инструмент. Настраивается его поведение в webpack.config.js и package.json.
Создадим в этом же проекте папку [assets](gitlab-java-springboot-files/02-demo-with-assets/assets/). В ней следующая структура

```
├── default.conf
├── dist
│   └── index.html
├── package.json
├── src
│   └── index.js
└── webpack.config.js
```

При настройке ассетов есть один нюанс - при подготовке ассетов мы не рекомендуем использовать какие-либо изменяемые переменные _на этапе сборки_. Потому что собранный бинарный образ должен быть независимым от конкретного окружения. А значит во время сборки у нас не может быть, например, указано домена для которого производится сборка, user-generated контента и подобных вещей.

В связи с описанным выше, все ассеты должны быть сгенерированы одинаково для всех окружений. А вот использовать их стоит в зависимости от окружения в котором запущено приложение. Для этого можно подсунуть приложению уже на этапе деплоя конфиг, который будет зависеть от окружения. Реализуется это через [configmap](https://kubernetes.io/docs/concepts/configuration/configmap/). Кстати, он же будет нам полезен, чтобы положить конфиг nginx внутрь alpine-контейнера. Обо всем этом далее.

<a name="assets-implementation" />

## Какие изменения необходимо внести

Генерация ассетов происходит в отдельном артефакте в 2-х стадиях - `install` и `setup`. На первой стадии мы выполняем `npm install`, на второй уже `npm build`. Не забываем про кеширование стадий и зависимость сборки от изменения определенных файлов в репозитории.
Так же нам потребуются изменения в шаблонах helm. Нам нужно будет описать процесс запуска контейнера с ассетами. Мы не будем их класть в контейнер с приложением. Запустим отдельный контейнер с nginx и в нем уже будем раздавать статику. Соответственно нам потребуются изменения в deployment. Так же потребуется добавить configmap и описать в нем файл с переменными для JS и конфиг nginx. Так же, чтобы трафик попадал по нужному адресу - Потребуются правки в ingress и service.

### Изменения в сборке

В конфиги сборке (werf.yaml) потребуется описать артефакт для сборки и импорт результатов сборки в контейнер с nginx. Делается это аналогично сборке на Java, разумеется используются команды сборки специфичные для webpack/nodejs. Из интересеного - мы добавляем не весь репозиторий для сборки, а только содержимое assets.

```yaml
git:
  - add: /assets
    to: /app
    excludePaths:

    stageDependencies:
      install:
      - package.json
      - webpack.config.js
      setup:
      - "**/*"
```

[werf.yaml](gitlab-java-springboot-files/02-demo-with-assets/werf.yaml:42-53)

Для артефакта сборки ассетов настроили запуск `npm install` в случае изменения `package.json` и `webpack.config.js`. А так же запуск `npm run build` при любом изменении файла в репозитории.

Также стоит исключить assets из сборки Java:

```yaml
git:
- add: /
  to: /app
  excludePaths:
  - assets
```

[werf.yaml](gitlab-java-springboot-files/02-demo-with-assets/werf.yaml:7-10)

Получившийся общий werf.yaml можно посмотреть [в репозитории]([werf.yaml](02-demo-with-assets/werf.yaml)).

### Изменения в деплое

Статику логично раздавать nginx-ом. Значит нам нужно запустить nginx в том же поде с приложением но в отдельном контейнере.

В deployment допишем что контейнер будет называться frontend, будет слушать на 80 порту

```yaml
      containers:
      - name: frontend
{{ tuple "frontend" . | include "werf_container_image" | indent 8 }}
        ports:
        - name: http-front
          containerPort: 80
          protocol: TCP
```

[deployment.yaml](gitlab-java-springboot-files/02-demo-with-assets/.helm/templates/10-deployment.yaml:23-29)

Разумеется, нам нужно положить nginx-конфиг для раздачи ассетов внутрь контейнера. Поскольку в нем могут (в итоге, например для доступа к s3) использоваться переменные - добавляем именно configmap, а не подкладываем файл на этапе сборки образа.
Здесь же добавим js-файл, содердащий переменные, к которым обращается js во время выполнения. Подкладываем его на этапе деплоя именно для того, чтобы иметь гибкость в работе с ассетами в условиях одинаковых исходных образов для разных окружений.

```yaml
        volumeMounts:
        - name: nginx-config
          mountPath: /etc/nginx/conf.d/default.conf
          subPath: default.conf
        - name: env-js
          mountPath: /app/dist/env.js
          subPath: env.js
```

[deployment.yaml](gitlab-java-springboot-files/02-demo-with-assets/.helm/templates/10-deployment.yaml:30-36)

Не забываем указать что файлы мы эти берем из определенного конфигмапа:

```yaml
      volumes:
      - name: nginx-config
        configMap:
          name: {{ .Chart.Name }}-configmap
      - name: env-js
        configMap:
          name: {{ .Chart.Name }}-configmap
```

[deployment.yaml](gitlab-java-springboot-files/02-demo-with-assets/.helm/templates/10-deployment.yaml:52-58)


Здесь мы добавили подключение configmap к deployment. В самом configmap пропишем конфиг nginx - default.conf и файл с переменными для js - env.js. Вот так, например выглядит конфиг с переменной для JS:

```yaml
...
  env.js: |
    module.exports = {
        url: {{ pluck .Values.global.env .Values.app.url | first | default .Values.app.url._default |quote }},
        environment: {{ .Values.global.env |quote }}
    };

```

[Полный файл](gitlab-java-springboot-files/02-demo-with-assets/.helm/templates/01-cm.yaml)

Еще раз обращу внимание на env.js - мы присваиваем url значение в зависимости от окружения указанное в values.yaml. Как раз та "изменчивость" js что нам нужна.

Так же, чтобы kubernetes мог направить трафик в nginx с ассетами нужно дописать port 80 в service, а затем мы направим внешний трафик предназначенный для ассетов в этот сервис.

```yaml
...
  - name: http-front
    port: 80
    targetPort: 80
```

[service.yaml](gitlab-java-springboot-files/02-demo-with-assets/.helm/templates/20-service.yaml:10-12)

### Изменения в роутинге

В ingress направим все что связано с ассетами на порт 80, чтобы все запросы к, в нашем случае example.com/static/ попали в нужный контейнер на 80-ый порт, где у нас будет отвечать nginx, прописанный выше.

```yaml
...
      - path: /static/
        backend:
          serviceName: {{ .Chart.Name | quote }}
          servicePort: 80
```

[ingress.yaml](gitlab-java-springboot-files/02-demo-with-assets/.helm/templates/90-ingress.yaml:17-20)

<a name="files" />

# Работа с файлами

Секция про ассеты подводит нас к другому вопросу - как мы можем сохранять файлы в условиях kubernetes? Нужно учитывать то, что наше приложение в любой момент времени может быть перезапущено kubernetes (и это нормально). Да, kubernetes умеет работать с сетевыми файловыми системами (EFS, NFS, к примеру), которые позволяют подключать общую директорию ко многим подам одновременно. Но правильный путь - использовать s3 для хранения файлов. 
Тем более, что в нашем случае его подключить достаточно просто.
Нужно прописать в pom.xml использование aws-java-sdk.

```xml
<dependency>
   <groupId>com.amazonaws</groupId>
   <artifactId>aws-java-sdk</artifactId>
   <version>1.11.133</version>
</dependency>
```

[pom.xml](gitlab-java-springboot-files/02-demo-with-assets/pom.xml:27-31)

Затем добавить в файл properties приложения - в нашем случае `src/main/resources/application.properties` - сопоставление перменных java с переменными окружения:

```yaml
amazonProperties:
  endpointUrl: ${S3ENDPOINT}
  accessKey: ${S3KEY}
  secretKey: ${S3SECRET}
  bucketName: ${S3BUCKET}
```

[application.properties](gitlab-java-springboot-files/02-demo-with-assets/src/main/resources/application.properties)

Сами доступы нужно записать в values.yaml и secret-values.yaml (используя `werf helm secret values edit .helm/secret-values.yaml` как было описано в главе Секретные переменные). Например:

```yaml
app:
  s3:
    epurl:
      _default: https://s3.us-east-2.amazonaws.com
    bucket:
      _default: mydefaultbucket
      production: myproductionbucket
```

[values.yaml](gitlab-java-springboot-files/02-demo-with-assets/.helm/values.yaml:8-13)

и в secret-values:

```yaml
app:
  s3:
    key:
      _default: mys3keyidstage
      production: mys3keyidprod
    secret:
      _default: mys3keysecretstage
      production: mys3keysecretprod
```

При звкрытии редактора значения [зашифруются](gitlab-java-springboot-files/02-demo-with-assets/.helm/secret-values.yaml).

Чтобы пробросить эти переменные в контейнер нужно в разделе env deployment их отдельно объявить. Наприме:

```yaml
       env:
{{ tuple "hello" . | include "werf_container_env" | indent 8 }}
        - name: S3ENDPOINT
          value: {{ pluck .Values.global.env .Values.app.s3.epurl | first | default .Values.app.s3.epurl._default | quote }}
        - name: S3KEY
          value: {{ pluck .Values.global.env .Values.app.s3.key | first | default .Values.app.s3.key._default | quote }}
        - name: S3SECRET
          value: {{ pluck .Values.global.env .Values.app.s3.secret | first | default .Values.app.s3.secret._default | quote }}
        - name: S3BUCKET
          value: {{ pluck .Values.global.env .Values.app.s3.bucket | first | default .Values.app.s3.bucket._default | quote }}
```

[deployment.yaml](gitlab-java-springboot-files/02-demo-with-assets/.helm/templates/10-deployment.yaml:53-60)

И мы, в зависимости от используемого окружения, можем подставлять нужные нам значения.

<a name="email" />

# Работа с электронной почтой

На наш взгляд самым правильным способом отправки email-сообщений будет внешнее api - провайдер для почты. Например sendgrid, mailgun, amazon ses и подобные.

Рассмотрим на примере sendgrid. spring умеет с ним работать, даже есть [автоконфигуратор](https://docs.spring.io/spring-boot/docs/current/api/org/springframework/boot/autoconfigure/sendgrid/SendGridAutoConfiguration.html). Для этого нужно использовать [библиотеку sendgrid для java](https://github.com/sendgrid/sendgrid-java)
Включим библиотеку в pom.xml, согласно документации:
```xml
...
dependencies {
  ...
  implementation 'com.sendgrid:sendgrid-java:4.5.0'
}

repositories {
  mavenCentral()
}
```

Доступы к sendgrid, как и в случае с s3 пропишем в .helm/values.yaml, пробросим их в виде переменных окружения в наше приложение через deployment, а в коде (в application.properties) сопоставим [java-переменные](https://docs.spring.io/spring-boot/docs/current/reference/html/appendix-application-properties.html) используемые в коде и переменные окружения:

```
# SENDGRID (SendGridAutoConfiguration)
spring.sendgrid.api-key= ${SGAPIKEY}
spring.sendgrid.username= ${SGUSERNAME}
spring.sendgrid.password= ${SGPASSWORD}
spring.sendgrid.proxy.host= ${SGPROXYHOST} #optional
```

Теперь можем использовать эти данные в приложении для отправки почты.

<a name="redis" />

# Подключаем redis

Допустим к нашему приложению нужно подключить простейшую базу данных, например, redis или memcached. Возьмем первый вариант.

В простейшем случае нет необходимости вносить изменения в сборку — всё уже собрано для нас. Надо просто подключить нужный образ, а потом в вашем Java-приложении корректно обратиться к этому приложению.

<a name="redis-to-kubernetes" />

## Завести Redis в Kubernetes

Есть два способа подключить: прописать helm-чарт самостоятельно или подключить внешний. Мы рассмотрим второй вариант.

Подключим redis как внешний subchart.

Для этого нужно:

1. прописать изменения в yaml файлы;
2. указать редису конфиги
3. подсказать werf, что ему нужно подтягивать subchart.

Добавим в файл `.helm/requirements.yaml` следующие изменения:

```yaml
dependencies:
- name: redis
  version: 9.3.2
  repository: https://kubernetes-charts.storage.googleapis.com/
  condition: redis.enabled
```

Для того чтобы werf при деплое загрузил необходимые нам сабчарты - нужно добавить команды в `.gitlab-ci`

```yaml
.base_deploy:
  stage: deploy
  script:
    - werf helm repo init
    - werf helm dependency update
    - werf deploy
```

Опишем параметры для redis в файле `.helm/values.yaml`

```yaml
redis:
  enabled: true
```

При использовании сабчарта по умолчанию создается master-slave кластер redis.

Если посмотреть на рендер (`werf helm render`) нашего приложения с включенным сабчартом для redis, то можем увидеть какие будут созданы сервисы:

```yaml
# Source: example-2/charts/redis/templates/redis-master-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: example-2-stage-redis-master

# Source: example-2/charts/redis/templates/redis-slave-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: example-2-stage-redis-slave
```

<a name="redis-to-app" />

## Подключение приложения к Redis

В нашем приложении - мы будем  подключаться к мастер узлу редиса. Нам нужно, чтобы при выкате в любое окружение приложение подключалось к правильному редису.

Рассмотрим настройки подключения к redis из нашего приложения.

Чтобы начать использовать redis - указаваем в pom.xml нужные denendency для работы с redis

```xml
    <dependency>
      <groupId>org.springframework.boot</groupId>
      <artifactId>spring-boot-starter-data-redis-reactive</artifactId>
    </dependency>
```

[pom.xml](gitlab-java-springboot-files/03-demo-db/pom.xml:32-35)

[werf.yaml](gitlab-java-springboot-files/03-demo-db/werf.yaml) остается неизменным - будем пользоваться тем, что получили в главе оптимизация сборки - чтобы не отвлекаться на сборку ассетов.
Сопоставим переменные java, используемые для подключения к redis и переменные окружения контейнера. 
Как и в случае с работой с файлами выше пропишем в application.properties:

```yaml
spring.redis.host=${REDISHOST}
spring.redis.port=${REDISPORT}
```

[application.properties](gitlab-java-springboot-files/03-demo-db/src/main/resources/application.properties:1-2)

и пеередаем в секции env deployment-а:

```yaml
       env:
{{ tuple "hello" . | include "werf_container_env" | indent 8 }}
        - name: REDISHOST
          value: {{ pluck .Values.global.env .Values.redis.host | first | default .Values.redis.host_default | quote }}
        - name: REDISPORT
          value: {{ pluck .Values.global.env .Values.redis.port | first | default .Values.redis.port_default | quote }}
```

[deployment.yaml](gitlab-java-springboot-files/03-demo-db/.helm/templates/10-deployment.yaml)

Разумеется, требуется и прописать значения для этих переменных в [values.yaml](03-demo-db/.helm/values.yaml).

<a name="database" />

# Подключаем базу данных

Тут будут единые примеры, которые во всех статьях используются. Ибо для сборки и деплоя как-то не сильно разно получается, и хельм одинаковый.

<a name="database-app-generic" />

## Общий подход

TODO: куда-то делась

<a name="database-app-connection" />

## Как подключить БД

Разумеется, нашему приложению может потребоваться не только подключение к redis, но и к другим БД. Но от предыдущего пункта настройка принципиально не отличается. 
Описываем подключение application.properties, прописывая вместо реальных хостов переменные окружения. Правим helm-чарты, добавляя туда нужные переменные окружения и их значения.
Сгенерируем скелет приложения в start.spring.io. В зависимости добавим web и mysql. Скачаем, разархивируем. Все как и раньше. Либо можно просто добавить в pom.xml, который использовался для redis:

```
                <dependency>
                        <groupId>org.springframework.boot</groupId>
                        <artifactId>spring-boot-starter-data-jpa</artifactId>
                </dependency>
...
                <dependency>
                        <groupId>mysql</groupId>
                        <artifactId>mysql-connector-java</artifactId>
                        <scope>runtime</scope>
                </dependency>

```

werf.yaml подойдет от предыдущего шага с redis или тот что использовался в главе оптимизация сборки - они одинаковые. Без ассетов - для более простого восприятия.

Для наглядности результата поправим код приложения, как это описано в https://spring.io/guides/gs/accessing-data-mysql/. Только имя пакета оставим demo. Для единообразия.

Добавим в файл application.properties данные для подключения к mysql. Берем их, как обычно, из переменных окружения, которые передадим в контейнер при деплое.

```
spring.datasource.url=jdbc:mysql://${MYSQL_HOST:localhost}:3306/${MYSQL_DATABASE}
spring.datasource.username=${MYSQL_USER}
spring.datasource.password=${MYSQL_PASSWORD}
```

[application.properties](gitlab-java-springboot-files/03-demo-db/src/main/resources/application.properties:4-6)

Пропишем их в [values.yaml](gitlab-java-springboot-files/03-demo-db/.helm/values.yaml) и [secret-values.yaml](gitlab-java-springboot-files/03-demo-db/.helm/values.yaml), как было описано в прошлых главах. Для secret-values использовались следующие значения:

```yaml
mysql:
  password:
    _default: ThePassword
  rootpassword:
    _default: root
```

Также нам понадобится чарт с mysql. Опишем его сами в [statefulset](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) в helm-чарте. Причины почему стоит использовать именно statefulset для mysql в основном заключаются в требовниях к его стабильности и сохранности данных.
Внутрь контейнера с mysql так же передадим параметры для его работы через переменные окружения часть из которых прописана в values.yaml и шифрованная часть в secret-values.yaml.

```yaml
       env:
       - name: MYSQL_DATABASE
         value: {{ pluck .Values.global.env .Values.infr.mysql.db | first | default .Values.infr.mysql.db._default }}
       - name: MYSQL_PASSWORD
         value: {{ pluck .Values.global.env .Values.infr.mysql.password | first | default .Values.infr.mysql.password._default }}
       - name: MYSQL_ROOT_PASSWORD
         value: root
       - name: MYSQL_USER
         value: {{ pluck .Values.global.env .Values.infr.mysql.user | first | default .Values.infr.mysql.user._default }}
```

[mysql.yaml](gitlab-java-springboot-files/03-demo-db/.helm/templates/20-mysql.yaml:23-30)

Тут возникает еще нюанс - запуск нашего нового приложения полностью зависит от наличия mysql, и, если приложение не сможет достучаться до mysql - оно упадет с ошибкой. Для того чтобы этого избежать добавим init-container для нашего приложения, чтобы оно не запускалось раньше, чем появится коннекта к mysql (statefulset запустится.)

добавится следующая секция в deployment.yaml:

```yaml
    spec:
     initContainers:
     - name: wait-mysql
       image: alpine:3.9
       command:
       - /bin/sh
       - -c
       - while ! getent ahostsv4 $MYSQL_HOST; do echo waiting for mysql; sleep 2; done
```

[deployment.yaml](gitlab-java-springboot-files/03-demo-db/.helm/templates/10-deployment.yaml:16-22)

Все готово - собираем приложение аналогично тому что описано ранее. Делаем запрос на него:
`curl example.com/demo/all`

Получим следующий результат, что говорит о том, что приложение и mysql успешно общаются:

```shell
$ curl example.com/demo/all
[]
$ curl example.com/demo/add -d name=First -d email=someemail@someemailprovider.com
Saved
$ curl example.com/demo/add -d name=Second -d email=some2email@someemailprovider.com     
Saved
$ curl example.com/demo/all |jq
[
 {
   "id": 1,
   "name": "First",
   "email": "someemail@someemailprovider.com"
 },
 {
   "id": 2,
   "name": "Second",
   "email": "some2email@someemailprovider.com"
 }
]
```

<a name="database-migrations" />

## Выполнение миграций

Обычной практикой при использовании kubernetes является выполнение миграций в БД с помощью Job в кубернетес. Это единоразовое выполнение команды в контейнере с определенными параметрами.
Однако в нашем случае spring сам выполняет миграции в соответствии с правилами, которые описаны в нашем коде. Подробно о том как это делается и настраивается прописано в [документации](https://docs.spring.io/spring-boot/docs/2.1.1.RELEASE/reference/html/howto-database-initialization.html). Там рассматривается несколько инструментов для выполнения миграций и работы с бд вообще - Spring Batch Database, Flyway и Liquibase. С точки зрения helm-чартов или сборки совсем ничего не поменятся. Все доступы у нас уже прописаны.

<a name="database-fixtures" />

## Накатка фикстур

Аналогично предыдущему пункту про выполнение миграций, накатка фикстур производится фреймворком самостоятельно используя вышеназванные инструменты для миграций. Для накатывания фикстур в зависимости от стенда (dev-выкатываем, production - никогда не накатываем фикстуры) мы опять же можем передавать переменную окружения в приложение. Например, для Flyway это будет SPRING_FLYWAY_LOCATIONS, для Spring Batch Database нужно присвоить Java-переменной spring.batch.initialize-schema значение переменной из environment. В ней мы в зависимости от окружения проставляем либо always либо never. Реализации разные, но подход один - обязательно нужно принимать эту переменную извне. Документация та же что и в предыдущем пункте.

<a name="unit-testing" />

# Юнит-тесты и Линтеры

Java - компилируемый язык, значит в случае проблем в коде приложение с большой вероятностью просто не соберется. Тем не менее хорошо бы получать информацию о проблеме в коде не дожидаясь пока упадет сборка.
Чтобы этого избежать пробежимся по коду линтером, а затем запустим unit-тесты.
Для запуска линта воспользуемся [maven checkstyle plugin](https://maven.apache.org/plugins/maven-checkstyle-plugin/usage.html). Запускать можно его несколькими способами - либо вынести на отдельную стадию в gitlab-ci - перед сборкой или вызывать только при merge request. Например:

``` yaml
test: 
  stage: test
  script: mvn checkstyle:checkstyle
  only:
  - merge_requests
```

Так же можно добавить этот плагин в pom.xml в секцию build (подробно описано в [документации](https://maven.apache.org/plugins/maven-checkstyle-plugin/usage.html) или можно посмотреть на готовый [pom.xml](воттут)) и тогда checkstyle будет выполняться до самой сборки при выполнении `mvn package`. Воспользуемся как раз этим способом для нашего примера. Стоит отметить, что в нашем случае используется [google_checks.xml](https://github.com/checkstyle/checkstyle/blob/master/src/main/resources/google_checks.xml) для описания checkstyle и мы запускаем их на стадии validate - до компиляции.

Для unit-тестирования воспользуемся инструментом предлагаемым по умолчанию - junit. Если вы пользовались start.spring.io - то он уже включен в pom.xml автоматичекси, если нет, то нужно его там прописать.
Запускаются тесты при выполнении `mvn package`.

Подробно посмотреть как это работает можно у нас в этом [репозитории](gitlab-java-springboot-files/04-demo-tests/)

<a name="multiple-apps" />

# Несколько приложений в одной репе

Как было уже описано в главе про сборку ассетов - можно использовать один репозиторий для сборки нескольких приложений - у нас это были бек на Java и фронт, представляющий собой собранную webpack-ом статику в контейнере с nginx.
Подробно такая сборка описана в отдельной [статье](https://ru.werf.io/documentation/guides/advanced_build/multi_images.html).
Вкратце напомню о чем шла речь.
У нас есть основное приложение - Java и ассеты, собираемые webpack-ом. 
Все что связано с генерацией ассетов лежит в assets. Мы добавили отдельный artifact (на основе nodejs) и image (на основе nginx:alpine). 
В artifact добавляем в /app нашу папку assets, там все собираем, как описано было ранее и отдаем результат в image с nginx.

```yaml
git:
  - add: /assets
    to: /app
    excludePaths:

    stageDependencies:
      install:
      - package.json
      - webpack.config.js
      setup:
      - "**/*"
```

[werf.yaml](gitlab-java-springboot-files/05-demo-complete/werf.yaml:45-53)

Такая организация репозитория удобна, если нужно выкатывать приложение целиком, или команда разработки не большая. Мы уже [рассказывали](https://www.youtube.com/watch?v=g9cgppj0gKQ) в каких случаях это правильный путь.

<a name="feature-envs" />

# Динамические окружения

Во время разработки и эксплуатации приложения может потребоваться использовать не только условные stage и production окружения. Зачастую, удобно что-то разрабатывать в изолированном от других задач стенде.
Поскольку у нас уже готовы описание сборки, helm-чарты, достаточно прописать запуск таких стендов (review-стенды или feature-стенды) из веток. Для этого добавим в .gitlab-ci.yaml кнопки для [start](gitlab-java-springboot-files/05-demo-complete/.gitlab-ci.yaml:62-71):

```yaml
Deploy to Review:
  extends: .base_deploy
  stage: deploy
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    url: review-${CI_COMMIT_REF_SLUG}.example.com
    on_stop: Stop Review
  only:
    - /^feature-*/
  when: manual
```

и [стоп](gitlab-java-springboot-files/05-demo-complete/.gitlab-ci.yaml:73-82) этих стендов:
```yaml
Stop Review:
  stage: deploy
  script:
    - werf dismiss --env $CI_ENVIRONMENT_SLUG --namespace ${CI_ENVIRONMENT_SLUG} --with-namespace
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  only:
    - /^feature-*/
  when: manual
```

здесь мы видим новую для нас команду [werf dismiss](https://werf.io/documentation/cli/main/dismiss.html). Она удалит приложение из kubernetes, helm-релиз так же будет удален вместе с namespace в который выкатывалось приложение.

В .gitlab-ci, как уже упоминалось ранее удобно передавать environment url - его видно на вкладке environments. Можно не вспоминая у какой задачи какая ветка визуально её найти и нажать на ссылку в gitlab.

Нам сейчас нужно использовать этот механизм для формирования наших доменов в ingress. И передать в приложение, если требуется.
Передадим .helm в .base_deploy это переменную аналогично тому что уже сделано с environment:


```
     --set "global.env=${CI_ENVIRONMENT_SLUG}"
     --set "global.ci_url=$(basename ${CI_ENVIRONMENT_URL})"

```

* CI_ENVIRONMENT_SLUG - встроенная перменная gitlab, "очишенная" от некорректных с точки зрения DNS, URL. В нашем случае позволяет избавиться от символов, которые kubernetes не может воспринять - слеши, подчеркивания. Заменяет на "-".
* CI_ENVIRONMENT_URL - мы прокидываем переменную которую указали в environment url.

В .helm сделали себе переменную ci_url. Теперь можно использовать её в [ingress](gitlab-java-springboot-files/05-demo-complete/.helm/templates/90-ingress.yaml:10):

```
  - host: {{ .Values.global.ci_url }}
```

или в любом месте где объявляется она - например, как в случае с [ассетами](gitlab-java-springboot-files/05-demo-complete/.helm/templates/01-cm.yaml:19)

При выкате такого окружения мы получим полноценное окружение - redis, mysql, ассеты. Все это собирается одинаково, что для stage, что для review. 

Разумеется, в production не стоит использовать БД в кубернетес - её нужно будет исключить из шаблона, как показано в примере. Для всего что не production - выкатываем бд в кубернетес. А для всего остального - service и связанный с ним объект endpoint -подробнее -в [документации kubernetes](https://kubernetes.io/docs/concepts/services-networking/service/). И в репозитории - в шаблоне [mysql](gitlab-java-springboot-files/05-demo-complete/.helm/templates/20-mysql.yaml:54-78) и в [values](gitlab-java-springboot-files/05-demo-complete/.helm/values.yaml:18-21), куда добавляются данные для подключения к внешней БД. Но с точки зрения работы приложения - это никак его не касается - он так же ходит на хост mysql. Все остальные параметры продолжают браться из старых перменных.

Таким образом мы можем автоматически заводить временные окружения в кубернетес и автоматически их закрывать при завершении тестирования задачи. При этом использовать один и тот же образ начиная с review вплоть до production, что позволяет нам тестировать именно то, что в итоге будет работать в production.
