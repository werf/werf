---
title: Как устроено кэширование
sidebar: not_used
permalink: not_used/cache_for_advanced_build.html
author: Alexey Igrychev <alexey.igrychev@flant.com>, Timofey Kirillov <timofey.kirillov@flant.com>
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

## Кеширование отдельных инструкций стадии

На этапе разработки dappfile, при выполнении ресурсоемких инструкций и перезапуске сборки ошибочной стадии, это может требовать значительных временных затрат. Для удобства разработки и отладки в dapp была введена **альтернативная схема кэширования**, которая доступна **только в YAML-конфигурации** и включается директивой `asLayers`.

Директива `asLayers` может быть указана в конфигурации (dappfile.yaml) для конкретного приложения или его артефакта. При включении альтернативной схемы кэширования, dapp формирует один docker-образ на каждую команду для shell или каждый task для Ansible, что приводит к тому, что инструкции при сборке кэшируются по отдельности. Пересборка образа в этом случае осуществляется только при изменении порядка инструкций.

**Важно**. Альтернативная схема кэширования порождает избыточное количество docker-образов и не рассчитана на инкрементальную сборку (увеличивается время ожидания и размер сборочного кэша) - она предназначена для разработки, по-этому ее не нужно использовать в продуктивных средах.

Для облегчения локальной разработки, в dapp предусмотрен специальный **режим разработчика**, отличие которого, с точки зрения кэширования, - в наличии отдельного кэша.

Для примера, docker-образы из сборочного кэша приложения dapp-example-chef-for-advanced-build-2 будут выглядеть так:

```shell
REPOSITORY                                         TAG                                                                IMAGE ID            CREATED             SIZE
dimgstage-dapp-example-chef-for-advanced-build-2   5ca21f76a99a4a9ac1995d9a8a9c20779795c1c6eb2f8ed64bd773cb9f976589   363f6a74a9d3        25 hours ago        696.4 MB
dimgstage-dapp-example-chef-for-advanced-build-2   dd1eeec8047f1734873a5774b1a6c5b420d20fd285565d25251ded257abe5c29   d7634201dc13        25 hours ago        696.4 MB
dimgstage-dapp-example-chef-for-advanced-build-2   e91bc27414fe2dd8a5d5caf3c76e734a675111ccd7e3857e9e62389bb6d7c566   3a38f7584edf        25 hours ago        696.4 MB
dimgstage-dapp-example-chef-for-advanced-build-2   ea1be67f7231595de75d6c7b523cd3f800be6d56aa3892b65d4b55f9750566bb   fe17f466c4bc        25 hours ago        696.4 MB
dimgstage-dapp-example-chef-for-advanced-build-2   d85001d96adf7b7f8db44deb7d9f299004a64dde382c5ffa4aff03909760e33f   003334f54a59        25 hours ago        696.4 MB
dimgstage-dapp-example-chef-for-advanced-build-2   54adf11783b64cf173d289110b686e1c59cf260bc8f0229d23d556d235d78dc6   3cad59340388        25 hours ago        696.4 MB
dimgstage-dapp-example-chef-for-advanced-build-2   712f10136d4f3eebe0323d60c32f55983b0e3cc7a0c26e2f24f2f8ed1907ce34   89d1b78e46f8        25 hours ago        696.4 MB
dimgstage-dapp-example-chef-for-advanced-build-2   0de3ccb420da3c0933afeb6da5e3cb4e8bb4fef9cd1cb972d9e0e5ffa7fdb250   17745eedc11d        25 hours ago        696.4 MB
dimgstage-dapp-example-chef-for-advanced-build-2   68cfaf347f8a36d82e760f0beabe0010cff0a36414a91f961572858de862f146   058cad43eaee        25 hours ago        696.4 MB
dimgstage-dapp-example-chef-for-advanced-build-2   0f9f8bf59e0c41c356f421c7172fbb986bc7f0ab44ffda994368ffefcbbbd694   38436caa71c0        25 hours ago        696.4 MB
dimgstage-dapp-example-chef-for-advanced-build-2   572c4398a323ee3c4381537652969b90bd15a9f01c3fc18dc49aa7686073e580   1bc8cc41f341        25 hours ago        696.4 MB
dimgstage-dapp-example-chef-for-advanced-build-2   334e64203908c61c3f9b5a52f277730f4e322a832fe3808a42b90c4fcc2339d9   01c35cfa9f1c        25 hours ago        696.4 MB
dimgstage-dapp-example-chef-for-advanced-build-2   d5f44f46d94b344fd2eacf8ec6e92aee3e106dcaeff14e81bade0553cc235f31   2ebbf65b0968        25 hours ago        696.4 MB
dimgstage-dapp-example-chef-for-advanced-build-2   ba1e802a942cbc09b361a5dc9572911e98644e63d1aa501fe6a464e3a546a8c3   33a7862bf459        25 hours ago        696.4 MB
dimgstage-dapp-example-chef-for-advanced-build-2   79d5a4ccbd4fbc1db248037b5b333145574030f5c4d666d4056f562abff456fc   84fd57136bad        25 hours ago        696.4 MB
dimgstage-dapp-example-chef-for-advanced-build-2   b4eabf30cbfb46e6f4448eef6bd8460de40708213cfce81e9c2a7712c8b93694   90926f90755d        25 hours ago        681.3 MB
dimgstage-dapp-example-chef-for-advanced-build-2   dc86cf4c0af49a1c6250e83b7eab1b7632a851caedad5abe5cc98ff004127aaf   c14050f50d63        25 hours ago        681.3 MB
dimgstage-dapp-example-chef-for-advanced-build-2   25d9ad1d8304c4e1d9b23bda1ee42bcc239cb8a9ca11c47ba1bc5a7b9a1a6e63   2dfae5f84164        25 hours ago        681.3 MB
dimgstage-dapp-example-chef-for-advanced-build-2   ff48395b0d95ea1047e0f9ce96f48deefe98aa66aebe1f12d9376cce336b15f0   93256ff2c00e        25 hours ago        681.3 MB
dimgstage-dapp-example-chef-for-advanced-build-2   7454175f0f7915026bf78852bab47d49c1fdbe3586ec7c8528252059969b8d34   e0baa72ddbac        25 hours ago        681.3 MB
```

Docker-образы разных dimg, собираемых через один Dappfile, будут иметь одинаковые имена образов, различаться будут только теги. В имени образа используется:

* либо имя репозитория, в котором находится Dappfile (при этом Dappfile не обязательно должен находиться в корне этого репозитория);
* либо имя директории, в которой находится Dappfile.

Соответственно в примере выше `dapp-example-chef-for-advanced-build-2` — это имя репозитория тестового проекта [https://github.com/flant/dapp-example-chef-for-advanced-build-2](https://github.com/flant/dapp-example-chef-for-advanced-build-2).

## Оптимизация конечного размера образа приложения

Зачастую, в процессе сборки образа приложения в файловой системе возникают вспомогательные файлы, которые необходимо исключить из итогового образа. Например:

* Большинство пакетных менеджеров дистрибутивов создают общесистемный кэш пакетов и других служебных файлов, например:
  * При обновлении списка пакетов дистрибутива apt сохраняет его в директории `/var/lib/apt/lists`;
  * При установке пакетов apt оставляет скачанные пакеты в директории `/var/cache/apt/`;
  * При установке пакетов yum может оставлять скачанные пакеты в директории `/var/cache/yum/.../packages`.
* Такая же ситуация может быть с пакетными менеджерами языков программирования (npm для nodejs, glide для go, pip для python и прочие).
* Сборка c, c++ и подобных приложений оставляет после себя объектные файлы и другие файлы используемых систем сборок.

Полученные таким образом файлы:

* как правило не нужны в конечных образах, т.к. не используются;
* могут значительно увеличивать размер образа;
* могут понадобиться как кэш при последующих сборках новых версий образов.

Уменьшения размера конечного образа приложения можно добиться с помощью монтирования в сборочный образ внешних папок. Для монтирования внешних папок в dappfile предусмотрена [директива mount]({{ site.baseurl }}/reference/dappfile/mount_directive.html), которая позволяет монтировать внутрь сборочного контейнера не только произвольные директории, но и служебные директории двух типов:
- `tmp_dir` - временная директория, которая создается для каждого приложения **на время сборки** его образа;
- `build_dir` - временная директория, которая используется всеми приложениями описанными **в рамках одного проекта** для хранения кэша и производных данных сборки образов dapp.

### Использование временной директории на период сборки

Использование при монтировании директории `tmp_dir`, позволяет размещать определенные директории на период сборки во временном хранилище - временной директории, которая не будет включена в конечный образ. Например, есть Dappfile:

```ruby
dimg do
  docker.from 'ubuntu:16.04'

  shell.before_install do
    run 'apt-get update'
    run 'apt-get install -y nginx curl'
  end
end
```

В итоговом образе в `/var/lib/apt/lists` останутся списки пакетов, занимающие около 40Мб.

В следующем Dappfile при каждом новом запуске сборки в `/var/lib/apt/lists` будет чистая директория:

```ruby
dimg do
  docker.from 'ubuntu:16.04'

  mount '/var/lib/apt/lists' do
    from :tmp_dir
  end

  shell.before_install do
    run 'apt-get update'
    run 'apt-get install -y nginx curl'
  end
end
```

Т.к. указание в Dappfile директивы `from :tmp_dir` означает использование временной директории выделяемой на каждую новую сборку образов, после обновления списка пакетов в сборочном контейнере во время сборки произойдет наполнение этой директории, однако в итоговый образ эти файлы не попадут. В итоговом образе будет пустая директория `/var/lib/apt/lists`.

### Кэширование сборочных файлов между сборками

Использование при монтировании директории `build_dir`, позволяет включить кэширование содержимого примонтированной директории между сборками. В отличие от директории `tmp_dir` при каждом новом запуске сборки указанная директория будет сохраняться, при условии что служебная build-директория (`~/.dapp/builds/<dapp name>`) между сборками не была удалена. Т.е. в качестве хранилища будет выступать служебная build-директория.

Дополним Dappfile из примера выше таким образом, чтобы при пересборке стадии `before_install`, apt не скачивал заново устанавливаемые пакеты. Для этого примонтируем директорию `build_dir` по пути `/var/cache/apt`. Также, отключим стандартную автоочистку скачанных пакетов в `/etc/apt/apt.conf.d/docker-clean`.

```ruby
dimg do
  docker.from 'ubuntu:16.04'

  mount '/var/lib/apt/lists' do
    from :tmp_dir
  end

  mount '/var/cache/apt' do
    from :build_dir
  end

  shell.before_install do
    run 'sed -i -e "s/DPkg::Post-Invoke.*//" /etc/apt/apt.conf.d/docker-clean'
    run 'sed -i -e "s/APT::Update::Post-Invoke.*//" /etc/apt/apt.conf.d/docker-clean'

    run 'apt update'
    run 'apt install -y nginx curl'
  end
end
```

В следующем примере происходит установка rails в образ с использованием rvm. Для rvm включен кэш gem'ов, поэтому стоит примонтировать директорию `/usr/local/rvm/gems/cache`, используя `from :build_dir`.

```ruby
dimg do
  docker.from 'ubuntu:16.04'

  mount '/var/lib/apt/lists' do
    from :tmp_dir
  end

  mount '/var/cache/apt/archives' do
    from :build_dir
  end

  mount '/usr/local/rvm/gems/cache' do
    from :build_dir
  end

  shell.before_install do
    run 'apt-get update'
    run 'apt-get install -y nginx curl'
    run 'curl -sSL https://get.rvm.io | bash'
    run '. /etc/profile.d/rvm.sh && rvm install 2.3.1 && rvm use 2.3.1'
    run '. /etc/profile.d/rvm.sh && rvm gemset globalcache enable'
  end

  shell.install do
    run '. /etc/profile.d/rvm.sh && gem install rails'
  end
end
```

Внутри итогового образе видим, что все примонтированные директории пусты, а содержимое директорий использующих `build_dir` в качестве хранилища остается вне образа, в директории  `~/.dapp/builds/<dapp name>`:

```shell
root@a96aa4d220c3:/# ls /var/lib/apt/lists /var/cache/apt/archives /usr/local/rvm/gems/cache
/usr/local/rvm/gems/cache:

/var/cache/apt/archives:

/var/lib/apt/lists:
```

```shell
$ tree ~/.dapp/builds/<dapp name>/mount
/home/test/.dapp/builds/<dapp name>/mount
├── usr
│   └── local
│       └── rvm
│           └── gems
│               └── cache
│                   ├── actioncable-5.0.2.gem
│                   ├── actionmailer-5.0.2.gem
...
│                   ├── websocket-driver-0.6.5.gem
│                   └── websocket-extensions-0.1.2.gem
└── var
    └── cache
        └── apt
            └── archives
                ├── autoconf_2.69-9_all.deb
                ├── automake_1%3a1.15-4ubuntu1_all.deb
...
                ├── partial
                └── zlib1g-dev_1%3a1.2.8.dfsg-2ubuntu4_amd64.deb
```


### Исключение файлов полученных из базового образа

Не всегда есть возможность контролировать создание базового образа и в некоторых образах оказываются зашиты лишние вспомогательные файлы. Использование директории в качестве mount-point автоматически очищает эту директорию, даже если в ней что-то было до этого. Т.о, в Dappfile приведенном в примере выше, `/var/lib/apt/lists` будет очищен, даже если в образе, указанном в `docker.from`, эта директория не пуста. Принудительную очистку dapp делает на стадии `from`.

## Перенос кэша dapp на другую машину

Dapp предоставляет команды для переноса сборочного кэша и кэша из build-директории с машины на машину. Процесс переноса представляет собой следующие шаги:

### Экспорт кэша

Экспорт кэша — это создание набора архивов с помощью команды `dapp dimg build-context export`. По умолчанию экспорт создает в текущей директории набор архивов: build.tar (build-директория) и images.tar (сборочный кэш из docker).

### Импорт кэша

После переноса build.tar и images.tar на другую машину в директорию проекта, загрузка осуществляется следующей командой: `dapp dimg build-context import`.

*Примечание*. Для экспорта и импорта можно переопределить директорию, содержащую архивы с помощью опции `--build-context-directory`.
