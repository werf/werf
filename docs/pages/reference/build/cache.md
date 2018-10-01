---
title: Cache
sidebar: reference
permalink: reference/build/cache.html
---

Dapp использует в работе разноуровневое эффективное кэширование, правильное использование которого может существенно сократить время сборки приложения и уменьшить размер его конечного образа. Далее рассматривается принцип организации сборочного кэша приложения и способы уменьшения итогового размера образа приложения с использованием примонтированных директорий.

В dapp при сборке образов используются следующие кэши:

* Сборочный кэш приложения — образы в docker registry.
* Кэш из build-директории:
  * Кэш установленных cookbook'ов для chef сборщика.
  * Кэш для примонтированных сборочных директорий с использованием опции `from :build_dir`.

## Сборочный кэш приложения

Сборочный кэш приложения — это набор docker-образов, который определяется стадиями приложения: одной стадии соответствует один docker-образ.

Любая стадия в dapp имеет идентификатор содержимого стадии, называемый сигнатурой стадии. Сигнатура стадии зависит от правил её сборки и представляет собой контрольную сумму от этих правил. Если разные стадии имеют одинаковую сигнатуру — это означает, что содержимое этих образов одинаково. Соответственно стадии, имеющие одинаковую сигнатуру представлены одним docker образом. Любое изменение инструкций стадии приводит к полной пересборке стадии, со всеми её инструкциями.

Стандартный процесс сборки приложения через dapp представляет собой сборку набора docker-образов. В процессе сборки docker-образы тех стадий, которые собрались остаются "невидимыми" пользователю dapp, они имеют только временный идентификатор в docker. После успешной сборки все промежуточные образы именуются и попадают в кэш образов. Образы из кэша используются при повторных сборках, а также их можно интроспектить вручную через docker run.

## Создание сборочного кэша

Технически docker-образ попадает в сборочный кэш в тот момент, когда dapp тегирует этот образ специальным образом: используя `dimgstage-<имя-репозитория-приложения>` в качестве имени образа и сигнатуры стадии в качестве тега образа. Таким образом, стандартное поведение dapp предполагает, что при сборке формируется один docker-образ на все инструкции каждой стадии.

Dapp сохраняет собранные слои в кэш только после успешной сборки всех стадий. Если в процессе сборки некоторой стадии произошла ошибка, то все успешно собранные на этот момент стадии будут потеряны. При повторном запуске сборка начнется с того же образа, с которого была начата предыдущая сборка.

Такой механизм работы кэша необходим, чтобы обеспечить строгую корректность сохраненного кэша.

## Принудительное сохранение образов в кэш после сборки

Для разработчика конфигурации было бы удобнее, если бы все успешно собранные стадии сразу сохранялись в кэш docker образов. В таком случае, при возникновении ошибки, пересборка бы всегда начиналась с ошибочной стадии.

Для этой цели в dapp предусмотрена возможность принудительного сохранения кэша, включаемая либо опцией `--force-save-cache`, либо наличием переменной окружения `DAPP_FORCE_SAVE_CACHE=1`.

На примере dappfile:

```yaml
dimg: ~
from: ubuntu:16.04
shell:
  beforeInstall:
  - apt-get update
  - apt-get install -y curl --quiet
  install:
  - apt-get install -y non-existing
```

Запустим сборку:

 ```shell
$ dapp dimg build --force-save-cache
nameless: calculating stages signatures                                     [RUNNING]
nameless: calculating stages signatures                                          [OK] 0.16 sec
From ...                                                                         [OK] 1.09 sec
  signature: dimgstage-myapp2:41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11
Before install                                                             [BUILDING]
Get:1 http://archive.ubuntu.com/ubuntu xenial InRelease [247 kB]
Get:2 http://archive.ubuntu.com/ubuntu xenial-updates InRelease [109 kB]
...
done.
Before install                                                                   [OK] 22.85 sec
  signature: dimgstage-myapp2:8f8307adb4d2434822cdbb44950868b1a312d1a0e536ae54debff9640f371645
  commands:
    apt-get update
    apt-get install -y curl --quiet
Install group
  Install                                                                  [BUILDING]
Reading package lists...
Building dependency tree...
Reading state information...
E: Unable to locate package non-existing
  Install                                                                    [FAILED] 2.03 sec
    signature: dimgstage-myapp2:1c0aca95f86933173709388f4f75cdc50e210a861d3e85193f14556bf4a798f8
    commands:
      apt-get install -y non-existing
Running time 28.02 seconds
Stacktrace dumped to /tmp/dapp-stacktrace-f9333e01-c9b9-4f31-809a-12ada6f7c64d.out
ruby2go_image command `build` failed!
```

При повторном запуске стадия Before install более не будет пересобираться, т.к. была закэширована при первом запуске:

```shell
$ dapp dimg build --force-save-cache
nameless: calculating stages signatures                                     [RUNNING]
nameless: calculating stages signatures                                          [OK] 0.16 sec
From                                                                    [USING CACHE]
  signature: dimgstage-myapp2:41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11
  date: 2018-08-29 19:07:02 +0300
  size: 113.913 MB
Before install                                                          [USING CACHE]
  signature: dimgstage-myapp2:8f8307adb4d2434822cdbb44950868b1a312d1a0e536ae54debff9640f371645
  date: 2018-08-29 19:07:24 +0300
  difference: 57.087 MB
Install group
  Install                                                                  [BUILDING]
Reading package lists...
Building dependency tree...
Reading state information...
E: Unable to locate package non-existing
  Install                                                                    [FAILED] 2.03 sec
    signature: dimgstage-myapp2:1c0aca95f86933173709388f4f75cdc50e210a861d3e85193f14556bf4a798f8
    commands:
      apt-get install -y non-existing
Running time 3.85 seconds
Stacktrace dumped to /tmp/dapp-stacktrace-9adbb391-79e4-421f-83b1-dcdad372051c.out
ruby2go_image command `build` failed!
```

## Почему dapp не сохраняет кэш ошибочных сборок по умолчанию?

Режим работы DAPP_FORCE_SAVE_CACHE может привести к созданию некорректного кэша. Легко исправить такую ситуацию можно будет лишь ручным удалением этого кэша.

В каком случае может получится ситуация, когда был сохранен некорректный кэш проще пояснить на примере.

Инициализируем приложение стандартным dappfile:

```yaml
dimg: ~
from: ubuntu:16.04
git:
- add: /
  to: /app
```

Собираем:

```shell
$ dapp dimg build --force-save-cache
nameless: calculating stages signatures                                     [RUNNING]
  Repository `own`: latest commit `3d70fcec74abf7b8197230830bb6d7ccf5826952` to `/app`
nameless: calculating stages signatures                                          [OK] 0.24 sec
From ...                                                                         [OK] 1.56 sec
  signature: dimgstage-myapp:41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11
Git artifacts: create archive ...                                                [OK] 1.37 sec
  signature: dimgstage-myapp:d1aa6029faae81733618867c217c9e0e9d70e56ab1fc2554e790d9b14f16b96c
Setup group
  Git artifacts: apply patches (after setup) ...                                 [OK] 1.39 sec
    signature: dimgstage-myapp:336636cedd354d7903d71d242b4a8c40dd0bf81728b0e189deee26cd1d59ec6b
Running time 13.9 seconds
```

Сборка прошла успешно, сборочный кэш наполнился корректными стадиями. Далее добавим сборочную инструкцию, которая использует файл из git. Но намеренно совершим ошибку в этой инструкции -- попытаемся скопировать файл `/app/hello`, которого нет в git. Например, пользователь забыл его добавить.

```yaml
dimg: ~
from: ubuntu:16.04
git:
- add: /
  to: /app
shell:
  install:
  - cp /app/hello /hello
```

Сборка с этим dappfile приводит к ошибке:

```shell
$ dapp dimg build --force-save-cache
nameless: calculating stages signatures                                     [RUNNING]
  Repository `own`: latest commit `895f42cd25d025018c00ad5ac6fe88764cfca980` to `/app`
nameless: calculating stages signatures                                          [OK] 0.33 sec
From                                                                    [USING CACHE]
  signature: dimgstage-myapp:41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11
  date: 2018-08-29 18:22:07 +0300
  size: 113.913 MB
Git artifacts: create archive                                           [USING CACHE]
  signature: dimgstage-myapp:d1aa6029faae81733618867c217c9e0e9d70e56ab1fc2554e790d9b14f16b96c
  date: 2018-08-29 18:22:08 +0300
  difference: 0.0 MB
Install group
  Git artifacts: apply patches (before install) ...                              [OK] 1.47 sec
    signature: dimgstage-myapp:3a4b24a524f72e259bc8e5d6335ca7aaa4504d08da9d63d31c42df92331fd24d
  Install                                                                  [BUILDING]
cp: cannot stat '/app/hello': No such file or directory
    Launched command: `cp /app/hello /hello`
  Install                                                                    [FAILED] 1.43 sec
    signature: dimgstage-myapp:003e8da0e54baddc3ebc5e499fdd29d1af4dbd88626a9606d9dc32df725b433e
    commands:
      cp /app/hello /hello
Running time 5.01 seconds
Stacktrace dumped to /tmp/dapp-stacktrace-38fa7ded-c542-4fef-9f1f-5cf6cae662f9.out
>>> START STREAM
cp: cannot stat '/app/hello': No such file or directory
>>> END STREAM
```

При сборке dapp заметил, что добавились инструкции сборки стадии install и начал пересобирать эту стадию. Предварительно перед сборкой стадии install была собрана стадия `Git artifacts: apply patches (before install)`, в которой накладывается патч с изменениями из git. Это делается чтобы на стадии install были доступны актуальные файлы из git.

Далее попытаемся исправить ошибку и добавим файл hello в git. Запускаем пересборку и видим ту же самую ошибку: нет файла hello.

```shell
$ dapp dimg build --force-save-cache
nameless: calculating stages signatures                                     [RUNNING]
  Repository `own`: latest commit `a6d7b54cd8055df635475c7e9972237a0974142b` to `/app`
nameless: calculating stages signatures                                          [OK] 0.4 sec
From                                                                    [USING CACHE]
  signature: dimgstage-myapp:41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11
  date: 2018-08-29 18:22:07 +0300
  size: 113.913 MB
Git artifacts: create archive                                           [USING CACHE]
  signature: dimgstage-myapp:d1aa6029faae81733618867c217c9e0e9d70e56ab1fc2554e790d9b14f16b96c
  date: 2018-08-29 18:22:08 +0300
  difference: 0.0 MB
Install group
  Git artifacts: apply patches (before install)                         [USING CACHE]
    signature: dimgstage-myapp:3a4b24a524f72e259bc8e5d6335ca7aaa4504d08da9d63d31c42df92331fd24d
    date: 2018-08-29 18:35:51 +0300
    difference: 0.0 MB
  Install                                                                  [BUILDING]
cp: cannot stat '/app/hello': No such file or directory
    Launched command: `cp /app/hello /hello`
  Install                                                                    [FAILED] 1.25 sec
    signature: dimgstage-myapp:003e8da0e54baddc3ebc5e499fdd29d1af4dbd88626a9606d9dc32df725b433e
    commands:
      cp /app/hello /hello
Running time 2.07 seconds
Stacktrace dumped to /tmp/dapp-stacktrace-10b05694-bdc5-463c-8abb-3748b20d5acb.out
>>> START STREAM
cp: cannot stat '/app/hello': No such file or directory
>>> END STREAM
```

Этот файл должен был попасть в образ на стадии `Git artifacts: apply patches (before install)`, однако данная стадия была закэширована на тот момент, когда файла еще не было в git. Для исправления придется вручную удалить эту стадию.

В этом особенность работы кэша dapp. Сигнатура данной стадии не может зависеть от произвольных файлов в git, иначе теряется весь смысл кэширования стадий install, before setup, setup. А если не меняется сигнатура, то не меняется и кэш.

Данной проблемы можно избежать, если сохранять в кэш только стадии корректных сборок. Поэтому опцию DAPP_FORCE_SAVE_CACHE рекомендуется использовать с осторожностью.
