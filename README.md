# dapp [![Gem Version](https://badge.fury.io/rb/dapp.svg)](https://badge.fury.io/rb/dapp) [![Build Status](https://travis-ci.org/flant/dapp.svg)](https://travis-ci.org/flant/dapp) [![Code Climate](https://codeclimate.com/github/flant/dapp/badges/gpa.svg)](https://codeclimate.com/github/flant/dapp) [![Test Coverage](https://codeclimate.com/github/flant/dapp/badges/coverage.svg)](https://codeclimate.com/github/flant/dapp/coverage)

## Reference

### Dappfile

#### Основное

##### name \<name\>
Базовое имя для собираемых docker image`ей: <базовое имя>-dappstage:<signature>
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
    shell.infra_install 'apt-get install service'
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

#### Артифакты
*TODO*

#### Docker
*TODO*

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
  * attributes/\<stage\>/ -> attributes/default/

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
  * attributes/\<stage\>/ -> attributes/default/

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
Собрать приложения, удовлетворяющие хотя бы одному из **PATTERN**-ов (по умолчанию *).

```
dapp build [options] [PATTERN ...]
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
Выкатить собранное приложение с именем **REPO**.

```
dapp push [options] [PATTERN...] REPO
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
* Выкатить приложение **app** в репозиторий test, именем myapp и тегом latest:
```bash
$ dapp push app test/myapp
```
* Выкатить приложение с произвольными тегами:
```bash
$ dapp push app test/myapp --tag 1 --tag test
```
* Запустить вхолостую и посмотреть какие образы могут быть выкачены:
```bash
$ dapp push app test/myapp --tag-commit --tag-branch --dry-run
test/myapp:2c622c16c39d4938dcdf7f5c08f7ed4efa8384c4
test/myapp:master
```

#### dapp smartpush
Выкатить каждое собранное приложение с именем **REPOPREFIX**/имя приложения.

```
dapp smartpush [options] [PATTERN ...] REPOPREFIX
```

Опции такие же как у **dapp push**.

##### Примеры
* Выкатить все приложения в репозиторий test и тегом latest:
```bash
$ dapp smartpush test
```
* Запустить вхолостую и посмотреть какие образы могут быть выкачены:
```bash
$ dapp smartpush test --tag yellow --tag-branch --dry-run
backend
  test/app:yellow
  test/app:master
frontend
  test/app:yellow
  test/app:0.2
```

#### dapp list
Вывести список приложений.

```
dapp list [options] [PATTERN ...]
```

#### dapp run
Запустить собранное приложение с докерными аргументами **DOCKER ARGS**.

```
dapp run [options] [PATTERN...] [DOCKER ARGS]
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
dapp stages flush [options] [PATTERN...]
```

#### dapp stages cleanup
Удаляет все нетегированный кэш приложений (см. [Кэш стадий](#Кэш-стадий)).

```
dapp stages cleanup [options] [PATTERN...]
```

## Architecture

### Стадии
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
