---
title: Базовые команды
sidebar: doc_sidebar
permalink: base_commands.html
folder: command
---

### dapp build
Собрать dimg-ы, удовлетворяющие хотя бы одному из **DIMG**-ов (по умолчанию *).

```
dapp build [options] [DIMG ...]
```

#### Опции среды сборки

##### --dir PATH
Определяет директорию, которая используется при поиске одного или нескольких **Dappfile**.

По умолчанию поиск ведётся в текущей директории пользователя.

##### --build-dir PATH
Переопределяет директорию хранения кэша, который может использоваться между сборками.

##### --tmp-dir-prefix PREFIX
Переопределяет префикс временной директории, файлы которой используются только во время сборки.

#### Опции логирования

##### --dry-run
Позволяет запустить сборщик вхолостую и посмотреть процесс сборки.

##### --verbose
Подробный вывод.

##### --color MODE
Отвечает за регулирование цвета при выводе в терминал.

Существует несколько режимов (**MODE**): **on**, **off**, **auto**.

По умолчанию используется **auto**, который окрашивает вывод, если вывод производится непосредственно в терминал.

##### --time
Добавляет время каждому событию лога.

#### Опции интроспекции
Позволяют поработать с образом на определённом этапе сборки.

##### --introspect-stage STAGE
После успешного прохождения стадии **STAGE**.

##### --introspect-before-error
Перед выполением команд несобравшейся стадии.

##### --introspect-error
После завершения команд стадии с ошибкой.

#### Примеры

##### Сборка в текущей директории
```bash
$ dapp build
```

##### Сборка dimg-ей из соседней директории
```bash
$ dapp build --dir ../project
```

##### Запуск вхолостую с подробным выводом процесса сборки
```bash
$ dapp build --dry-run --verbose
```

##### Выполнить сборку, а в случае ошибки, предоставить образ для тестирования
```bash
$ dapp build --introspect-error
```

### dapp push
Выкатить собранные dimg-ы в репозиторий, в следующем формате **REPO**:**ИМЯ DIMG**-**TAG**. В случае, если **ИМЯ DIMG** отсутствует, формат следующий **REPO**:**TAG**.

```
dapp push [options] [DIMG ...] REPO
```

#### --with-stages
Также выкатить кэш стадий.

#### --force
Позволяет перезаписывать существующие образы.

#### Опции тегирования
Отвечают за тег(и), с которыми выкатывается собранные dimg-ы.

Могут быть использованы совместно и по несколько раз.

В случае отсутствия, используется тег **latest**.

##### --tag TAG
Добавляет произвольный тег **TAG**.

##### --tag-branch
Добавляет тег с именем ветки сборки.

##### --tag-commit
Добавляет тег с коммитом сборки.

##### --tag-build-id
Добавляет тег с идентификатором сборки (CI).

##### --tag-ci
Добавляет теги, взятые из переменных окружения CI систем.

#### Примеры

##### Выкатить все dimg-ы в репозиторий localhost:5000/test и тегом latest
```bash
$ dapp push localhost:5000/test
```

##### Посмотреть, какие образы могут быть добавлены в репозиторий
```bash
$ dapp push localhost:5000/test --tag yellow --tag-branch --dry-run
backend
  localhost:5000/test:backend-yellow
  localhost:5000/test:backend-master
frontend
  localhost:5000/test:frontend-yellow
  localhost:5000/test:frontend-0.2
```

### dapp spush
Выкатить собранный dimg в репозиторий, в следующем формате **REPO**:**TAG**.

```
dapp spush [options] [DIMG] REPO
```

Опции такие же как у **dapp push**.

#### Примеры

##### Выкатить собранный dimg **app** в репозиторий test, именем myapp и тегом latest
```bash
$ dapp spush app localhost:5000/test
```

##### Выкатить собранный dimg с произвольными тегами
```bash
$ dapp spush app localhost:5000/test --tag 1 --tag test
```

##### Посмотреть, какие образы могут быть добавлены в репозиторий
```bash
$ dapp spush app localhost:5000/test --tag-commit --tag-branch --dry-run
localhost:5000/test:2c622c16c39d4938dcdf7f5c08f7ed4efa8384c4
localhost:5000/test:master
```

### dapp stages push
Выкатить [кэш собранных приложений](definitions.html#кэш-приложения) [проекта](definitions.html#проект) в репозиторий.

```
dapp stages push [options] [DIMG ...] REPO
```

#### Примеры

##### Выкатить кэш собранных dimg-ей проекта в репозиторий localhost:5000/test
```bash
$ dapp stages push localhost:5000/test
```

##### Посмотреть, какие образы могут быть добавлены в репозиторий
```bash
$ dapp stages push localhost:5000/test --dry-run
backend
  localhost:5000/test:dimgstage-be032ed31bd96506d0ed550fa914017452b553c7f1ecbb136216b2dd2d3d1623
  localhost:5000/test:dimgstage-2183f7db73687e727d9841594e30d8cb200312290a0a967ef214fe3771224ee2
  localhost:5000/test:dimgstage-f7d4c5c420f29b7b419f070ca45f899d2c65227bde1944f7d340d6e37087c68d
  localhost:5000/test:dimgstage-256f03ccf980b805d0e5944ab83a28c2374fbb69ef62b8d2a52f32adef91692f
  localhost:5000/test:dimgstage-31ed444b92690451e7fa889a137ffc4c3d4a128cb5a7e9578414abf107af13ee
  localhost:5000/test:dimgstage-02b636d9316012880e40da44ee5da3f1067cedd66caa3bf89572716cd1f894da
frontend
  localhost:5000/test:dimgstage-192c0e9d588a51747ce757e61be13acb4802dc2cdefbeec53ca254d014560d46
  localhost:5000/test:dimgstage-427b999000024f9268a46b889d66dae999efbfe04037fb6fc0b1cd7ebb4600b0
  localhost:5000/test:dimgstage-07fe13aec1e9ce0fe2d2890af4e4f81aaa984c89a2b91fbd0e164468a1394d46
  localhost:5000/test:dimgstage-ba95ec8a00638ddac413a13e303715dd2c93b80295c832af440c04a46f3e8555
  localhost:5000/test:dimgstage-02b636d9316012880e40da44ee5da3f1067cedd66caa3bf89572716cd1f894da
```

### dapp stages pull
Импортировать необходимый [кэш приложений](definitions.html#кэш-приложения) [проекта](definitions.html#проект), если он присутствует в репозитории **REPO**.

Если не указана опция **--all**, импорт будет выполнен до первого найденного кэша стейджа для каждого dimg-a.

```
dapp stages pull [options] [DIMG ...] REPO
```

#### --all
Попробовать импортировать весь необходимый кэш.

#### Примеры

##### Импортировать кэш dimg-ей проекта из репозитория localhost:5000/test
```bash
$ dapp stages pull localhost:5000/test
```

##### Посмотреть, поиск каких образов в репозитории localhost:5000/test может быть выполен
```bash
$ dapp stages pull localhost:5000/test --all --dry-run
backend
  localhost:5000/test:dimgstage-be032ed31bd96506d0ed550fa914017452b553c7f1ecbb136216b2dd2d3d1623
  localhost:5000/test:dimgstage-2183f7db73687e727d9841594e30d8cb200312290a0a967ef214fe3771224ee2
  localhost:5000/test:dimgstage-f7d4c5c420f29b7b419f070ca45f899d2c65227bde1944f7d340d6e37087c68d
  localhost:5000/test:dimgstage-256f03ccf980b805d0e5944ab83a28c2374fbb69ef62b8d2a52f32adef91692f
  localhost:5000/test:dimgstage-31ed444b92690451e7fa889a137ffc4c3d4a128cb5a7e9578414abf107af13ee
  localhost:5000/test:dimgstage-02b636d9316012880e40da44ee5da3f1067cedd66caa3bf89572716cd1f894da
frontend
  localhost:5000/test:dimgstage-192c0e9d588a51747ce757e61be13acb4802dc2cdefbeec53ca254d014560d46
  localhost:5000/test:dimgstage-427b999000024f9268a46b889d66dae999efbfe04037fb6fc0b1cd7ebb4600b0
  localhost:5000/test:dimgstage-07fe13aec1e9ce0fe2d2890af4e4f81aaa984c89a2b91fbd0e164468a1394d46
  localhost:5000/test:dimgstage-ba95ec8a00638ddac413a13e303715dd2c93b80295c832af440c04a46f3e8555
  localhost:5000/test:dimgstage-02b636d9316012880e40da44ee5da3f1067cedd66caa3bf89572716cd1f894da
```

### dapp tag
Протегировать собранный dimg тегом **TAG**.

```
dapp tag [options] [DIMG] TAG
```

#### Примеры

##### Протегировать собранный dimg тегом test:111.
```bash
$ dapp tag test:111 --verbose
```