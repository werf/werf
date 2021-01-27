---
title: Быстрый старт
permalink: documentation/quickstart.html
description: Разверните ваше первое приложение с werf
sidebar: documentation
---

В этой статье мы покажем, как развернуть простое [приложение](https://github.com/werf/quickstart-application) (для голосования в нашем случае) с помощью werf. Прежде всего рекомендуем ознакомиться с [кратким введением]({{ "introduction.html" | true_relative_url }}), если вы этого еще не сделали.

Чтобы повторить все шаги, изложенные в этом кратком руководстве, необходимо [установить werf]({{ "installation.html" | true_relative_url }}).

## Подготовьте вашу систему

 1. Установите [зависимости]({{ "installation.html#установка-зависимостей" | true_relative_url }}).
 2. Установите [multiwerf и werf]({{ "installation.html#установка-werf" | true_relative_url }}).

Прежде чем переходить к следующим шагам, надо убедиться что команда `werf` доступна в вашем shell:

```
werf version
```

## Подготовьте свою инсталляцию Kubernetes и Docker Registry

У вас должен быть доступ к кластеру Kubernetes и возможность push'ить образы в Docker Registry. Docker Registry также должен быть доступен из кластера для извлечения образов.

Если кластер Kubernetes и реестр Docker у вас уже настроены и работают, достаточно:

 1. Выполнить стандартный вход в реестр Docker со своего хоста.
 2. Убедиться, что кластер Kubernetes доступен с хоста (дополнительная настройка `werf`, скорее всего, не потребуется, если у вас уже установлен и работает `kubectl`).
 
<br>

В ином случае выполните одну из следующих инструкций, чтобы настроить локальный кластер Kubernetes и Docker Registry в вашей системе:

<div class="details">
<a href="javascript:void(0)" class="details__summary">Windows</a>
<div class="details__content" markdown="1">
 1. Установите [minikube](https://github.com/kubernetes/minikube#installation).
 2. Запустите minikube:

    {% raw %}
    ```shell
    minikube start --driver=docker
    ```
    {% endraw %}
    
    **ВАЖНО:** Если minikube уже запущен в вашей системе, то надо удостоверится, что используется driver под названием `docker`. Если нет, то требуется перезапустить minikube с помощью команды `minikube delete` и команды для старта, показанной выше.

 3. Включите дополнение minikube registry:

    {% raw %}
    ```shell
    minikube addons enable registry
    ```
    {% endraw %}
    
    Вывод команды должен содержать похожую строку:
 
    ```
    ...
    * Registry addon on with docker uses 32769 please use that instead of default 5000
    ...
    ```

    Запоминаем порт `32769`.    
    
 4. Запустите следующий проброс портов в отдельном терминале, заменив порт `32769` вашим портом из шага 3:

    ```shell
    docker run -ti --rm --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:host.docker.internal:32769"
    ```

 5. Запустите сервис с привязкой к порту 5000:

    {% raw %}
    ```shell
    kubectl -n kube-system expose rc/registry --type=ClusterIP --port=5000 --target-port=5000 --name=werf-registry --selector=actual-registry=true
    ```
    {% endraw %}

 6. Запустите следующий проброс портов в отдельном терминале:

    {% raw %}
    ```shell
    kubectl port-forward --namespace kube-system service/werf-registry 5000
    ```
    {% endraw %}
</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">MacOS</a>
<div class="details__content" markdown="1">
 1. Установите [minikube](https://github.com/kubernetes/minikube#installation).
 2. Запустите minikube:

    {% raw %}
    ```shell
    minikube start --driver=docker
    ```
    {% endraw %}
    
    **ВАЖНО:** Если minikube уже запущен в вашей системе, то надо удостоверится, что используется driver под названием `docker`. Если нет, то требуется перезапустить minikube с помощью команды `minikube delete` и команды для старта, показанной выше.

 3. Включите дополнение minikube registry:

    {% raw %}
    ```shell
    minikube addons enable registry
    ```
    {% endraw %}
    
    Вывод команды должен содержать похожую строку:
 
    ```
    ...
    * Registry addon on with docker uses 32769 please use that instead of default 5000
    ...
    ```

    Запоминаем порт `32769`.    
    
 4. Запустите следующий проброс портов в отдельном терминале, заменив порт `32769` вашим портом из шага 3:

    ```shell
    docker run -ti --rm --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:host.docker.internal:32769"
    ```

 5. Запустите следующий проброс портов в отдельном терминале, заменив порт `32769` вашим портом из шага 3:
 
    ```shell
    brew install socat
    socat TCP-LISTEN:5000,reuseaddr,fork TCP:host.docker.internal:32769
    ```
</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">Linux</a>
<div class="details__content" markdown="1">
 1. Установите [minikube](https://github.com/kubernetes/minikube#installation).
 2. Запустите minikube:

    {% raw %}
    ```shell
    minikube start --driver=docker
    ```
    {% endraw %}
    
    **ВАЖНО:** Если minikube уже запущен в вашей системе, то надо удостоверится, что используется driver под названием `docker`. Если нет, то требуется перезапустить minikube с помощью команды `minikube delete` и команды для старта, показанной выше.

 3. Включите дополнение minikube registry:

    {% raw %}
    ```shell
    minikube addons enable registry
    ```
    {% endraw %}
 
 4. Запустите следующий проброс портов в отдельном терминале:

    {% raw %}
    ```shell
    sudo apt-get install -y socat
    socat -d -d TCP-LISTEN:5000,reuseaddr,fork TCP:$(minikube ip):5000
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

Как вы помните, наше приложение представляет собой простую голосовалку. Чтобы принять участие в голосовании, перейдите по ссылке, которую выдаст следующая команда:

{% raw %}
```
minikube service --namespace quickstart-application --url vote
```
{% endraw %}

Чтобы увидеть результаты голосования, перейдите по ссылке, которую выдаст следующая команда:

{% raw %}
```
minikube service --namespace quickstart-application --url result
```
{% endraw %}

## Принципы работы

Чтобы развернуть приложение с помощью `werf`, необходимо описать желаемое состояние в Git (как описано во [введении]({{ "introduction.html" | true_relative_url }})).

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

  ![architecture]({{ "images/quickstart-architecture.svg" | true_relative_url }})

   - Фронтенд-приложение на Python или ASP.NET Core позволяет пользователю проголосовать за один из двух вариантов;
   - Очередь на базе Redis или NATS получает новые голоса;
   - Worker на основе .NET Core, Java или .NET Core 2.1 собирает голоса и сохраняет их в...
   - Базу данных Postgres или TiDB в томе Docker;
   - Веб-приложение на Node.js или ASP.NET Core SignalR в реальном времени показывает результаты голосования.

## Что дальше?

Рекомендуем ознакомиться со статьей ["Использование werf с системами CI/CD"]({{ "documentation/using_with_ci_cd_systems.html" | true_relative_url }}) или обратиться к соответствующим [руководствам]({{ "documentation/guides.html" | true_relative_url }}).
