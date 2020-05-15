---
author_team: "Alfa"
author_name: "Andrey Koregin"
ci: "gitlab"
language: "nodejs"
framework: "react"
is_compiled: 0
package_managers_possible:
 - npm
 - yarn
 - pnpm
package_managers_chosen: "npm"
unit_tests_possible:
 - flask-sqlalchemy
 - pytest
 - unittest
 - nose
 - nose2
unit_tests_chosen: "flask-sqlalchemy"
assets_generator_possible:
 - webpack
 - gulp
assets_generator_chosen: "webpack"
---

# Чек-лист готовности статьи
<ol>
<li>Все примеры кладём в <a href="https://github.com/flant/examples">https://github.com/flant/examples</a>

<li>Для каждой статьи может и должно быть НЕСКОЛЬКО примеров, условно говоря — по примеру на главу это нормально.

<li>Делаем примеры И на Dockerfile, И на Stapel

<li>Про хельм говорим, про особенности говорим, но в подробности не вдаёмся — считаем, что человек умеет в кубовые ямлы.

<li>Обязательно тестируйте свои примеры перед публикацией
</li>
</ol>

# Введение

Рассмотрим разные способы которые помогут Nodejs программисту собрать приложение и запустить его в kubernetes кластере.

Предполагается что читатель имеет базовые знания в разработке на Nodejs а также немного знаком с Gitlab CI и примитивами kubernetes, либо готов во всём этом разобраться самостоятельно. Мы постараемся предоставить все ссылки на необходимые ресурсы, если потребуется приобрести какие то новые знания.  

Собирать приложения будем с помощью werf. Данный инструмент работает в Linux MacOS и Windows, инструкция по [установке](https://ru.werf.io/documentation/guides/installation.html) находится на официальном [сайте](https://ru.werf.io/). В качестве примера - также приложим Docker файлы.

Для иллюстрации действий в данной статье - создан репозиторий с исходным кодом, в котором находятся несколько простых приложений. Мы постараемся подготовить примеры чтобы они запускались на вашем стенде и постараемся подсказать, как отлаживать возможные проблемы при вашей самостоятельной работе.


## Подготовка приложения

Наилучшим образом приложения будут работать в Kubernetes - если они соответствуют [12 факторам heroku](https://12factor.net/). Благодаря этому - у нас в kubernetes работают stateless приложения, которые не зависят от среды. Это важно, так как кластер может самостоятельно переносить приложения с одного узла на другой, заниматься масштабированием и т.п. — и мы не указываем, где конкретно запускать приложение, а лишь формируем правила, на основании которого кластер принимает свои собственные решения.

Договоримся что наши приложения соответствуют этим требованиям. На хабре уже было описание данного подхода, вы можете почитать про него например [тут](https://12factor.net/).


## Подготовка и настройка среды

Для того, чтобы пройти по этому гайду, необходимо, чтобы

*   У вас был работающий и настроенный Kubernetes кластер
*   Код приложения находился в Gitlab
*   Был настроен Gitlab CI, подняты и подключены к нему раннеры

Для пользователя под которым будет производиться запуск runner-а - нужно установить multiwerf - данная утилита позволяет переключаться между версиями werf и автоматически обновлять его. Инструкция по установке - доступна по [ссылке](https://ru.werf.io/documentation/guides/installation.html#installing-multiwerf).

Для автоматического выбора актуальной версии werf в канале stable, релиз 1.1 выполним следующую  команду:

```
. $(multiwerf use 1.1 stable --as-file)
```

Перед деплоем нашего приложения необходимо убедиться что у нас подготовлены инфраструктурные компоненты:

*   К gitlab подключен shell runner с тегом werf. [Инструкция](https://ru.werf.io/documentation/guides/gitlab_ci_cd_integration.html#%D0%BD%D0%B0%D1%81%D1%82%D1%80%D0%BE%D0%B9%D0%BA%D0%B0-runner) по подготовке gitlab runner.
*   Ранеры включены и активны для репозитория с нашим приложением
*   Для пользователя под которым запускается сборка и деплой установлен kubectl и добавлен конфигурационный файл для подключения к kubernetes.
*   Для gitlab включен и настроен gitlab registry
*   Раннер запущен на отдельной виртуалке, имеет доступ к API kubernetes и запускается по тегу werf  

## Настройка Gitlab Runner

TODO: наверное надо это сюда в шаблон вписать

# Hello world

В первой главе мы покажем поэтапную сборку и деплой приложения без задействования внешних ресурсов таких как база данных и сборку ассетов.

Наше приложение будет состоять из одного docker образа собранного с помощью werf. Его единственной задачей будет вывод сообщения “hello world” по http.

В этом образе будет работать один основной процесс node, который запустит приложение.

Управлять маршрутизацией запросов к приложению будет управлять Ingress в kubernetes кластере.

Мы реализуем два стенда: production и staging. В рамках hello world приложения мы предполагаем, что разработка ведётся локально, на вашем компьютере.

_В ближайшее время werf реализует удобные инструменты для локальной разработки, следите за обновлениями._

В качестве иллюстрации будут использованы шаблоны приложений [https://github.com/chandantudu/nodejs-login-registration](https://github.com/chandantudu/nodejs-login-registration) и [https://itnext.io/build-a-group-chat-app-in-30-lines-using-node-js-15bfe7a2417b](https://itnext.io/build-a-group-chat-app-in-30-lines-using-node-js-15bfe7a2417b).  \
С их помощью мы сможем реализовать процесс сборки и деплоя легковесного чата с собственной авторизацией.


## Локальная сборка

Для того чтобы werf смогла начать работу с нашим приложением - необходимо в корне нашего репозитория создать файл werf.yaml в которым будут описаны инструкции по сборке. Для начала соберем образ локально не загружая его в registry чтобы разобраться с синтаксисом сборки. 

С помощью werf можно собирать образы с используя Dockerfile или используя синтаксис, описанный в документации werf (мы называем этот синтаксис и движок, который этот синтаксис обрабатывает, stapel). Для лучшего погружения - соберем наш образ с помощью stapel.

Прежде всего нам необходимо собрать docker image с нашим приложением внутри. 

Клонируем наши исходники любым удобным способом. В нашем случае это:


```
git clone git@gitlab-example.com:article/chat.git
```

После, в корне склоненного проекта, создаём файл _werf.yaml. _Данный файл будет отвечать за сборку вашего приложения и он обязательно должен находиться в корне проекта. Исходный код находится в отдельной папке _node,_ в данном случае это сделано просто для удобства, чтобы подразделить исходный код проекта от сборочной части.

![images]( /werf-articles/gitlab-nodejs-files/images/structure.png "Image Title")


Итак, начнём с самой главной секции нашего werf.yaml файла, которая должна присутствовать в нём **всегда**. Называется она [meta config section](https://werf.io/documentation/configuration/introduction.html#meta-config-section) и содержит всего два параметра.

werf.yaml:
```
project: chat
configVersion: 1
```

**_project_** - поле, задающее имя для проекта, которым мы определяем связь всех docker images собираемых в данном проекте. Данное имя по умолчанию используется в имени helm релиза и имени namespace в которое будет выкатываться наше приложение. Данное имя не рекомендуется изменять (или подходить к таким изменениям с должным уровнем ответственности) так как после изменений уже имеющиеся ресурсы, которые выкачаны в кластер, не будут переименованы.

**_configVersion_** - в данном случае определяет версию синтаксиса используемую в `werf.yaml`.

После мы сразу переходим к следующей секции конфигурации, которая и будет для нас основной секцией для сборки - [image config section](https://werf.io/documentation/configuration/introduction.html#image-config-section). И чтобы werf понял что мы к ней перешли разделяем секции с помощью тройной черты.


```
project: chat
configVersion: 1
---
image: node
from: node:14-stretch
```

**_image_** - поле задающее имя нашего docker image, с которым он будет запушен в registry.

**_from _** - задает имя базового образа который мы будем использовать при сборке. Задаем мы его точно так же, как бы мы это сделали в dockerfile, т.к. приложение у нас на node js, мы берём готовый docker image - _node_  с тэгом _14-stretch_. (означает что будет использована 14 версия nodejs, а базовый имидж построен на debian системе)

Теперь встает вопрос о том как нам добавить наш исходный код внутрь нашего docker image. И для этого мы можем использовать Git! И нам даже не придётся устанавливать его внутрь docker image.

**_git_**, на наш взгляд это самый правильный способ добавления ваших исходников внутрь docker image, хотя существуют и другие. Его преимущество в том что он именно клонирует, и в дальнейшем накатывает коммитами изменения в тот исходный код что мы добавили внутрь нашего docker image, а не просто копирует файлы. Вскоре мы узнаем зачем это нужно.

```
project: chat
configVersion: 1
---
{{image: jopa}}
{{from: jopa:14-01-03}}
git:
- add: /
  to: /app
```

Werf подразумевает что ваша сборка будет происходить внутри директории склонированного git репозитория. Потому мы списком можем указывать директории и файлы относительно корня репозитория которые нам нужно добавить внутрь image.

`add: /` - та директория которую мы хотим добавить внутрь docker image, мы указываем, что это весь наш репозиторий

`to: /app` - то куда мы клонируем наш репозиторий внутри docker image. Важно заметить что директорию назначения werf создаст сам.

 Есть возможность даже добавлять внешние репозитории внутрь проекта не прибегая к предварительному клонированию, как это сделать можно узнать [тут](https://werf.io/documentation/configuration/stapel_image/git_directive.html).

Но на этом наша сборка, сама собой, не заканчивается и теперь пора приступать к действиям непосредственно внутри имиджа. Для этого мы будем описывать сборку через ansible. Да-да вы не ослышались, сборку приложения можно сконфигурировать так же как обычный, знакомый всем, ansible playbook. \
	Прежде чем описывать задачи в ansible, необходимо добавить еще два интересных поля. \

```
---
image: node
from: node:14-stretch
git:
- add: /ws-backend
  to: /app
ansible:
  beforeInstall:
  install:
```


Это поля **_beforeInstall_** и **_install_**.

 \
В их понимании нам поможет это изображение: \

![]( /werf-articles/gitlab-nodejs-files/images/stages.png "Stages")

Полный список поддерживаемых модулей ansible в werf можно найти [тут](https://werf.io/documentation/configuration/stapel_image/assembly_instructions.html#supported-modules).

Не забыв [установить werf](https://werf.io/documentation/guides/installation.html) локально, запускаем сборку с помощью [werf build](https://werf.io/documentation/cli/main/build.html)!

```
$  werf build --stages-storage :local
```

![alt_text](images/-0.gif "image_tooltip")

    Эта картинка упрощенно иллюстрирует процесс сборки имиджа с помощью werf. Тут мы видим что данные поля названы как стадии сборки на картинке. \
Главное, что сразу можно увидеть, **_beforeInstall_** выполняется раньше чем мы добавляем исходники с помощью **_git_**.

Это сделано специально чтобы мы могли произвести базовую настройку имиджа до того как внутрь попадёт исходный код и начнется сборка. Если вы предпочитаете собирать образ с 0 используя такие образа как alpine, то эта стадия будет для вас особенно полезна, чтобы поставить все необходимые пакеты. \
 \
	Остальные три стадии **_install_**,**_beforeSetup_**,**_setup _**не имеют принципиальных различий, но такое подразделение было сделано не просто так. \
Дело в том что werf умеет отслеживать изменения производимые с файлами в репозитории, благодаря git с помощью которого мы добавляем наш исходный код внутрь docker image. Тем самым мы можем контролировать процесс сборки не пересобирая при изменениях весь имидж с 0. Про стадии можно узнать подробнее [тут](https://werf.io/documentation/configuration/stapel_image/assembly_instructions.html#usage-of-user-stages). На самом деле их может быть неограниченное количество.

Давайте посмотрим как это выглядит:


```
---
image: node
from: node:14-stretch
git:
- add: /node
  to: /app
  stageDependencies:
    install:
    - package.json
ansible:
  beforeInstall:
  - name: Install dependencies
    apt:
      name:
      - tzdata
      - locales
      update_cache: yes
  install:
  - name: npm сi
    shell: npm сi
    args:
      chdir: /app
```



    Мы добавили значительно количество строк, но на самом деле те кто уже имел


    дело с ansbile, уже должны были узнать в них обычные ansible tasks.

В **_beforeInstall_** мы, c помощью [apt](https://docs.ansible.com/ansible/latest/modules/apt_module.html), добавили установку обычных deb пакетов отвечающих за таймзону и локализацию. 

А в **_install _**у нас расположился запуск установки зависимостей с помощью npm, запуская его просто как команду через модуль [shell](https://docs.ansible.com/ansible/latest/modules/shell_module.html).

Полный список поддерживаемых модулей ansible в werf можно найти [тут](https://werf.io/documentation/configuration/stapel_image/assembly_instructions.html#supported-modules).

Но еще мы добавили в git следующую конструкцию:


```
  stageDependencies:
    install:
    - package.json
```


Те кто уже сталкивался со сборками nodejs приложений знают, что в файле package.json указываются зависимости которые вам нужны для сборки вашего приложения. Потому самое логичное указать данный файл в зависимости сборки, чтобы в случае если файл изменится была перезапущена сборка только стадии **_install_**. Мы также можем использовать регулярные выражения для файлов, в отслеживании изменений которых мы нуждаемся. Как задавать регулярные выражения можно узнать [тут](https://werf.io/documentation/configuration/stapel_image/git_directive.html).

И на этом всё! Мы описали необходимый минимум нужный для сборки нашего приложения. Теперь нам достаточно её запустить! \


Мы просто берём последнюю стабильную версию werf исполняя эту команду:


```
type multiwerf && . $(multiwerf use 1.1 stable --as-file)
```


И видим что werf был установлен:


```
user:~/chat$ type multiwerf && . $(multiwerf use 1.1 stable --as-file)
multiwerf is /home/user/bin/multiwerf
multiwerf v1.3.0
Starting multiwerf self-update ...
Self-update: Already the latest version
GC: Actual versions: [v1.0.13 v1.1.10-alpha.6 v1.1.8+fix16 v1.1.9+fix6]
GC: Local versions:  [v1.0.13 v1.1.9+fix6]
GC: Nothing to clean
The version v1.1.8+fix16 is the actual for channel 1.1/stable
Downloading the version v1.1.8+fix16 ...
The actual version has been successfully downloaded
```


Теперь наконец запускаем сборку с помощью [werf build](https://werf.io/documentation/cli/main/build.html) !


```
user:~/chat$  werf build --stages-storage :local
```

![build](/werf-articles/gitlab-nodejs-files/images/build.gif "build")

Вот и всё, наша сборка успешно завершилась. К слову если сборка падает и вы хотите изнутри контейнера её подебажить вручную, то вы можете добавить в команду сборки флаги:

```
--introspect-before-error
```

или

```
--introspect-error
``` 

Которые при падении сборки на одном из шагов автоматически откроют вам shell в контейнер, перед исполнением проблемной инструкции или после.

В конце werf отдал информацию о готовом image:

![alt_text](images/-1.png "image_tooltip")

Теперь его можно запустить локально используя image_id просто с помощью docker.
Либо вместо этого использовать [werf run](https://werf.io/documentation/cli/main/run.html):


```
werf run --stages-storage :local --docker-options="-d -p 8080:8080 --restart=always" -- node /app/src/js/index.js
```

Первая часть команды очень похожа на build, а во второй мы задаем [параметры](https://docs.docker.com/engine/reference/run/) docker и через двойную черту команду с которой хотим запустить наш image.

Небольшое пояснение про `--stages-storage :local `который мы использовали и при сборке и при запуске приложения. Данный параметр указывает на то где werf хранить стадии сборки. На момент написания статьи это возможно только локально, но в ближайшее время появится возможность сохранять их в registry.

Теперь наше приложение доступно локально на порту 8080:

![redyapp](/werf-articles/gitlab-nodejs-files/images/readyapp.gif "readyapp")

На этом часть с локальным использованием werf мы завершаем и переходим к той части для которой werf создавался, использовании его в CI.

## Построение CI-процесса

После того как мы закончили со сборкой, которую можно производить локально, мы приступаем к базовой настройке CI/CD на базе Gitlab.

Начнем с того что добавим нашу сборку в CI с помощью .gitlab-ci.yml, который находится внутри корня проекта. Нюансы настройки CI в Gitlab можно найти [тут](https://docs.gitlab.com/ee/ci/).

Мы предлагаем простой флоу, который мы называем [fast and furious](#). Такой флоу позволит вам осуществлять быструю доставку ваших изменений в production согласно методологии GitOps и будут содержать два окружения, production и stage.

На стадии сборки мы будем собирать образ с помощью werf и загружать образ в registry, а затем на стадии деплоя собрать инструкции для kubernetes, чтобы он скачивал нужные образы и запускал их.

### Сборка в Gitlab CI

Для того, чтобы настроить CI-процесс создадим .gitlab-ci.yaml в корне репозитория с таким содержанием:

```
variables:
  WERF_VERSION: "1.1 beta"

stages:
  - build
  - deploy

Build:
  stage: build
  before_script:
    - type multiwerf && source <(multiwerf use ${WERF_VERSION})
    - type werf && source <(werf ci-env gitlab --tagging-strategy stages-signature --verbose)
  script:
    - werf build-and-publish --stages-storage :local
  tags:
    - werf
```
В переменных мы сохраняем версию werf, для того чтобы в будущем было проще его обновлять.

Werf имеет свои релиз каналы, по которым легко можно получить нужную вам версию. \
Про релиз каналы и их актуальное состояние можно узнать [тут](https://werf.io/releases.html).

 \
Начиная со стадии билда мы уже видим изменения при сборке. Так как в стадию сборки

у нас будет входить также и push имиджей в регистри. ( Для этого на gitlab-runner инстансе обязательно должен быть установлен multiwerf, от пользователя gitlab-runner, как это сделать можно узнать [тут](https://werf.io/documentation/guides/installation.html#installing-multiwerf). )

Преимущества multiwerf в установке сразу нескольких версий werf на инстанс уже указывались в предыдущей главе.


```
    - type multiwerf && source <(multiwerf use ${WERF_VERSION})
```


Первой строкой мы как раз и указываем версию werf которая нам нужна в этом проекте.

 \



```
    - type werf && source <(werf ci-env gitlab --tagging-strategy stages-signature --verbose)
```


Второй строкой мы выполняем настройку нашего werf на работу в Gitlab окружении, этой строкой мы автоматически настраиваем werf на сбор из окружения переменных, которые Gitlab автоматически подставляет в pipeline. Переменные Gitlab можно найти [тут](https://docs.gitlab.com/ee/ci/variables/README.html). \


Флаг` --tagging-strategy stages-signature `выставляет стратегию тэгирования ваших docker images на основе контента, который они содержат. Это называется [content based tagging](https://werf.io/documentation/reference/publish_process.html#content-based-tagging). 

Суть её в том что тэг является хэшсуммой всех стадий сборки. Такой подход избавляет нас от лишних пересборок.

Пример переменных автоматически выставляемых этой командой:

```
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

В рамках статьи нам хватит значений выставляемых по умолчанию. \

```
werf build-and-publish --stages-storage :local
```
Ну и последняя команда непосредственно запускает билд ваших образов. Указание **_--stages-storage :local_** имеет то же значение, что и в локальной сборке.

Дело в том что werf хранит стадии сборки раздельно, как раз для того чтобы мы могли не пересобирать весь образ, а только отдельные его части.  \
	Плюс стадий в том, что они имеют собственный тэг, который представляет собой хэш содержимого нашего образа. Тем самым позволяя полностью избегать не нужных пересборок наших имиджей. Если вы собираете ваше приложение в разных ветках, и исходный код в них различается только конфигами которые используются для генерации статики на последней стадии. То при сборке имиджа одинаковые стадии пересобираться не будут, будут использованы уже собранные стадии из соседней ветки. Тем самым мы резко снижаем время доставки кода. \
 \
Следующие параметры тем кто работал с гитлаб уже должны быть знакомы.

**_tags _**- нужен для того чтобы выбрать наш раннер, на который мы навесили этот тэг. В данном случае наш gitlab-runner в Gitlab имеет тэг werf


```
  except:
    - schedules  
  tags:
    - werf
```


Теперь мы можем запушить наши изменения и увидеть что наша стадия успешно выполнилась.
![gbuild](/werf-articles/gitlab-nodejs-files/images/gbuild.png "gbuild")


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

### Деплой

werf использует встроенный Helm для применения конфигурации в Kubernetes. Для описания объектов Kubernetes werf использует конфигурационные файлы Helm: шаблоны и файлы с параметрами (например, values.yaml). Помимо этого, werf поддерживает дополнительные файлы, такие как файлы c секретами и с секретными значениями (например secret-values.yaml), а также дополнительные Go-шаблоны для интеграции собранных образов.

Werf (по аналогии с helm) берет yaml шаблоны, генерирует из них  огромную простыню с финальными ямлами, куда подставлены все значения. В этой простыне ямла — аннотации для кубернетеса. Эта простыня закидывается в кубернетес кластер, который парсит инструкции в ямле и вносит изменения в кластер. Верфь смотрит за тем, как кубернетес вносит изменения и дожидается, чтобы реально всё было применено.

Внутри Werf доступны команды Helm-а, например, проверить какие файлы получаются в результате работы werf с шаблонами можно выполнив команду рендер:

```
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

Для работы нашего приложения в среде Kubernetes понадобится описать сущности Deployment, Service и завернуть трафик на приложение, донастроив роутинг в кластере.

##### Запуск контейнера
Начнем с описания deployment.yaml

```
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}
spec:
  revisionHistoryLimit: 3
  strategy:
    type: RollingUpdate
  replicas: 1
  selector:
    matchLabels:
      app: {{ $.Chart.Name }}
  template:
    metadata:
      labels:
        app: {{ $.Chart.Name }}
    spec:
      imagePullSecrets:
      - name: "registrysecret"
      containers:
      - name: {{ $.Chart.Name }}
{{ tuple "node" . | include "werf_container_image" | indent 8 }}
        workingDir: /app
        command: ["node","/app/src/js/index.js"]
        ports:
        - containerPort: 8080
          protocol: TCP
```
Коснусь только шаблонизированных параметров. Значение остальных параметров можно найти в документации [Kubernetes](https://kubernetes.io/docs/concepts/).

{{ . Chart.Name }} - значение для данного параметра берётся из файла werf.yaml из поля **_project \
_**werf.yaml


```
project: chat
configVersion: 1
```



```
{{ tuple "node" . | include "werf_container_image" | indent 8 }}
```


Данный шаблон отвечает за то чтобы вставить информацию касающуюся местонахождения нашего doсker image в registry, чтобы kubernetes знал откуда его скачать. А также политику пула этого образа. И в итоге эта строка будет заменена helm’ом на это: \



```
   image: registry.gitlab-example.com/chat/node:6e3af42b741da90f8bc674e5646a87ad6b81d14c531cc89ef4450585   
   imagePullPolicy: IfNotPresent
```


Замену производит сам werf из собственных переменных. Изменять эту конструкцию нужно только в двух местах: \
1. Рядом в первой части “node”  -  это название вашего docker image, которые мы указывали в werf.yaml в поле **image**, когда описывали сборку.

2. Intent 8 - параметр указывает какое количество пробелов вставить перед блоком, делаем мы это чтобы не нарушить синтаксис yaml, где пробелы(отступы) играют важную разделительную роль.  \
При разработке особенно важно учитывать что yaml не воспринимает табуляцию **только пробелы**!


##### Переменные окружения

Для корректной работы нашего приложения ему нужно узнать переменные окружения.
Для Express nodejs это, например, указание окружение в котором он будет запускаться.

И эти переменные можно параметризовать с помощью файла `values.yaml`.

Так например, мы пробросим значение переменной NODE_ENV в наш контейнер из `values.yaml`

```
app:
  env:
    stage: "staging"
    production:
```

```
    23          - name: NODE_ENV
    24            value: {{ pluck .Values.global.env .Values.app.env | first }}
```
}}

Переменные окружения иногда используются для того, чтобы не перевыкатывать контейнеры, которые не менялись.

Werf закрывает ряд вопросов, связанных с перевыкатом контейнеров с помощью конструкции  [werf_container_env](https://ru.werf.io/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_env). Она возвращает блок с переменной окружения DOCKER_IMAGE_ID контейнера пода. Значение переменной будет установлено только если .Values.global.werf.is_branch=true, т.к. в этом случае Docker-образ для соответствующего имени и тега может быть обновлен, а имя и тег останутся неизменными. Значение переменной DOCKER_IMAGE_ID содержит новый ID Docker-образа, что вынуждает Kubernetes обновить объект.

```yaml
    21          env:
    22  {{ tuple "rails" . | include "werf_container_env" | indent 8 }}
```

Аналогично можно пробросить секретные переменные (пароли и т.п.) и у Верфи есть специальный механизм для этого. Но к этому вопросу мы вернёмся позже.

##### Логгирование

{{ОБЯЗАТЕЛЬНО написать про то, что логи в stdout и как это сделать в этом фреймворке (не исключает использования Sentry-подобных штук, о чём будет позже)}}


##### Роутинг и заворачивание трафика на приложение

Нам надо будет пробить порт у пода, сервиса и настроить Ingress, который выступает у нас в качестве балансера.

Если вы мало работали с Kubernetes — эта часть может вызвать у вас много проблем. Большинство тех, кто начинает работать с Kubernetes по невнимательности допускают ошибки при конфигурировании labels и затем занимаются долгой и мучительной отладкой.


###### Прокрутить порт

Для того чтобы мы смогли общаться с нашим приложением извне необходимо привязать к нашему deployment объект Service.

В наш service.yaml нужно добавить: \

```
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Chart.Name }}
spec:
  selector:
    app: {{ .Chart.Name }}
  clusterIP: None
  ports:
  - name: http
    port: 8080
    protocol: TCP
```
Обязательно нужно указывать порты, на котором будет слушать наше приложение внутри контейнера.

Сама привязка к deployment происходит с помощью блока **selector:**


```
  selector:
    app: {{ .Chart.Name }}
```


Тут у нас указан лэйбл <code><em>app: {{ .Chart.Name }}</em> </code>он должен полностью совпадать с блоком <strong>labels </strong>в Deployment который мы описали выше: \



```
  template:
    metadata:
      labels:
        app: {{ $.Chart.Name }}
```


Иначе Service не поймет адрес какого пода добавлять в DNS.

###### Роутинг на Ingress

Теперь можно говорить nginx ingress куда проксировать запросы извне: \



```
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: nginx
  name: {{ .Chart.Name }}
spec:
  rules:
  - host: {{ pluck .Values.global.env .Values.domain | first }}
    http:
      paths:
      - backend:
          serviceName: {{ .Chart.Name }}
          servicePort: 8080
        path: /
```


Тут я кратко опишу только блок с **rules. **Так как именно он отвечает непосредственно за проксирование. Тут мы говорим nginx ingress прямо, что все запросы на корень `path: /` домена `{{ pluck .Values.global.env .Values.domain | first }}` нам нужно направлять в сервис с именем `{{ .Chart.Name }} `на порт `8080`.

`{{ pluck .Values.global.env .Values.domain | first }} `- это шаблон для получения домена в разных энвайроментах. Сами домены мы как раз указываем в values.yaml.

Название environment в файле values.yaml должно быть точно такое же какое мы укажем в .gitlab-ci.yaml позже.


```
domain:
  stage: stage.my-chat.com
  sroduction: my-chat.com
```


На этом написание простейшего деплоя для нашего приложения окончено.

#### Секретные переменные

Для хранения в репозитории паролей, файлов сертификатов и т.п., рекомендуется использовать подсистему работы с секретами werf.

Идея заключается в том, что конфиденциальные данные должны храниться в репозитории вместе с приложением, и должны оставаться независимыми от какого-либо конкретного сервера.


Для этого в werf существует инструмент. [Werf helm secret.](https://werf.io/documentation/reference/deploy_process/working_with_secrets.html) Мы можем зашифровать любые данные с помощью werf и спокойно положить их прямиком в репозиторий. При деплое Werf впоследствии сам их расшифрует.

Чтобы воспользоваться шифрованием нам сначала нужно создать ключ, сделать это можно так: \



```
user:~/chat$ werf helm secret generate-secret-key
ad747845284fea7135dca84bde9cff8e
user:~/chat$ export WERF_SECRET_KEY=ad747845284fea7135dca84bde9cff8e
```


После того как мы сгенерировали ключ, добавим его в переменные окружения у себя локально. \

секретные данные мы можем добавить создав рядом с values.yaml файл secret-values.yaml

Теперь использовав команду:


```
user:~/chat$ werf helm secret values edit ./helm/secret-values.yaml
```


Откроется текстовый редактор по-умолчанию, где мы сможем добавить наши секретные данные как обычно:


```
app:
  s3:
    access_key:
      _default: bNGXXCF1GF
    secret_key:
      _default: zpThy4kGeqMNSuF2gyw48cOKJMvZqtrTswAQ
```


После того как вы закроете редактор, werf зашифрует их и secret-values.yaml будет выглядеть так: \

![svalues](/werf-articles/gitlab-nodejs-files/images/secret_values.png "svalues")

#### Деплой в Gitlab CI

Далее нам нужно описать стадии выката.

В общей сложности они у нас будут отличаться только параметрами, потому мы напишем для них небольшой шаблон:

```
.base_deploy: &base_deploy
  script:
    - type multiwerf && source <(multiwerf use ${WERF_VERSION})
    - type werf && source <(werf ci-env gitlab --tagging-strategy stages-signature --verbose)
    - werf deploy --stages-storage :local
  dependencies:
    - Build
  tags:
    - article-werf
```

Скрипт стадий выката отличается от сборки всего одной командой:

```
    - werf deploy --stages-storage :local
```

И тут назревает вполне логичный вопрос - а как?

Как werf понимает куда нужно будет деплоить и каким образом? На это есть два ответа.

Первый из них вы уже видели и заключается он в команде `werf ci-env` которая берёт нужные переменные прямиком из pipeline Gitlab - и в данном случае ту что касается названия окружения.
А о втором мы поговорим чуть ниже после того как добавим сами стадии деплоя в `.gitlab-ci.yml`

```
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

Deploy to Production:
  extends: .base_deploy
  stage: deploy
  environment:
    name: production
  except:
    - schedules
  only:
    - master
```

Описание деплоя содержит в себе немного. Скрипт, указание принадлежности к стадии **deploy**, которую мы описывали в начале gitlab-ci.yml, и **dependencies** что означает что стадия не может быть запущена без успешного завершения стадии **Build**. Также мы указали с помощью **only**, ветку _master_, что означает что стадия будет доступна только из этой ветки. **environment** указали потому что werf необходимо понимать в каком окружении он работает. В дальнейшем мы покажем, как создать CI для нескольких окружений. Остальные параметры вам уже известны.

Теперь обязательно на сервере с gitlab-runner, который занимается сборкой вашего приложения, установить [kubectl](https://kubernetes.io/ru/docs/tasks/tools/install-kubectl/). И положить в домашнюю директорию пользователя gitlab-runner конфиг kubernetes в папку `.kube` (нужно её создать, если её нет)

Конфиг можно взять на мастере кластера kubernetes в `/etc/kubernetes/admin.conf`

Скопируйте его и положите в папку `.kube` переименовав файл в `config`.

![alt_text](images/-5.png "image_tooltip")

Обычно в конфиге указан прямой адрес мастера kubernetes. Соответственно нужно чтобы мастер нашего кластера был доступен по сети для gitlab-runner.

Всё это мы сделали для того чтобы werf мог общаться с API kubernetes и деплоить в него наши приложения, это и есть второй ответ на вопрос “Как werf понимает куда ему нужно деплоить?”. По умолчанию деплой будет происходить в namespace состоящий из имени проекта задаваемого в `werf.yaml` и окружения задаваемого в `.gitlab-ci.yml` куда мы деплоим наше приложение.

Ну а теперь достаточно создать Merge Request и нам будет доступна кнопка Deploy to Stage.

![alt_text](images/-6.png "image_tooltip")

Посмотреть статус выполнения pipeline можно в интерфейсе gitlab **CI / CD - Pipelines**

![alt_text](images/-7.png "image_tooltip")


Список всех окружений - доступен в меню **Operations - Environments**

![alt_text](images/-8.png "image_tooltip")

Из этого меню - можно так же быстро открыть приложение в браузере.

{{И тут в итоге должна быть картинка как аппка задеплоилась и объяснение картинки}}

# Подключаем зависимости

В nodejs в качестве менеджера зависимостей мы используем npm. А все зависимости которые нам нужны хранятся в package.json. Прямо в разделе dependencies.


```
  "dependencies": {
    "amqplib": "^0.5.5",
    "bcrypt": "^4.0.1",
```


Когда вы ведёте разработку, и вам нужно добавить какой либо пакет для того чтобы наше приложение работало достаточно у команды _npm_ прописать флаг _--save_


```
user:~/chat$ npm install cookie-parser --save
```


На самом деле этот флаг у _npm_ стоит по умолчанию, потому каждый раз когда вы устанавливаете пакет он сохраняется в package.json



*   **-P, --save-prod**: Пакет будет добавлен в **dependencies**. Значение по-умолчанию если не указаны флаги **-D** или **-O** .
*   **-D, --save-dev**: Пакет появится в **devDependencies**. (Пакеты нужные только для разработки, но не для работы приложения)
*   **-O, --save-optional**: Будет сохранено в **optionalDependencies**.(Пакеты без которых приложение всё равно будет работать)
*   **--no-save**: Не сохранять ни в какие **dependencies**.

Ну а так как мы уже подключили наш файл на отслеживание изменений в другой главе, мы можем быть уверены, что стадия установки зависимостей будет перезапущена:

Код из werf.yaml:


```
  stageDependencies:
    install:
    - package.json
```

# Генерируем и раздаем ассеты

Один из самых популярных способов генерации ассетов в nodejs является [webpack](https://webpack.js.org/).

Обычно запуск генерации с webpack вставляют сразу в package.json как скрипт \



```
  "scripts": {
    "test": "echo \"Error: no test specified\" && exit 1",
    "build": "webpack --mode production",
    "watch": "webpack --mode development --watch",
    "start": "webpack-dev-server --mode development --open",
    "clear": "del-cli dist",
    "migrate": "node-pg-migrate"
  },
```


Используя несколько режимов для удобной разработки, но конечный вариант который идёт в kubernetes (в любое из окружений) всегда сборка как в production. Остальная отладка производится только локально. Почему необходимо чтобы сборка всегда была одинаковой?

Потому что наш docker image должен всегда поддерживать иммутабельность. Создано это для того чтобы мы всегда были уверены в нашей сборке, т.е. образ оттестированный на stage окружении должен попадать в production точно таким же. Для всего остального мира наше приложение должно быть чёрным ящиком, которое лишь может принимать параметры.

Особенно это касается готовых сгенерированных ассетов.


```
<script>
    var login_url = "https://login.example.ru/oauth/ae?client_id=my-chat.ru&response_type=code&redirect_uri=https://my-chat.ru/oauth&scope=openid+profile";
    function go() {
        window.location.href = login_url
    }
    document.cookie = "chatredirect" + "=" + escape(document.location.href);    
</script>
```


Это пример неправильно сгенерированного ассета. Потому что в зависимости от окружения мы можем подключаться к разным ресурсам, а в таком случае у нас всегда будет одна и так же ссылка с одними и теми же параметрами. \
Для того чтобы такого не случалось, все ссылки ведущие на ваше приложение следует сделать относительными. И либо сделать так чтобы все ссылки и на ваше приложение и на удаленные ресурсы конфигурировались. 

Сделать это можно несколькими способами. \
1. Мы можем динамически в зависимости от окружение монтировать в контейнер с нашим приложением json с нужными параметрами. Для этого нам нужно создать объект configmap в .helm/templates. \
10-app-config.yaml


```
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Chart.Name }}-config
data:
  config.json: |-
   {
    "domain": "{{ pluck .Values.global.env .Values.domain | first }}",
    "loginUrl": "{{ pluck .Values.global.env .Values.loginUrl | first }}"
   }
```


А затем примонтировать к нашему приложению в то место где мы могли получать его по запросу от клиента: \
Код из 01-app.yaml:


```
       volumeMounts:
          - name: app-config
            mountPath: /app/dist/config.json
            subPath: config.json
      volumes:
        - name: app-config
          configMap:
            name: {{ .Chart.Name }}-config
```


2. И второй вариант перед запуском приложения получать конфиги например из [consul](https://www.consul.io/). Подробно не будем расписывать данный вариант, так как он достоин отдельной главы. Дадим лишь два совета: \
 



*   Можно запускать его перед запуском приложения добавив в  `command: ["node","/app/src/js/index.js"] `его запуск через ‘&&’ как в синтаксисе любого shell языка:  \


    ```
command: 
- /usr/bin/bash
- -c
- --
- "consul kv get app/config/urls && node /app/src/js/index.js"]
```


*   Либо добавив его как [init-container](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/) и подключив между инит контейнером и основным контейнером общий [volume](https://kubernetes.io/docs/concepts/storage/volumes/), это означает что просто нужно смонтировать volume в оба контейнера. Например [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir). Тем самым консул отработав в инит контейнере, сохранит конфиг в наш volume, а приложение из основного контейнера просто заберёт этот конфиг. \


## Какой сценарий сборки ассетов

{{Как происходит сборка ассетов, описан ли где-то этот процесс в виде кода? Или он прошит во фреймворке и нужно просто следовать каким-то правилам, и если это так — то где прочитать эти правила?}}

## Как конфигурируем сценарий сборки?

{{Про то, что есть два типа опций при сборке:}}

{{*   Те, что будут зависеть от окружения. И привести примеры таких вещей для вашего случая (урлики какие-нибудь в конфигах). }}

{{И объясняем, как мы подходим к конфигурированию таких вещей концептуально.}}

{{*   Те, что не зависят от окружения. И привести пример для вашего случая (например, это бывает волшебный хэшик для обхода кэша браузера) }}

{{И объясняем, как мы подходим к конфигурированию таких вещей концептуально.}}

{{*   Что-то ещё???}}

## Какие изменения необходимо внести

{{Концептуальное описание того, что у нас теперь два разных пода, два разных контейнера и как мы будем распределять между ними запросы и вот это всё.}}

### Изменения в сборке

{{Какие вносим изменения в стадию сборки. Подробности о том, как конкретно пробрасываем конфиги в стадию сборки.}}

### Изменения в деплое

{{Какие вносим изменения в стадию деплоя.}}

### Изменения в роутинге

{{Про ингрессы коротко.}}

# Работа с файлами и электронной почтой

 разработке может встретиться возможность когда требуется сохранять загружаемые пользователями файлы. Встает резонный вопрос о том каким образом их нужно хранить, и как после этого получать. \
	Первый и более общий способ. Это использовать как volume в подах [NFS](https://kubernetes.io/docs/concepts/storage/volumes/#nfs), [CephFS](https://kubernetes.io/docs/concepts/storage/volumes/#cephfs) или [hostPath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath), который будет направлен на директорию на ноде, куда будет подключено одно из сетевых хранилищ.

Мы не рекомендуем этот способ, потому что при возникновении неполадок с такими типами volume’ов мы будем влиять на работоспособность всего докера, контейнера и демона docker в целом, тем самым могут пострадать приложения которые даже не имеют отношения к вашему приложению. \
	Мы рекомендуем пользоваться S3. Такой способ выглядит намного надежнее засчет того что мы используем отдельный сервис, который имеет свойства масштабироваться, работать в HA режиме, и будет иметь высокую доступность.  \
Есть cloud решения S3, такие как AWS S3, Google Cloud Storage, Microsoft Blobs Storage и т.д. которые будут самым надежным решением из всех что мы можем использовать. \
Но для того чтобы просто посмотреть на то как работать с S3 или построить собственное решение, хватит Minio.


## Подключаем наше приложение к S3 Minio \


Сначала с помощью npm устанавливает пакет, который так и называется Minio.


```
user:~/chat$ npm install minio --save
```


После в исходники в src/js/index.js мы добавляем следующие строки: \



```
const Minio = require("minio");
const S3_ENDPOINT = process.env.S3_ENDPOINT || "127.0.0.1";
const S3_PORT = Number(process.env.S3_PORT) || 9000;
const TMP_S3_SSL = process.env.S3_SSL || "true";
const S3_SSL = TMP_S3_SSL.toLowerCase() == "true";
const S3_ACCESS_KEY = process.env.S3_ACCESS_KEY || "SECRET";
const S3_SECRET_KEY = process.env.S3_SECRET_KEY || "SECRET";
const S3_BUCKET = process.env.S3_BUCKET || "avatars";
const CDN_PREFIX = process.env.CDN_PREFIX || "http://127.0.0.1:9000";

// S3 client
var s3Client = new Minio.Client({
  endPoint: S3_ENDPOINT,
  port: S3_PORT,
  useSSL: S3_SSL,
  accessKey: S3_ACCESS_KEY,
  secretKey: S3_SECRET_KEY,
});
```


И этого вполне хватит для того чтобы вы могли использовать minio S3 в вашем приложении.

Полный пример использования можно посмотреть в тут.

Важно заметить, что мы не указываем жестко параметры подключения прямо в коде, а производим попытку получения их из переменных. Точно так же как и с генерацией статики тут нельзя допускать фиксированных значений. \
 \
Остается только настроить наше приложение со стороны Helm.

Добавляем значения в values.yaml


```
app:
  s3:
    host:
      _default: chat-test-minio
    port:
      _default: 9000
    bucket:
      _default: 'avatars'
    ssl:
      _default: 'false'
```
И в secret-values.yaml

```
app:
  s3:
    access_key:
      _default: bNGXXCF1GF
    secret_key:
      _default: zpThy4kGeqMNSuF2gyw48cOKJMvZqtrTswAQ
```
Далее мы добавляем переменные непосредственно в Deployment с нашим приложением: \



```
        - name: CDN_PREFIX
          value: {{ printf "%s%s" (pluck .Values.global.env .Values.app.cdn_prefix | first | default .Values.app.cdn_prefix._default) (pluck .Values.global.env .Values.app.s3.bucket | first | default .Values.app.s3.bucket._default) | quote }}
        - name: S3_SSL
          value: {{ pluck .Values.global.env .Values.app.s3.ssl | first | default .Values.app.s3.ssl._default | quote }}
        - name: S3_ENDPOINT
          value: {{ pluck .Values.global.env .Values.app.s3.host | first | default .Values.app.s3.host._default }}
        - name: S3_PORT
          value: {{ pluck .Values.global.env .Values.app.s3.port | first | default .Values.app.s3.port._default | quote }}
        - name: S3_ACCESS_KEY
          value: {{ pluck .Values.global.env .Values.app.s3.access_key | first | default .Values.app.s3.access_key._default }}
        - name: S3_SECRET_KEY
          value: {{ pluck .Values.global.env .Values.app.s3.secret_key | first | default .Values.app.s3.secret_key._default }}
        - name: S3_BUCKET
          value: {{ pluck .Values.global.env .Values.app.s3.bucket | first | default .Values.app.s3.bucket._default }}
```
Тот способ которым мы передаем переменные в под, с помощью GO шаблонов означает: \
	1. `{{ pluck .Values.global.env .Values.app.s3.access_key | first `пробуем взять из поля _app.s3.access_key_ значение из поля которое равно environment.

            2.  `default .Values.app.s3.access_key._default }} `и если такого нет, то мы берём значение из поля _default.

И всё, этого достаточно!

# Подключаем redis

Допустим к нашему приложению нужно подключить простейшую базу данных, например, redis или memcached. Возьмем первый вариант.

В простейшем случае нет необходимости вносить изменения в сборку — всё уже собрано для нас. Надо просто подключить нужный образ, а потом в вашем {{Python}} приложении корректно обратиться к этому приложению.

## Завести Redis в Kubernetes

Есть два способа подключить: прописать helm-чарт самостоятельно или подключить внешний. Мы рассмотрим второй вариант.

Подключим redis как внешний subchart.

Для этого нужно:

1. прописать изменения в yaml файлы; 
2. указать редису конфиги
3. подсказать werf, что ему нужно подтягивать subchart.

Добавим в файл `.helm/requirements.yaml` следующие изменения:

```
dependencies:
- name: redis
  version: 9.3.2
  repository: https://kubernetes-charts.storage.googleapis.com/
  condition: redis.enabled
```

Для того чтобы werf при деплое загрузил необходимые нам сабчарты - нужно добавить команды в `.gitlab-ci`

```
.base_deploy:
  stage: deploy
  script:
    - werf helm repo init
    - werf helm dependency update
    - werf deploy
```

Опишем параметры для redis в файле `.helm/values.yaml`

```
redis:
  enabled: true
```

При использовании сабчарта по умолчанию создается master-slave кластер redis. 

Если посмотреть на рендер (`werf helm render`) нашего приложения с включенным сабчартом для redis, то можем увидеть какие будут созданы сервисы:

```
# Source: example-2/charts/redis/templates/redis-master-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: chat-stage-redis-master

# Source: example-2/charts/redis/templates/redis-slave-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: chat-stage-redis-slave
```

## Подключение Nodejs приложения к базе redis

В нашем приложении - мы будем  подключаться к мастер узлу редиса. Нам нужно, чтобы при выкате в любое окружение приложение подключалось к правильному редису.
В src/js/index.js мы добавляем:


```
const REDIS_URI = process.env.SESSION_REDIS || "redis://127.0.0.1:6379";
const SESSION_TTL = process.env.SESSION_TTL || 3600;
const COOKIE_SECRET = process.env.COOKIE_SECRET || "supersecret";
// Redis connect
const expSession = require("express-session");
const redis = require("redis");
let redisClient = redis.createClient(REDIS_URI);
let redisStore = require("connect-redis")(expSession);

var session = expSession({
  store: new redisStore({ client: redisClient, ttl: SESSION_TTL }),
  secret: "keyboard cat",
  resave: false,
  saveUninitialized: false,
});
var sharedsession = require("express-socket.io-session");
app.use(session);
```

Добавляем в values.yaml

values.yaml


```
app:
  redis:
     host:
       master:
         stage: chat-stage-redis-master
       slave:
         stage: chat-stage-redis-slave
```


secret-values.yaml


```
app:
 redis:
    password:
      _default: 100067e35229a23c5070ad5407b7406a7d58d4e54ecfa7b58a1072bc6c34cd5d443e
```


И наконец добавляем подстановку переменных в сам манифест с нашим приложением: \



```
      - name: SESSION_REDIS
        value: "redis://root:{{ pluck .Values.global.env .Values.app.redis.password | first | default .Values.app.redis.password._default }}@{{ pluck .Values.global.env .Values.app.redis.host.master | first }}:6379"
```



В данном случае Redis подключается как хранилище для сессий.

# Подключаем базу данных

{{Для текущего примера в приложении должны быть установлены необходимые зависимости. В качестве примера - мы возьмем приложение для работы которого необходима база данных.}}


## Как подключить БД

Мы можем использовать базу данных в нашем приложении для хранения любой информации, и подключение базы данных не будет сильно отличаться от предыдущих примеров: \



```
const pgconnectionString =
  process.env.DATABASE_URL || "postgresql://127.0.0.1/postgres";

// Postgres connect
const pool = new pg.Pool({
  connectionString: pgconnectionString,
});
pool.on("error", (err, client) => {
  console.error("Unexpected error on idle client", err);
  process.exit(-1);
});
```


В данном случае мы также используем сабчарт для деплоя базы из того же репозитория. Этого должно хватить для нашего небольшого приложения. В случае большой высоконагруженной инфраструктуры деплой базы непосредственно в кубернетес не рекомендуется. 

Добавляем информацию о подключении в values.yaml


```
app:
 postgresql:
    host:
      _default: chat-test-postgresql
    user: 
      _default: chat
    db: 
      _default: chat
```


И в secret-values.yaml


```
app:
  postgresql:
    password:
      _default: 1000acb579eaee19bec317079a014346d6aab66bbf84e4a96b395d4a5e669bc32dd1
```


Далее привносим подключени внутрь манифеста нашего приложения: \



```
       - name: DATABASE_URL
          value: "postgresql://{{ pluck .Values.global.env .Values.app.postgresql.user | first | default .Values.app.postgresql.user._default }}:{{ pluck .Values.global.env .Values.app.postgresql.password | first | default .Values.app.postgresql.password._default }}@{{ pluck .Values.global.env .Values.app.postgresql.host | first | default .Values.app.postgresql.host._default }}:5432/{{ pluck .Values.global.env .Values.app.postgresql.db | first | default .Values.app.postgresql.db._default }}"
```


## Выполнение миграций

Миграции в nodejs можно вызывать с помощью скрипта в npm: \



```
/* eslint-disable camelcase */

exports.shorthands = undefined;

exports.up = pgm => {
    pgm.createTable('users', {
        id: 'id',
        name: { type: 'varchar(255)', notNull: true, unique: true },
        email: { type: 'varchar(255)', notNull: true, unique: true },
        password: { type: 'varchar(60)', notNull: true },
        createdAt: {
          type: 'timestamp',
          notNull: true,
          default: pgm.func('current_timestamp'),
        },
      })
    pgm.createTable('messages', {
        id: 'id',
        userId: {
          type: 'integer',
          notNull: true,
          references: '"users"',
          onDelete: 'cascade',
        },
        body: { type: 'text', notNull: true },
        createdAt: {
          type: 'timestamp',
          notNull: true,
          default: pgm.func('current_timestamp'),
        },
      })
      pgm.createIndex('messages', 'userId')
      pgm.createIndex('messages', 'createdAt')
};

exports.down = pgm => {
    pgm.dropTable('messages')
    pgm.dropTable('users')
};
```

{{Объясняем, куда мы это вписываем вот во всех этих сборках-деплоях и почему именно туда}}

## Накатка фикстур при первом выкате

{{Объясняем, куда это вписывать во всех тих сборках-деплоях.}}


# Юнит-тесты и Линтеры

{{Говорим, что за юнит-тесты/линтеры в этом фреймворке есть?}}

{{Объясняем, куда мы прописываем запуск этих тестов-линтеров и почему именно туда.}}

# Несколько приложений в одной репе

{{1. Добавляем кронджоб}}
{{2. Добавляем воркер/консюмер}}
{{3. Добавляем вторую приложуху на другом языке (например, это может быть webscoket’ы на nodejs; показать организацию helm, организацию werf.yaml, и ссылку на другую статью)}}

# Динамические окружения

Если для командной работы большой группе разработчиков необходимо проверять и делиться своими разработками с другими членами команды - можно воспользоваться динамическими окружениями.

Плюсом использования такого подхода - является то что если у нас приложение уже подготовлено для запуска в kubernetes - то нам нужно только добавить несколько стадий в ci

Рассмотрим пример деплоя


```
Deploy to future:
  extends: .base_deploy
  stage: deploy
  environment:
    name: ${CI_COMMIT_REF_SLUG}
    url: http://${CI_COMMIT_REF_SLUG}.k8s.example.com
  only:
  - future/*
  when: manual
```


При таком ci - мы можем выкатывать каждую ветку future/* в отдельный namespace с изолированной базой данных, накатом необходимых миграций и например проводить тесты для данного окружения.

В репозитории с примерами будет реализовано отдельное приложение которое показывает реализацию данного подхода.

