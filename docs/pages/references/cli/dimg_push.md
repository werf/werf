---
title: dapp dimg push
sidebar: reference
permalink: dimg_push.html
folder: command
---

### dapp dimg push
Выкатить собранные dimg-ы в docker registry, в формате **REPO**/**ИМЯ DIMG**:**TAG**. В случае, если **ИМЯ DIMG** отсутствует, формат следующий **REPO**:**TAG**.

```
dapp dimg push [options] [DIMG ...] REPO
```

##### `--dir PATH`
Определяет директорию, которая используется при поиске **dappfile**.

##### `--name NAME`
Переопределяет имя проекта dapp. По умолчанию, имя проекта — это последний элемент пути к git репозиторию из параметра конфигурации remote.origin.url или, при отсутствии git или параметра конфигурации remote.origin.url, это имя директории dapp.

При переопределении имени проекта в `dapp dimg build`, его следует переопределять и в последующем, - например при вызове команд `dapp dimg push`, `dapp dimg cleanup`, `dapp kube deploy`.

##### `--with-stages`
Также выкатить кэш стадий.

#### Опции доступа к registry
Для аутентификации в docker registry используются значения
      * значения пользовательских опций `--registry-username`, `--registry-password`;
      * переменная окружения CI_JOB_TOKEN;
      * полученный token из конфига docker-а.

##### `--registry-username USERNAME`
Определяет имя для доступа к registry.

##### `--registry-password PASSWORD`
Определяет пароль для доступа к registry.

#### Опции логирования и отладки

##### `--dev`
Включает режим разработчика.

##### `--dry-run`
Позволяет запустить выкат вхолостую и посмотреть процесс (например - результат применения опций тегирования).

##### `--ignore-config-sequential-processing-warnings`
Отключает вывод предупреждений при обработке dappfile.

##### `-q, --quiet`
Отключает логирование.

##### `--color MODE`
Отвечает за регулирование цвета при выводе в терминал.

Существует несколько режимов (**MODE**): **on**, **off**, **auto**.

По умолчанию используется **auto**, который окрашивает вывод, если вывод производится непосредственно в терминал.

##### `--time`
Добавляет время каждому событию лога.

##### `--lock-timeout TIMEOUT`
Определяет таймаут ожидания заблокированных ресурсов в секундах (по умолчанию - 86400 секунд, т.е. 24 часа).

#### Опции тегирования
Отвечают за тег(и), с которыми выкатываются собранные dimg-ы.

Могут быть использованы совместно и по несколько раз.

В случае отсутствия, используется тег **latest**.

##### `--tag TAG`, `--tag-slug TAG`
Добавляет произвольный тег **TAG**, который слагифицируется перед использованием (см. например команду [dapp dimg slug](dimg_slug.html)).

##### `--tag-plain TAG`
Добавляет тег **TAG**.

##### `--tag-branch`
Добавляет тег с именем ветки сборки.

##### `--tag-commit`
Добавляет тег с коммитом сборки.

##### `--tag-build-id`
Добавляет тег с идентификатором сборки (CI).

##### `--tag-ci`
Добавляет теги, взятые из переменных окружения CI систем.

#### Примеры

##### Выкатить все dimg-ы в репозиторий localhost:5000/test и тегом latest
```bash
$ dapp dimg push localhost:5000/test
```

##### Посмотреть, какие образы могут быть добавлены в репозиторий
```bash
$ dapp dimg push localhost:5000/test --tag yellow --tag-branch --dry-run
backend
  localhost:5000/test:backend-yellow
  localhost:5000/test:backend-master
frontend
  localhost:5000/test:frontend-yellow
  localhost:5000/test:frontend-0.2
```


### dapp dimg tag
Повторяет поведение и опции команды [dapp dimg push](#dapp-dimg-push), при этом, сохраняя dimg-ы локально. Если не указать *REPO*, будет использовано имя проекта.

```
dapp dimg tag [options] [DIMG...] [REPO]
```


### dapp dimg spush
Выкатить собранный dimg в docker registry, в формате **REPO**:**TAG**.

```
dapp dimg spush [options] [DIMG] REPO
```

Опции - такие же как у [**dapp dimg push**](#dapp-dimg-push).

`spush` - сокращение от `simple push`, и отличие `spush` от `dapp dimg push` в том, что при использовании `spush` пушится только один образ, и формат его имени в registry - простой (**REPO**:**TAG**). Т.о., если в dappfile описано несколько образов, то указание образа при вызове `dapp dimg spush` является необходимым.

#### Примеры

##### Выкатить собранный dimg **app** в репозиторий test и тегом latest
```bash
$ dapp dimg spush app localhost:5000/test
```

##### Выкатить собранный dimg с произвольными тегами
```bash
$ dapp dimg spush app localhost:5000/test --tag 1 --tag test
```

##### Посмотреть, какие образы могут быть добавлены в репозиторий
```bash
$ dapp dimg spush app localhost:5000/test --tag-commit --tag-branch --dry-run
localhost:5000/test:2c622c16c39d4938dcdf7f5c08f7ed4efa8384c4
localhost:5000/test:master
```
