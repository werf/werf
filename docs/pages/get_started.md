---
title: Первое приложение на dapp
sidebar: doc_sidebar
permalink: get_started.html
---

# Начало работы с dapp

* Быстрый старт на примере простого php приложения
* Dappfile как замена Dockerfile
* Зачем нужен dapp?

Здесь будет описана сборка приложения с помощью утилиты dapp. Перед изучением dapp желательно представлять, что такое Dockerfile и его основные директивы  (https://docs.docker.io/).

Для запуска примеров понадобятся:
* dapp (Установка описана [здесь])
* docker версии не ниже 1.10
* git

## Сборка простого приложения

Начнём с простого приложения на php. Создайте директорию для тестов и склонируйте репозиторий:

```
git clone https://github.com/awslabs/opsworks-demo-php-simple-app
```

Это совсем небольшое приложение с одной страницей и статическими файлами. Чтобы приложение можно было просто запустить, нужно запаковать его в контейнер, например, с php и apache. Для этого достаточно такого Dockerfile.

```
$ vi Dockerfile
FROM php:7.0-apache

COPY . /var/www/html/

EXPOSE 80
EXPOSE 443
```

Чтобы собрать и запустить приложение нужно выполнить:

```
$ docker build -t simple-app-v1 .
$ docker run -d --name simple-app simple-app-v1
```

Проверить можно либо зайдя браузером на порт 80, либо локально через curl:

```
$ docker exec -ti simple-app bash
root@da234e2a7777:/var/www/html# curl 127.0.0.1
...
                <h1>Simple PHP App</h1>
                <h2>Congratulations!</h2>
...
```

Теперь соберём образ с помощью dapp. Для этого нужно создать `Dappfile`.

* В репозитории могут находится одновременно и Dappfile и Dockerfile - они друг другу не мешают.
* Среди директив Dappfile есть семейство docker.* директив, которые повторяют аналогичные из Dockerfile.

```
$ vi Dappfile

dimg 'simple-php-app' do
  docker.from 'php:7.0-apache'

  git do
    add '/' do
      to '/var/www/html'
      include_paths '*.php', 'assets'
    end
  end

  docker do
    expose 80
    expose 443
  end
end
```

Рассмотрим подробнее этот файл.

`dimg` — эта директива определяет образ, который будет собран. Аргумент `simple-php-app` — имя этого образа, его можно увидеть, запустив `dapp list`

`docker.from` — аналог директивы `FROM`. Определяет образ, на основе которого будет собираться образ приложения.

`git` — директива, на первый взгляд аналог директив `ADD` или `COPY`, но с более тесной интеграцией с git. Подробнее про то, как dapp работает с git, можно прочесть в отдельной главе, а сейчас для нас главное увидеть, что директива `git` и вложенная директива `add` позволяют копировать содержимое локального git-репозитория в образ. Копирование производится из пути, указанного в `add`. `'/'` означает, что копировать нужно из корня репозитория. `to` задаёт конечную директорию в образе, куда попадут файлы. С помощью `include_paths` и `exclude_paths` можно задавать, какие именно файлы нужно скопировать или какие нужно пропустить.

`docker do` — директива `docker`, как и многие другие директивы `Dappfile`, имеет блочную запись, с помощью которой можно объединять несколько директив в короткой записи. Т.е. эти два определения эквивалентны:
```
# 1.
docker do
  expose 80
  expose 443
end
```

```
# 2.
docker.expose 80
docker.expose 443
```

Для сборки нужно выполнить команду `dapp build`

```
$ dapp build
simple-php-app
  From ...                                                   [OK] 0.55 sec
  Git artifacts dependencies ...                             [OK] 0.45 sec
  Git artifacts: create archive ...                          [OK] 0.52 sec
  Install group
    Git artifacts dependencies ...                           [OK] 0.39 sec
    Git artifacts: apply patches (after install) ...         [OK] 0.41 sec
  Setup group
    Git artifacts dependencies ...                           [OK] 0.42 sec
    Git artifacts: apply patches (before setup) ...          [OK] 0.69 sec
    Git artifacts dependencies ...                           [OK] 0.41 sec
    Git artifacts: apply patches (after setup) ...           [OK] 0.4 sec
  Git artifacts: latest patch ...                            [OK] 0.39 sec
  Docker instructions ...                                    [OK] 0.42 sec
```


```
$ dapp run -d --name simple-php-app
59ae767d497b4e4fb8c32cd97110cc0f17e67d8e3c7f540cef73b713ef995e5a
```
$ docker ps
CONTAINER ID        IMAGE                                                                                                     COMMAND                  CREATED             STATUS              PORTS                NAMES
ef6a519b7e9c        dimgstage-opsworks-demo-php-simple-app:351de177ee64c448767fd7525ec46e066d89791803109e766aef131490be38de   "docker-php-entrypoin"   4 seconds ago       Up 3 seconds        80/tcp               simple-php-app
```

Теперь можно проверить, как и ранее:

```
$ docker exec -ti simple-php-app bash
root@ef6a519b7e9c:/var/www/html# curl 127.0.0.1
...
                <h1>Simple PHP App</h1>
                <h2>Congratulations!</h2>
...
```

Ура! Первая сборка с помощью `dapp` прошла успешно

## Зачем нужен dapp?

Простое приложение показало, что `Dappfile` может использоваться как замена Dockerfile. Но в чём же плюсы, кроме синтаксиса, немного похожего на Vagrantfile?


Причесать ниже:

Гибкость там, где она нужна на самом деле — копирование файлов в образ
В корне репозитория есть файлы, которые не нужны в финальном образе. Нужно либо использовать .dockerignore, либо структурировать код, разложив по директориям. dapp позволяет выборочно скопировать нужные файлы в нужные директории. Директиву git можно использовать несколько раз.


# docker exec -ti simpl bash
root@725a5ae61322:/var/www/html# ls
Dockerfile  LICENSE.md	NOTICE.md  README.md  assets  index.php

dapp с git + include_paths
# docker exec -ti stoic_jennings bash
root@59ae767d497b:/var/www/html# ls
assets index.php


С помощью использования директив git мы можем положить статические файлы (assets) в одну директорию, а исполнимые (*.php) в другую.

Замечательно, но так раскладывать по директориям умеет и Dockerfile c помощью директив ADD и COPY.

Ок. Допустим, нужно внести правки в index.php, скажем поменять заголовок страницы. Как обычно, редактором меняется заголовок, файл сохраняется, запускается git commit -a -m “new version: v2”

Соберём образ ещё раз с помощью docker:






Чтобы ответить на этот вызов, нужно рассмотреть методы кэширования сборки docker-ом.

дальше про

и про то, что git можно заставить копировать файлы


Были такие идеи старта:
с помощью директивы git можно сделать структуру приложения. ADD И COPY это тоже могут
ADD и COPY не зависят от изменений в файлах — оказалось, что зависят

На этом всё застопорилось, нужно либо про стадии начать рассказывать прямо в этой части, либо про артефакты (как собрать с внешними утилитами, не занимаясь RUN “download-source && cmd && cmd2 && remove-source”


Копируем патчами! Вот в чём смысл.



RUN “download-source && cmd && cmd2 && remove-source”
А вот это можно решить с помощью
