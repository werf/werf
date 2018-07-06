---
title: Базовые команды
sidebar: doc_sidebar
permalink: base_commands.html
folder: command
---

### dapp dimg build
Собрать dimg-ы, удовлетворяющие хотя бы одному из **DIMG**-ов (по умолчанию *).

```
dapp dimg build [options] [DIMG ...]
```

#### Опции среды сборки

##### `--dir PATH`
Определяет директорию, которая используется при поиске **dappfile**.

По умолчанию поиск ведётся в текущей директории пользователя.

##### `--build-dir PATH`
Переопределяет директорию хранения кэша, который может использоваться между сборками. Значение по умолчанию - `~/.dapp/builds/<project-name>`. Изменение опции build-dir не ведет к изменениям в системе именования helm-релизов и docker-образов, поэтому это безопасное изменение.

##### `--tmp-dir-prefix PREFIX`
Переопределяет префикс временной директории, файлы которой используются только во время сборки.

##### `--name NAME`
Переопределяет [имя проекта dapp](definitions.html#имя-dapp). По умолчанию, имя проекта — это последний элемент пути к git репозиторию из параметра конфигурации remote.origin.url или, при отсутствии git или параметра конфигурации remote.origin.url, это имя [директории dapp](definitions.html#директория-dapp).

**Важно:** Изменение имени ранее собранного проекта приведет к тому, что кэш проекта будет пересобран заново. Также, будет изменено имя релиза в helm, поэтому указывать `--name` надо только при возникающей необходимости, когда есть например конфликтующее имя проекта в другой группе. `--build-dir` будет изменен автоматически, при условии, что опция `--build-dir` не указана (см описании опции `--build-dir` [выше](#build-dir-path).

При переопределении имени проекта в `dapp dimg build`, его следует переопределять и впоследствии, например при вызове команд `dapp dimg push`, `dapp dimg cleanup`, `dapp kube deploy`.

Использование опции `--name` может быть нужно, если например в рамках одного GitLab в разных группах конфликтуют имена проектов. В этом случае надо обязательно использовать опцию `--name`, в которой указать составленное уникальное имя проекта (например `<имя-группы>_<имя-проекта>`).

##### `--ssh-key SSH_KEY`
Позволяет указать ssh ключ для доступа к удаленному репозиторию, вместо используемого по умолчанию ssh-agent.

В случае, если явно не указана опция `--ssh-key` и не запущен системный ssh-agent (проверяется по переменной окружения SSH_AUTH_SOCK), dapp автоматически запускает временный ssh-agent и добавляет в него ключи по умолчанию - ~/.ssh/id_rsa, ~/.ssh/id_dsa. Данный агент будет использоваться для операций с git, chef berks, в сборочных контейнерах.

#### Опции логирования и отладки

##### `--dev`
Включает [режим разработчика](debug_for_advanced_build.html#режим-разработчика).

##### `--dry-run`
Позволяет запустить сборщик вхолостую и посмотреть процесс сборки.

##### `--ignore-config-sequential-processing-warnings`
Отключает вывод предупреждений при обработке dappfile.

##### `-q, --quiet`
Отключает логирование.

##### `--color MODE`
Отвечает за регулирование цвета при выводе в терминал.

Существует несколько режимов (**MODE**): **on**, **off**, **auto**.

По умолчанию используется **auto**, который окрашивает вывод, если вывод производится непосредственно в терминал.

##### `--build-context-directory DIR_PATH`
Включает сохранение контекста при неудачной сборке, а также определяет директорию сохранения.

##### `--use-system-tar`
При указании опции `--use-system-tar`, для создания tar-архивов будет использоваться вызов системной команды `tar`, вместо используемой по умолчанию ruby библиотеки `tar_writer`. Библиотека `tar_writer` накладывает ряд ограничений, и иногда ее использование может приводить к ошибкам (например такой - `File "XXX" has a too long name (should be 100 or less)`).

##### `--time`
Добавляет время каждому событию лога.

##### `--lock-timeout TIMEOUT`
Определяет таймаут ожидания заблокированных ресурсов в секундах (по умолчанию - 86400 секунд, т.е. 24 часа).

Если в параллельно запущенных сборках разных проектов в одинаковых стадиях присутствует одинаковый набор инструкций (например одинаковая стадия beforeInstall), то сигнатура образа стадии будет одинаковая. В этой ситуации два запущенных параллельно процеса dapp не будут собирать один и тот же образ (в этом нет смысла - результат обоих сборок будет одинаков). Сборка будет осуществляться первым запущенным процессом, а второй процесс будет ожидать окончания сборки до достижения таймаута.

#### Опции интроспекции
Позволяют поработать с образом на определённом этапе сборки. [Узнайте больше](debug_for_advanced_build.html), о работе с опциями интроспекции.

##### `--introspect-stage STAGE`, `--introspect-artifact-stage STAGE`
Интроспекция собранной стадии STAGE, т.е. после успешного прохождения стадии **STAGE** (`--introspect-artifact-stage` - используется для образа артефакта).

##### `--introspect-before STAGE`, `--introspect-artifact-before STAGE`
Интроспекция до выполнения инструкций стадии STAGE (`--introspect-artifact-before` - используется для образа артефакта).

##### `--introspect-before-error`
Интроспекция стадии предшествующей той, на которой произошла ошибка.

##### `--introspect-error`
Интроспекция стадии на которой произошла ошибка, сразу после исполнения завершившейся ошибкой команды.

#### Примеры

##### Сборка в текущей директории
```bash
$ dapp dimg build
```

##### Сборка dimg-ей из соседней директории
```bash
$ dapp dimg build --dir ../project
```

##### Запуск вхолостую с выводом процесса сборки
```bash
$ dapp dimg build --dry-run
```

##### Выполнить сборку, а в случае ошибки, предоставить образ для тестирования
```bash
$ dapp dimg build --introspect-error
```

##### Выполнить сборку, а в случае ошибки, сохранить контекст сборки в директории 'context'
```bash
$ dapp dimg build --build-context-directory context
```

### dapp dimg push
Выкатить собранные dimg-ы в docker registry, в формате **REPO**/**ИМЯ DIMG**:**TAG**. В случае, если **ИМЯ DIMG** отсутствует, формат следующий **REPO**:**TAG**.

```
dapp dimg push [options] [DIMG ...] REPO
```

##### `--dir PATH`
Определяет директорию, которая используется при поиске **dappfile**.

##### `--name NAME`
Переопределяет [имя проекта dapp](definitions.html#имя-dapp). По умолчанию, имя проекта — это последний элемент пути к git репозиторию из параметра конфигурации remote.origin.url или, при отсутствии git или параметра конфигурации remote.origin.url, это имя [директории dapp](definitions.html#директория-dapp).

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
Включает [режим разработчика](debug_for_advanced_build.html#режим-разработчика).

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
Добавляет произвольный тег **TAG**, который слагифицируется перед использованием (см. например команду [dapp slug](debugging_commands.html#dapp-slug)).

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

### dapp dimg stages push
Выкатить [кэш собранных приложений](definitions.html#кэш-приложения) [проекта](definitions.html#проект) в репозиторий.

```
dapp dimg stages push [options] [DIMG ...] REPO
```

#### Примеры

##### Выкатить кэш собранных dimg-ей проекта в репозиторий localhost:5000/test
```bash
$ dapp dimg stages push localhost:5000/test
```

##### Посмотреть, какие образы могут быть добавлены в репозиторий
```bash
$ dapp dimg stages push localhost:5000/test --dry-run
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

### dapp dimg stages pull
Импортировать необходимый [кэш приложений](definitions.html#кэш-приложения) [проекта](definitions.html#проект), если он присутствует в репозитории **REPO**.

Если не указана опция **`--all`**, импорт будет выполнен до первого найденного кэша стейджа для каждого dimg-a.

```
dapp dimg stages pull [options] [DIMG ...] REPO
```

#### `--all`
Попробовать импортировать весь необходимый кэш.

#### Примеры

##### Импортировать кэш dimg-ей проекта из репозитория localhost:5000/test
```bash
$ dapp dimg stages pull localhost:5000/test
```

##### Посмотреть, поиск каких образов в репозитории localhost:5000/test может быть выполнен
```bash
$ dapp dimg stages pull localhost:5000/test --all --dry-run
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

### dapp dimg tag
Повторяет поведение и опции команды [dapp dimg push](#dapp-dimg-push), при этом, сохраняя dimg-ы локально. Если не указать *REPO*, будет использовано [имя проекта](definitions.html#имя-dapp).

```
dapp dimg tag [options] [DIMG...] [REPO]
```
