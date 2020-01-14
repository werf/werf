---
title: Первые шаги
sidebar: documentation
permalink: documentation/guides/getting_started.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Обзор задачи

В статье рассматривается как быстро начать работу с werf, используя существующий в проекте Dockerfile. Мы выполним сборку Docker-образа приложения и загрузим его в Docker registry. В качестве примера приложения используется [Linux Tweet App](https://github.com/dockersamples/linux_tweet_app).

## Требования

* Минимальные знания [Docker](https://www.docker.com/) и структуры [Dockerfile](https://docs.docker.com/engine/reference/builder/).
* Установленные [зависимости werf]({{ site.baseurl }}/documentation/guides/installation.html#install-dependencies).
* Установленный [multiwerf](https://github.com/flant/multiwerf).

### Выбор версии werf

Перед началом работы с werf необходимо выбрать версию werf, которую вы будете использовать. Для выбора актуальной версии werf в канале stable, релиза 1.0, выполним следующую команду:

```shell
. <(multiwerf use 1.0 stable)
```

## Шаг 1: Конфигурация werf.yaml

Добавим файл `werf.yaml`, описывающий конфигурацию сборки образа приложения с использованием существующего в проекте [Dockerfile](https://github.com/dockersamples/linux_tweet_app/blob/master/Dockerfile).

1.  Склонируем репозиторий приложения [Linux Tweet App](https://github.com/dockersamples/linux_tweet_app):

    ```shell
    git clone https://github.com/dockersamples/linux_tweet_app.git
    cd linux_tweet_app
    ```

1.  В корневой папке приложения создадим файл `werf.yaml` со следующим содержимым:

    ```yaml
    project: g-started
    configVersion: 1
    ---
    image: ~
    dockerfile: Dockerfile
    ```

## Шаг 2: Сборка приложения

1.  Соберём образ приложения, выполнив команду в корневой папке:

    ```shell
    werf build --stages-storage :local
    ```

1.  Запустим контейнер на основе собранного образа:

    ```shell
    werf run --stages-storage :local --docker-options="-d -p 80:80"
    ```

1.  Проверим, что приложение запустилось и отвечает корректно, открыв в web-браузере `http://localhost:80` либо выполнив:

    ```shell
    curl localhost:80
    ```

## Шаг 3: Публикация образа в Docker registry

1.  Запустим Docker registry локально:

    ```shell
    docker run -d -p 5000:5000 --restart=always --name registry registry:2
    ```

2.  Загрузим образ приложения в Docker registry, предварительно протегировав его тегом `v0.1.0`:

    ```shell
    werf publish --stages-storage :local --images-repo localhost:5000/g-started --tag-custom v0.1.0
    ```

## Что дальше?

Вначале ознакомьтесь с документацией по теме:
* [werf configuration file]({{ site.base_url}}/documentation/configuration/introduction.html).
* [Dockerfile Image: complete directive list]({{ site.base_url}}/documentation/configuration/dockerfile_image.html).
* [Build procedure]({{ site.base_url}}/documentation/reference/build_process.html).
* [Publish procedure]({{ site.base_url}}/documentation/reference/publish_process.html).

Либо переходите к знакомству со следующими примерами:
* [Deploy an application to a Kubernetes cluster]({{ site.base_url}}/documentation/guides/deploy_into_kubernetes.html).
* [Advanced build with Stapel image]({{ site.base_url}}/documentation/guides/advanced_build/first_application.html).
