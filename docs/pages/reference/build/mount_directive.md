---
title: Using mounts
sidebar: reference
permalink: reference/build/mount_directive.html
---

```yaml
mount:
- from: build_dir
  to: <absolute_path>
- from: tmp_dir
  to: <absolute_path>
- fromPath: <absolute_path>
  to: <absolute_path>
```

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

