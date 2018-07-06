---
title: Как устроено кэширование
sidebar: doc_sidebar
permalink: cache_for_advanced_build.html
folder: advanced_build
author: Alexey Igrychev <alexey.igrychev@flant.com>
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

Стандартный процесс сборки приложения через dapp представляет собой сборку набора docker-образов. В процессе сборки docker-образы тех стадий, которые собрались остаются "невидимыми" пользователю dapp, они имеют только временный идентификатор в docker. В случае ошибки сборки на какой-либо стадии, dapp не удаляет из сборочного кэша приложения успешно собранные образы предыдущих стадий. При повторной сборке, они будут использованы, в случае если сигнатура стадии не изменилась. Технически docker-образ попадает в сборочный кэш в тот момент, когда dapp тегирует этот образ специальным образом: используя `dimgstage-<имя-репозитория-приложения>` в качестве имени образа и сигнатуры стадии в качестве тега образа. Таким образом, стандартное поведение dapp предполагает, что при сборке формируется один docker-образ на все инструкции каждой стадии.

На этапе разработки dappfile, при выполнении ресурсоемких инструкций и перезапуске сборки ошибочной стадии, это может требовать значительных временных затрат. Для удобства разработки и отладки в dapp была введена **альтернативная схема кэширования**, которая доступна **только в [YAML-конфигурации](yaml.html)** и включается директивой `asLayers`.

Директива `asLayers` может быть указана в конфигурации (dappfile.yaml) для конкретного приложения или его артефакта. При включении альтернативной схемы кэширования, dapp формирует один docker-образ на каждую команду для shell или каждый task для Ansible, что приводит к тому, что инструкции при сборке кэшируются по отдельности. Пересборка образа в этом случае осуществляется только при изменении порядка инструкций.

**Важно**. Альтернативная схема кэширования порождает избыточное количество docker-образов и не рассчитана на инкрементальную сборку (увеличивается время ожидания и размер сборочного кэша) - она предназначена для разработки, по-этому ее не нужно использовать в продуктивных средах.

Для облегчения локальной разработки, в dapp предусмотрен специальный **режим разработчика**, отличие которого, с точки зрения кэширования, - в наличии отдельного кэша. Подробнее познакомиться с режимом разработчика можно в соответствующем [разделе](debug_for_advanced_build.html#режим-разработчика).

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

Уменьшения размера конечного образа приложения можно добиться с помощью монтирования в сборочный образ внешних папок. Для монтирования внешних папок в dappfile предусмотрена [директива mount](mount_directives.html), которая позволяет монтировать внутрь сборочного контейнера не только произвольные директории, но и служебные директории двух типов:
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
