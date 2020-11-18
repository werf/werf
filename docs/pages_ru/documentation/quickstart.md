---
title: Быстрый старт
permalink: documentation/quickstart.html
description: Разверните ваше первое приложение с werf
sidebar: documentation
---

В этой статье мы покажем, как развернуть простое [приложение](https://github.com/werf/quickstart-application) (для голосования в нашем случае) с помощью werf. Прежде всего рекомендуем ознакомиться с [кратким введением]({{ "introduction.html" | relative_url }}), если вы этого еще не сделали.

Чтобы повторить все шаги, изложенные в этом кратком руководстве, необходимо [установить werf]({{ "installation.html" | relative_url }}).

## Подготовьте свою инсталляцию Kubernetes и реестр Docker

У вас должен быть доступ к кластеру Kubernetes и возможность push'ить образы в Docker Registry. Для извлечение образов Docker Registry также должен быть доступен из кластера.

Если кластер Kubernetes и реестр Docker у вас уже настроены и работают, достаточно:

 1. Выполнить стандартный вход в реестр Docker со своего хоста.
 2. Убедиться, что кластер Kubernetes доступен с хоста (дополнительная настройка `werf`, скорее всего, не потребуется, если у вас уже установлен и работает `kubectl`).
 
<br>

<div class="details">
<a href="javascript:void(0)" class="details__summary">В ином случае выполните следующие действия, чтобы настроить локальный кластер Kubernetes и Docker Registry.</a>
<div class="details__content" markdown="1">
 1. Установите [minikube](https://github.com/kubernetes/minikube#installation).
 2. Запустите minikube:

    {% raw %}
    ```shell
    minikube start
    ```
    {% endraw %}

 3. Включите дополнение minikube registry:

    {% raw %}
    ```shell
    minikube addons enable registry
    ```
    {% endraw %}

 4. Запустите сервис с привязкой к порту 5000:

    {% raw %}
    ```shell
    kubectl -n kube-system expose rc/registry --type=ClusterIP --port=5000 --target-port=5000 --name=werf-registry --selector='actual-registry=true'
    ```
    {% endraw %}

 5. Запустите порт-форвардер на хосте в отдельном терминале:

    {% raw %}
    ```shell
    kubectl port-forward --namespace kube-system service/werf-registry 5000
    ```
    {% endraw %}
</div>
</div>

## Разверните приложение-пример 

 1. Склонируйте репозиторий нашего приложения-примера:
 
    {% raw %}
    ```shell
    git clone https://github.com/werf/quickstart-application
    cd quickstart-application
    ```
    {% endraw %}

 2. Запустите команду converge, используя Docker Registry для хранения образов (в случае локального репозитория это будет `localhost:5000/quickstart-application`).

    {% raw %}
    ```shell
    werf converge --repo localhost:5000/quickstart-application
    ```
    {% endraw %}

_Примечание: для подключения к кластеру Kubernetes `werf` использует те же настройки, что и `kubectl`: файл `~/.kube/config` и переменную среды `KUBECONFIG`. Также поддерживаются флаги `--kube-config` и `--kube-config-base64` - с их помощью можно указывать кастомные файлы kubeconfig._

## Проверьте результаты

После успешного завершения команды `converge` можно считать, что наше приложение развернуто  и работает. Давайте его проверим!

Как вы помните, наше приложение представляет собой простую голосовалку. Чтобы принять участие в голосовании, перейдите по соответствующей ссылке (в нашем случае это [http://172.17.0.3:31000](http://172.17.0.3:31000)):

{% raw %}
```
minikube service --namespace quickstart-application --url vote
```
{% endraw %}

Чтобы увидеть результаты голосования, перейдите по другой ссылке (в нашем случае это [http://172.17.0.3:31001](http://172.17.0.3:31001)):

{% raw %}
```
minikube service --namespace quickstart-application --url result
```
{% endraw %}

## Принципы работы

Чтобы развернуть приложение с помощью `werf`, необходимо описать желаемое состояние в Git (как описано во [введении]({{ "introduction.html" | relative_url }})).

 1. В нашем репозитории имеются следующие Dockerfile'ы:

    {% raw %}
    ```
    vote/Dockerfile
    result/Dockerfile
    worker/Dockerfile
    ```
    {% endraw %}

 2. В `werf.yaml` на них прописаны соответствующие ссылки:

    {% raw %}
    ```
    configVersion: 1
    project: quickstart-application
    ---
    image: vote
    dockerfile: vote/Dockerfile
    context: vote
    ---
    image: result
    dockerfile: result/Dockerfile
    context: result
    ---
    image: worker
    dockerfile: worker/Dockerfile
    context: worker
    ```
    {% endraw %}


 3. Шаблоны для компонентов приложения `vote`, `db`, `redis`, `result` и `worker` описаны в каталоге `.helm/templates/`. Схема ниже показывает, как компоненты взаимодействуют между собой:

  ![architecture](/images/quickstart-architecture.svg)

   - Фронтенд-приложение на Python или ASP.NET Core позволяет пользователю проголосовать за один из двух вариантов;
   - Очередь на базе Redis или NATS получает новые голоса;
   - Worker на основе .NET Core, Java или .NET Core 2.1 собирает голоса и сохраняет их в...
   - Базу данных Postgres или TiDB в томе Docker;
   - Веб-приложение на Node.js или ASP.NET Core SignalR в реальном времени показывает результаты голосования.

## Что дальше?

Рекомендуем ознакомиться со статьей ["Использование werf с системами CI/CD"]({{ "documentation/using_with_ci_cd_systems.html" | relative_url }}) или обратиться к соответствующим [руководствам]({{ "documentation/guides.html" | relative_url }}).
