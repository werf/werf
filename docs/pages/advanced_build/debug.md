---
title: Debug
sidebar: doc_sidebar
permalink: debug_for_advanced_build.html
folder: advanced_build
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

### Интроспекция стадий в интерактивном режиме

Контролировать результат работы правил сборки образа позволяет возможность интроспекции стадий. Интроспекция представляет собой запуск shell-сессии для пользователя в интерактивном режиме в собираемом образе.

В общем случае, если падает набор команд, создающих образ стадии Y из образа стадии X, то:

* c опцией `--introspect-before-error` пользователь попадет в контейнер с образом X;
* с опцией `--introspect-error` пользователь попадет в сборочный контейнер в состоянии сразу после исполнения упавшей команды.

Например, есть Dappfile с ошибкой:

```ruby
dimg do
  docker.from 'ubuntu:16.04'

  shell.before_install do
    run 'apt-get update'
  end

  shell.install do
    run 'apt-get install -y nginx'
  end
end
```

При запуске сборки dapp упадет с подобным сообщением:

```shell
$ dapp dimg build
From ...                                                                              [OK] 0.95 sec
Before install ...                                                                    [OK] 10.98 sec

Install group
  Install ...   Launched command: `apt-get install -y nginx`
                                                                                      [FAILED] 1.93 sec
Stacktrace dumped to /tmp/dapp-stacktrace-736a2035-4c8e-4ee3-9b55-8cfe5b4704a0.out
>>> START STREAM
Reading package lists...

Building dependency tree...

Reading state information...
E: Unable to locate package nginx
>>> END STREAM
```

Произошло следующее: упал набор команд, собирающих образ стадии Install. С помощью опции `--introspect-error` для команды сборки пользователь получает доступ в сборочный контейнер в состоянии сразу после исполнения упавшей команды:

```shell
$ dapp dimg build --introspect-error
From ...                                                                              [OK] 0.9 sec
Before install ...                                                                    [OK] 10.24 sec
Install group
  Install ...   Launched command: `apt-get install -y nginx`
                                                                                      [FAILED] 1.91 sec
root@18ae29cf201a:/# apt-get install -y nginx
Reading package lists... Done
Building dependency tree       
Reading state information... Done
E: Unable to locate package nginx
root@18ae29cf201a:/# apt-get install -y nginx
...
root@18ae29cf201a:/# exit
Stacktrace dumped to /tmp/dapp-stacktrace-4ecac017-bcaf-4304-b6e7-fe7ca481c7af.out
>>> START STREAM
Reading package lists...

Building dependency tree...

Reading state information...
E: Unable to locate package nginx
>>> END STREAM
```

В данном контейнере можно вручную выполнить команды, просмотреть состояние системы и понять в чем проблема.

Если использовать опцию `--introspect-before-error` для команды сборки, то пользователь соответственно получит доступ в сборочный контейнер для стадии, предшествующей Before install. Т.е. ни одна команда для сборки стадии Before install в данном контейнере еще не будет выполнена.

### Интроспекция стадий (introspect-stage, introspect-artifact-stage) после успешной сборки

Во время разработки Dappfile, часто требуется запустить сборку, затем вручную проверить результат сборки стадии. В случае, если ошибок при сборке стадии не произошло, для этого используются опции `--introspect-stage=<stage>` для обычного образа и `--introspect-artifact-stage=<stage>` для образа артефакта. Возможно указание лишь одной стадии, для которой нужна интроспекция. Опции интроспекции стадии при возникновении ошибок и интроспекции успешно собранной стадии можно указывать одновременно. 

Возможные значения опции introspect-stage:

* from
* before_install
* before_install_artifact
* g_a_archive
* g_a_pre_install_patch
* install, g_a_post_install_patch
* after_install_artifact
* before_setup
* before_setup_artifact
* g_a_pre_setup_patch
* setup
* g_a_post_setup_patch
* after_setup_artifact
* g_a_latest_patch
* docker_instructions

Возможные значения опции introspect-artifact-stage:

* from
* before_install
* before_install_artifact
* g_a_archive
* g_a_pre_install_patch
* install
* g_a_post_install_patch
* after_install_artifact
* before_setup
* before_setup_artifact
* g_a_pre_setup_patch
* setup
* after_setup_artifact
* g_a_artifact_patch
* build_artifact

### Сборочный кэш и режим разработчика

Сборщик образов в dapp создает промежуточные docker образа после каждой успешной сборки стадии. Однако, создаваемые в процессе сборки образа являются скрытыми от пользователя dapp до того момента, как сборка закончится успешно. После успешной сборки все промежуточные образа именуются и попадают тем самым в кэш образов. Образа из кэша используются при повторных сборках, а также их можно интроспектить вручную через docker run.

Минусом данного механизма является то, что если в процессе сборки некоторой стадии произошла ошибка, то при повторном запуске сборка начнется с нуля, несмотря на то, что стадии до ошибочной были собраны успешно, т.к. образа не будут сохранены в кэше. Для приведенного выше примера при каждом повторном запуске сборки стадия Before install будет пересобираться по-новой.

Для разработчика конфигурации Dappfile было бы удобнее, если бы все успешно собранные стадии сразу сохранялись в кэш docker образов. В таком случае, при возникновении ошибки, пересборка бы всегда начиналась с ошибочной стадии. Этой цели служит режим разработчика, включаемый опцией --dev для всех команд, связанных с работой с образами. При этом кэш образов, создаваемых при сборке в режиме разработчика будет отдельным от основного кэша образов. Для приведенного выше примера использование режима разработчика будет выглядеть так:

```shell
$ dapp dimg build --dev
Before install ...                                                                    [OK] 21.8 sec
Install group
  Install ...   Launched command: `apt-get install -y nginx`
                                                                                      [FAILED] 1.83 sec
Stacktrace dumped to /tmp/dapp-stacktrace-f25448cf-085f-4e1b-8628-7c3288e7a5cf.out
>>> START STREAM
Reading package lists...

Building dependency tree...

Reading state information...
E: Unable to locate package nginx
>>> END STREAM
$ dapp dimg build --dev
Install group
  Install ...   Launched command: `apt-get install -y nginx`
                                                                                      [FAILED] 2.03 sec
Stacktrace dumped to /tmp/dapp-stacktrace-80a0d7a2-7448-4112-85ff-db5da7ba47fb.out
>>> START STREAM
Reading package lists...

Building dependency tree...

Reading state information...
E: Unable to locate package nginx
>>> END STREAM
```

Как видим, при повторном запуске стадия Before install более не пересобирается, т.к. была закэширована.
