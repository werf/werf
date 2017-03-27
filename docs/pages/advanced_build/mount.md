---
title: Уменьшаем размеры образа и включаем кеширование с помощью mount
sidebar: doc_sidebar
permalink: mount_for_advanced_build.html
folder: advanced_build
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Бывает, что в процессе сборки образа возникают вспомогательные файлы в файловой системе, которые необходимо исключить из итогового образа. Например:

* Большинство пакетных менеджеров дистрибутивов оставляют создают общесистемный кеш пакетов и других служебных файлов.
  * При обновлении списка пакетов дистрибутива apt сохраняет его в директории /var/lib/apt/lists.
  * При установке пакетов apt оставляет скачанные пакеты в директории /var/cache/apt/.
  * При установке пакетов yum может оставлять скачанные пакеты в директории /var/cache/yum/.../packages.
* Такая же ситуация может быть с пакетными менеджерами языков программирования (npm для nodejs, glide для go, pip для python и прочие).
* Сборка c, c++ и подобных приложений оставляет после себя объектные файлы и другие файлы используемых систем сборок.

Полученные таким образом файлы:

* как правило не нужны в конечных образах, т.к. не используются;
* могут значительно увеличивать размер образа;
* могут понадобиться как кеш при последующих сборках новых версий образов.

Для решения подобных проблем существует возможность монтирования внешней директории в сборочный образ с помощью директивы mount. Например, есть Dappfile:

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

В следующем Dappfile при каждом новом запуске сборки в `/var/lib/apt/lists` будет чистая директория. Указание `from :tmp_dir` означает использование в качестве хранилища временной директории, выделяемой на каждую новую сборку образов.

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

После обновления списка пакетов в сборочном контейнере эта директория заполнится. Однако в итоговый образ эти файлы не попадут. В итоговом образе будет пустая директория `/var/lib/apt/lists`.

### Исключение файлов полученных из базового образа

Не всегда есть возможность контролировать создание базового образа и в некоторых образах оказываются зашиты лишние вспомогательные файлы. Использование директории в качестве mount-point автоматически очищает эту директорию, даже если в ней что-то было до этого. Т.е. в примере выше `/var/lib/apt/lists` будет очищен, даже если в образе, указанном в `docker.from`, эта директория не пуста. Принудительную очистку делает dapp на стадии From.

### Кеширование сборочных файлов между сборками

Чтобы включить кеширование содержимого примонтированной директории между сборками используется параметр `from :build_dir`. В отличие от `from :tmp_dir`, при каждом новом запуске сборки указанная директория будет сохраняться при условии, что служебная build-директория (`.dapp_build`) между сборками не была удалена. Т.е. в качестве хранилища будет выступать служебная build-директория. Чтобы при пересборке стадии before_install apt не скачивал по-новой устанавливаемые пакеты, также примонтируем директорию `/var/cache/apt` с использованием `from :build_dir`. Также отключим стандартную автоочистку скачанных пакетов в `/etc/apt/apt.conf.d/docker-clean`.

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

В следующем примере происходит установка rails в образ с использованием rvm. Для rvm включен кеш gem'ов, поэтому стоит примонтировать директорию `/usr/local/rvm/gems/cache`, используя `from :build_dir`.

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

В итоговом образе видим, что все примонтированные директории пусты внутри образа, а содержимое директорий, использующих build_dir в качестве хранилища, остается вне образа, в .dapp_build:

```shell
root@a96aa4d220c3:/# ls /var/lib/apt/lists /var/cache/apt/archives /usr/local/rvm/gems/cache
/usr/local/rvm/gems/cache:

/var/cache/apt/archives:

/var/lib/apt/lists:
```

```shell
$ tree .dapp_build/mount
.dapp_build/mount
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
