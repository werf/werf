# dapp [![Gem Version](https://badge.fury.io/rb/dapp.svg)](https://badge.fury.io/rb/dapp) [![Build Status](https://travis-ci.org/flant/dapp.svg)](https://travis-ci.org/flant/dapp) [![Code Climate](https://codeclimate.com/github/flant/dapp/badges/gpa.svg)](https://codeclimate.com/github/flant/dapp) [![Test Coverage](https://codeclimate.com/github/flant/dapp/badges/coverage.svg)](https://codeclimate.com/github/flant/dapp/coverage)

## Reference

### Основные определения

#### Проект (project)
Проект — это директория, содержащая приложение или набор приложений (см. приложение).
* Приложение может находиться в корне проекта (в этом случае в корне проекта лежит соответствующий Dappfile).
* В случае, если в проекте есть несколько приложений — они находятся в директориях .dapps/\<имя-приложения\>/ (в каждой из которых есть соответствующий Dappfile).

#### Корень проекта
Корень проекта — это директория, в которой находится директория .dapps или, при ее отсутствии — это директория, содержащая Dappfile.

#### Имя проекта
Имя проекта — это последний элемент пути к git репозиторию из параметра конфигурации remote.origin.url или, при отсутствии git или параметра конфигурации remote.origin.url — имя директории корня проекта.

#### Приложение (application)
Приложение — это набор правил, объединенных в одном Dappfile, по которым происходит сборка одного или нескольких docker образов.
* В рамках одного приложения может быть описано дерево подприложений.
* При сборке дерева подприложений, docker образы будут собраны для всех подприложений листьев описанного дерева.

#### Базовое имя приложения (basename)
Базовое имя приложения ­— это имя, связанное с каждым Dappfile.
* По умолчанию базовое имя приложения ­— это имя директории, в которой находится Dappfile.
* Базовое имя приложения может быть переопределено в Dappfile (см. name).

#### Подприложение (app)
Подприложение — это средство группировки правил сборки в иерархию с наследованием.
* Подприложение наследует правила сборки того подприложения, в котором оно объявлено, и глобальные правила сборки.
* Сборка docker образов осуществляется только для тех подприложений, которые являются листьями в описанном дереве приложений.
* Для каждого подприложения в Dappfile указывается имя (см. app). При этом вложенные подприложения наследуют имена родительских подприложений и базовое имя приложения (см. базовое имя приложения).

#### Стадия (stage)
Стадия — это именованный набор инструкций для сборки docker образа.
* Собранное приложение представляет собой цепочку связанных стадий.
* TODO: правила именования

#### Стадии
TODO
from
before install
git artifact archive
install
git artifact pre install patch dependencies
git artifact pre install patch
install
git artifact post install patch
artifact
before setup
setup
git artifact pre setup patch
setup
chef cookbooks
git artifact post setup patch
git artifact latest patch
docker instructions

### Dappfile

#### Основное

##### name \<name\>
Базовое имя для собираемых docker image`ей: \<базовое имя\>-dappstage:\<signature\>.

Опционально, по умолчанию определяется исходя из имени директории, в которой находится Dappfile.

##### install\_depends\_on \<glob\>[,\<glob\>, \<glob\>, ...]
Список файлов зависимостей для стадии install.

* При изменении содержимого указанных файлов, произойдет пересборка стадии install.
* Учитывается лишь содержимое файлов и порядок в котором они указаны (имена файлов не учитываются).
* Поддерживаются glob-паттерны.
* Директории игнорируются.

##### setup\_depends\_on \<glob\>[,\<glob\>, \<glob\>, ...]
Список файлов зависимостей для стадии setup.

* При изменении содержимого указанных файлов, произойдет пересборка стадии setup.
* Учитывается лишь содержимое файлов и порядок в котором они указаны (имена файлов не учитываются).
* Поддерживаются glob-паттерны.
* Директории игнорируются.

##### builder \<builder\>
Тип сборки: **:chef** или **:shell**.
* Опционально, по умолчанию будет выбран тот сборщик, который будет использован первым (см. [Chef](#chef), [Shell](#shell)).
* При определении типа сборки, сборщик другого типа сбрасывается.
* При смене сборщика, необходимо явно изменять тип сборки.
* Пример:
  * Собирать приложение X с **:chef** сборщиком, а Y c **:shell**:
  ```ruby
  app 'X' do
    chef.module 'a', 'b'
  end
  app 'Y' do
    shell.before_install 'apt-get install service'
  end
  ```
  * Собирать приложения X-Y и Z с **:chef** сборщиком, а X-V c **:shell**:
  ```ruby
  chef.module 'a', 'b'
  
  app 'X' do
    builder :shell
   
    app 'Y' do
      builder :chef
      chef.module 'c'
    end
    
    app 'V' do
      shell.install 'application install'
    end
  end
  app 'Z'
  ```
    
##### app \<app\>[, &blk]
Определяет приложение <app> для сборки.

* Опционально, по умолчанию будет использоваться приложение с базовым именем (см. name \<name\>).
* Можно определять несколько приложений в одном Dappfile.
* При использовании блока создается новый контекст.
  * Наследуются все настройки родительского контекста.
  * Можно дополнять или переопределять настройки родительского контекста.
  * Можно использовать директиву app внутри нового контекста.
* При использовании вложенных вызовов директивы, будут использоваться только приложения указанные последними в иерархии. Другими словами, в описанном дереве приложений будут использованы только листья.
  * Правила именования вложенных приложений: \<app\>[-\<subapp\>-\<subsubapp\>...]
* Примеры:
  * Собирать приложения X и Y:
  ```ruby
  app 'X'
  app 'Y'
  ```
  * Собирать приложения X, Y-Z и Y-K-M:
  ```ruby
  app 'X'
  app 'Y' do
    app 'Z'

    app 'K' do
      app 'M'
    end
  end
  ```

#### Артефакты
*TODO*

#### Docker

##### docker.from \<image\>[, cache_version: \<cache_version\>]
Определить окружение приложения **\<image\>** (см. [Стадия from](#from)).

**\<image\>** имеет следующий формат 'REPOSITORY:TAG'.

Опциональный параметр **\<cache_version\>** участвует в формировании сигнатуры стадии.

##### docker.cmd \<cmd\>[, \<cmd\> ...]
Применить dockerfile инструкцию CMD (см. [CMD](https://docs.docker.com/engine/reference/builder/#/cmd "Docker reference")).

##### docker.env \<env_name\>: \<env_value\>[, \<env_name\>: \<env_value\> ...]
Применить dockerfile инструкцию ENV (см. [ENV](https://docs.docker.com/engine/reference/builder/#/env "Docker reference")).

##### docker.entrypoint \<cmd\>[, \<arg\> ...]
Применить dockerfile инструкцию ENTRYPOINT (см. [ENTRYPOINT](https://docs.docker.com/engine/reference/builder/#/entrypoint "Docker reference")).

##### docker.expose \<expose\>[, \<expose\> ...]
Применить dockerfile инструкцию EXPOSE (см. [EXPOSE](https://docs.docker.com/engine/reference/builder/#/expose "Docker reference")).

##### docker.label \<label_key\>: \<label_value\>[, \<label_key\>: \<label_value\> ...]
Применить dockerfile инструкцию LABEL (см. [LABEL](https://docs.docker.com/engine/reference/builder/#/label "Docker reference")).

##### docker.onbuild \<cmd\>[, \<cmd\> ...]
Применить dockerfile инструкцию ONBUILD (см. [ONBUILD](https://docs.docker.com/engine/reference/builder/#/onbuild "Docker reference")).

##### docker.user \<user\>
Применить dockerfile инструкцию USER (см. [USER](https://docs.docker.com/engine/reference/builder/#/user "Docker reference")).

##### docker.volume \<volume\>[, \<volume\> ...]
Применить dockerfile инструкцию VOLUME (см. [VOLUME](https://docs.docker.com/engine/reference/builder/#/volume "Docker reference")).

##### docker.workdir \<path\>
Применить dockerfile инструкцию WORKDIR (см. [WORKDIR](https://docs.docker.com/engine/reference/builder/#/workdir "Docker reference")).

#### Shell
*TODO*

#### Chef

##### chef.module \<mod\>[, \<mod\>, \<mod\> ...]
Включить переданные модули для chef builder в данном контексте.

* Для каждого переданного модуля может существовать по одному рецепту на каждую из стадий.
* Файл рецепта для \<stage\>: recipes/\<stage\>.rb
* При отсутствии файла рецепта в runlist для данной стадии используется пустой рецепт \<mod\>::void.
* Порядок вызова рецептов модулей в runlist совпадает порядком их описания в конфиге.
* При сборке стадии, для каждого из включенных модулей используются файлы cookbook`а:
  * recipes/\<stage\>.rb (если существует)
  * metadata.json
  * files/\<stage\>/ -> files/default/
  * templates/\<stage\>/ -> templates/default/
  * attributes/common/ -> attributes/
  * attributes/\<stage\>/ -> attributes/
* В attributes/common и attributes/\<stage\> не могут быть файлы с одинаковыми именами, т.к. они попадают в одну директорию для запуска стадии.

##### chef.skip_module \<mod\>[, \<mod\>, \<mod\> ...]
Выключить переданные модули для chef builder в данном контексте.

##### chef.reset_modules
Выключить все модули для chef builder в данном контексте.

##### chef.recipe \<recipe\>[, \<recipe\>, \<recipe\> ...]
Включить переданные рецепты из проекта для chef builder в данном контексте.

* Для каждого преданного рецепта может существовать файл рецепта в проекте на каждую из стадий.
* Файл рецепта для \<stage\>: recipes/\<stage\>/\<recipe\>.rb
* При отсутствии хотя бы одного файла рецепта из включенных, в runlist для данной стадии используется пустой рецепт \<projectname\>::void.
* Порядок вызова рецептов в runlist совпадает порядком их описания в конфиге.
* При сборке стадии, используются файлы cookbook`а:
  * для каждого из включенных рецептов:
    * recipes/\<stage\>/\<recipe\>.rb -> recipes/\<recipe\>.rb
  * metadata.json
  * files/\<stage\> -> files/default/
  * templates/\<stage\>/ -> templates/default/
  * attributes/common/ -> attributes/
  * attributes/\<stage\>/ -> attributes/
* В attributes/common и attributes/\<stage\> не могут быть файлы с одинаковыми именами, т.к. они попадают в одну директорию для запуска стадии.

##### chef.remove_recipe \<recipe\>[, \<recipe\>, \<recipe\> ...]
Выключить переданные рецепты из проекта для chef builder в данном контексте.

##### chef.reset_recipes
Выключить все рецепты из проекта для chef builder в данном контексте.

##### chef.reset_all
Выключить все рецепты из проекта и все модули для chef builder в данном контексте.

##### Примеры
* [Dappfile](doc/example/Dappfile.chef.1)

### Команды

#### dapp build
Собрать приложения, удовлетворяющие хотя бы одному из **APPS PATTERN**-ов (по умолчанию *).

```
dapp build [options] [APPS PATTERN ...]
```

##### Опции среды сборки

###### --dir PATH
Определяет директорию, которая используется при поиске одного или нескольких **Dappfile**.

По умолчанию поиск ведётся в текущей папке пользователя.

###### --build-dir PATH
Переопределяет директорию хранения кеша, который может использоваться между сборками.

###### --tmp-dir-prefix PREFIX
Переопределяет префикс временной директории, файлы которой используются только во время сборки.

##### Опции логирования

###### --dry-run
Позволяет запустить сборщик вхолостую и посмотреть процесс сборки.

###### --verbose
Подробный вывод.

###### --color MODE
Отвечает за регулирование цвета при выводе в терминал.

Существует несколько режимов (**MODE**): **on**, **of**, **auto**.

По умолчанию используется **auto**, который окрашивает вывод, если вывод производится непосредственно в терминал.

###### --time
Добавляет время каждому событию лога.

##### Опции интроспекции
Позволяют поработать с образом на определённом этапе сборки.

###### --introspect-stage STAGE
После успешного прохождения стадии **STAGE**.

###### --introspect-before-error
Перед выполением команд несобравшейся стадии.

###### --introspect-error
После завершения команд стадии с ошибкой.

##### Примеры
* Сборка в текущей директории:
```bash
$ dapp build
```
* Сборка приложений из соседней директории:
```bash
$ dapp build --dir ../project
```
* Запуск вхолостую с подробным выводом процесса сборки:
```bash
$ dapp build --dry-run --verbose
```
* Выполнить сборку, а в случае ошибки, предоставить образ для тестирования:
```bash
$ dapp build --introspect-error
```

#### dapp push
Выкатить собранное приложение в репозиторий, в следующем формате **REPO**:**ИМЯ ПРИЛОЖЕНИЯ**-**TAG**.

```
dapp push [options] [APPS PATTERN ...] REPO
```

##### --force
Позволяет перезаписывать существующие образы.

##### Опции тегирования
Отвечают за тег(и), с которыми выкатывается приложение.

Могут быть использованы совместно и по несколько раз.

В случае отсутствия, используется тег **latest**.

###### --tag TAG
Добавляет произвольный тег **TAG**.

###### --tag-branch
Добавляет тег с именем ветки сборки. 

###### --tag-commit
Добавляет тег с коммитом сборки. 

###### --tag-build-id
Добавляет тег с идентификатором сборки (CI).

###### --tag-ci
Добавляет теги, взятые из переменных окружения CI систем.

##### Примеры
* Выкатить все приложения в репозиторий test и тегом latest:
```bash
$ dapp push test
```
* Запустить вхолостую и посмотреть какие образы могут быть выкачены:
```bash
$ dapp push test --tag yellow --tag-branch --dry-run
backend
  test:backend-yellow
  test:backend-master
frontend
  test:frontend-yellow
  test:frontend-0.2
```

#### dapp spush
Выкатить собранное приложение в репозиторий, в следующем формате **REPO**:**TAG**.

```
dapp spush [options] [APP PATTERN] REPO
```

Опции такие же как у **dapp push**.

##### Примеры
* Выкатить приложение **app** в репозиторий test, именем myapp и тегом latest:
```bash
$ dapp spush app test/myapp
```
* Выкатить приложение с произвольными тегами:
```bash
$ dapp spush app test/myapp --tag 1 --tag test
```
* Запустить вхолостую и посмотреть какие образы могут быть выкачены:
```bash
$ dapp spush app test/myapp --tag-commit --tag-branch --dry-run
test/myapp:2c622c16c39d4938dcdf7f5c08f7ed4efa8384c4
test/myapp:master
```

#### dapp list
Вывести список приложений.

```
dapp list [options] [APPS PATTERN ...]
```

#### dapp run
Запустить собранное приложение с докерными аргументами **DOCKER ARGS**.

```
dapp run [options] [APPS PATTERN...] [DOCKER ARGS]
```

##### [DOCKER ARGS]
Может содержать докерные опции и/или команду.

Перед командой необходимо использовать группу символов ' -- '.

##### Примеры
* Запустить приложение с опциями:
```bash
$ dapp run -ti --rm
```
* Запустить с опциями и командами:
```bash
$ dapp run -ti --rm -- bash -ec true
```
* Запустить, передав только команды:
```bash
$ dapp run -- bash -ec true
```
* Посмотреть, что может быть запущено:
```bash
$ dapp run app -ti --rm -- bash -ec true
docker run -ti --rm app-dappstage:ea5ec7543c809ec7e9fe28181edfcb2ee6f48efaa680f67bf23a0fc0057ea54c bash -ec true
```

#### dapp stages flush
Удаляет весь тегированный кэш приложений (см. [Кэш стадий](#Кэш-стадий)).

```
dapp stages flush [options] [APPS PATTERN...]
```

## Architecture

### Стадии
| Имя                 | Краткое описание                     															                                        |
| ------------------- | --------------------------------------------------------------------------------------------------------- |
| from                | Выбор окружения                                 														                              |
| before_install       | Установка софта инфраструктуры                															                              |
| git_artifact_archive    | Создание архива                                															                              |
| git_artifact_pre_install_patch            | Наложение патча                              															                                |
| install             | Установка софта приложения                    															                              |
| artifact            | Копирование артефакта(ов)                     															                              |
| git_artifact_post_install_patch            | Наложение патча                               															                              |
| before_setup         | Настройка софта инфраструктуры                															                              |
| git_artifact_pre_setup_patch            | Наложение патча                               															                              |
| chef_cookbooks      | Установка cookbook`ов         																			                                      |
| setup               | Развёртывание приложения                    															                                |  
| git_artifact_post_setup_patch            | Наложение патча                               															                              |
| git_artifact_latest_patch            | Наложение патча                               	                                                          |
| docker_instructions | Применение докерфайловых инструкций (CMD, ENTRYPOINT, ENV, EXPOSE, LABEL, ONBUILD, USER, VOLUME, WORKDIR) |

#### from
*TODO*

#### before_install
*TODO*

#### git_artifact_archive
*TODO*

#### git_artifact_pre_install_patch
*TODO*

#### install
*TODO*

#### artifact
*TODO*

#### git_artifact_post_install_patch
*TODO*

#### before_setup
*TODO*

#### git_artifact_pre_setup_patch
*TODO*

#### chef_cookbooks
*TODO*

#### setup
*TODO*

#### git_artifact_post_setup_patch
*TODO*

#### git_artifact_latest_patch
*TODO*

#### docker_instructions
*TODO*

### Хранение данных

#### Кэш стадий
*TODO*

#### Временное
*TODO*

#### Метаданные
*TODO*

#### Кэш сборки
*TODO*
