---
title: Быстрый старт
permalink: quickstart.html
description: Разверните ваше первое приложение с werf
---

В этой статье мы покажем, как развернуть простое [приложение](https://github.com/werf/quickstart-application) (для голосования в нашем случае) с помощью werf. Прежде всего рекомендуем ознакомиться с [кратким введением](/introduction.html), если вы этого еще не сделали.

Чтобы повторить все шаги, изложенные в этом кратком руководстве, необходимо [установить werf](/installation.html).

## Подготовьте вашу систему

 1. Установите [зависимости](/installation.html#установка-зависимостей).
 2. Установите [multiwerf и werf](/installation.html#установка-werf/).

Прежде чем переходить к следующим шагам, надо убедиться что команда `werf` доступна в вашем shell:

```
werf version
```

## Подготовьте свою инсталляцию Kubernetes и container registry

У вас должен быть доступ к кластеру Kubernetes и возможность push'ить образы в container registry. Container registry также должен быть доступен из кластера для извлечения образов.

Если кластер Kubernetes и container registry у вас уже настроены и работают, достаточно:

 1. Выполнить стандартный вход в container registry со своего хоста.
 2. Убедиться, что кластер Kubernetes доступен с хоста (дополнительная настройка `werf`, скорее всего, не потребуется, если у вас уже установлен и работает `kubectl`).
 
<br>

В ином случае выполните одну из следующих инструкций, чтобы настроить локальный кластер Kubernetes и container registry в вашей системе:

<div class="details">
<a href="javascript:void(0)" class="details__summary">Windows — minikube</a>
<div class="details__content" markdown="1">
1. Установите [minikube](https://github.com/kubernetes/minikube#installation).
2. Запустите minikube:

   {% raw %}
   ```shell
   minikube start --driver=hyperv --insecure-registry registry.example.com:80
   ```
   {% endraw %}

   **ВАЖНО.** С параметром `--insecure-registry` мы подготавливаем такое окружение, которое сможет работать с Container Registry без TLS. В нашем случае для упрощения настройка TLS отсутствует.

3. Установка NGINX Ingress Controller:

   {% raw %}
   ```shell
   minikube addons enable ingress
   ```
   {% endraw %}

4. Установка Container Registry для хранения образов:

   {% raw %}
   ```shell
   minikube addons enable registry
   ```
   {% endraw %}

   Создадим Ingress для доступа к Container Registry:

   {% raw %}
   ```shell
   kubectl apply -f - << EOF
   ---
   apiVersion: networking.k8s.io/v1
   kind: Ingress
   metadata:
     name: registry
     namespace: kube-system
     annotations:
       nginx.ingress.kubernetes.io/proxy-body-size: "0"
   spec:
     rules:
     - host: registry.example.com
       http:
         paths:
         - path: /
           pathType: Prefix
           backend:
             service:
               name: registry
               port:
                 number: 80
   EOF
   ```
   {% endraw %}

5. Разрешаем доступ в Container Registry без TLS для docker:

   Через меню Docker Desktop -> Settings -> Docker Engine добавим новый ключ в конфигурацию:

   ```json
   {
      "insecure-registries": ["registry.example.com:80"]
   }
   ```

   Перезапустим Docker Desktop через меню, открывающееся правым кликом по иконке Docker Desktop в трее.

6. Разрешаем доступ в Container Registry без TLS для werf:

   В терминале где будет запускаться werf установим переменную окружения `WERF_INSECURE_REGISTRY=1`.

   Для cmd.exe:

   ```
   set WERF_INSECURE_REGISTRY=1
   ```

   Для bash:

   ```
   export WERF_INSECURE_REGISTRY=1
   ```

   Для PowerShell:

   ```
   $Env:WERF_INSECURE_REGISTRY = "1"
   ```

7. Мы будем использовать домены `vote.quickstart-application.example.com` и `result.quickstart-application.example.com` для доступа к приложению и домен `registry.example.com` для доступа к Container Registry.

   Обновим файл hosts. Сначала получите IP-адрес minikube:

   ```shell
   minikube ip
   ```

   Используя полученный выше IP-адрес minikube, добавьте в конец файла `C:\Windows\System32\drivers\etc\hosts` следующую строку:
   
   ```
   <IP-адрес minikube>    vote.quickstart-application.example.com result.quickstart-application.example.com registry.example.com
   ```

   Должно получиться примерно так:
   ```
   192.168.99.99          vote.quickstart-application.example.com result.quickstart-application.example.com registry.example.com
   ```

8. Также делаем доступ к домену `registry.example.com` из minikube node:

   ```shell
   minikube ssh -- "echo $(minikube ip) registry.example.com | sudo tee -a /etc/hosts"
   ```

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">MacOS — minikube</a>
<div class="details__content" markdown="1">
1. Установите [minikube](https://github.com/kubernetes/minikube#installation).
2. Запустите minikube:

   {% raw %}
   ```shell
   minikube start --vm=true --insecure-registry registry.example.com:80
   ```
   {% endraw %}

   **ВАЖНО.** С параметром `--insecure-registry` мы подготавливаем такое окружение, которое сможет работать с Container Registry без TLS. В нашем случае для упрощения настройка TLS отсутствует.

3. Установка NGINX Ingress Controller:

   {% raw %}
   ```shell
   minikube addons enable ingress
   ```
   {% endraw %}

4. Установка Container Registry для хранения образов:

   {% raw %}
   ```shell
   minikube addons enable registry
   ```
   {% endraw %}

   Создадим Ingress для доступа к Container Registry:

   {% raw %}
   ```shell
   kubectl apply -f - << EOF
   ---
   apiVersion: networking.k8s.io/v1
   kind: Ingress
   metadata:
     name: registry
     namespace: kube-system
     annotations:
       nginx.ingress.kubernetes.io/proxy-body-size: "0"
   spec:
     rules:
     - host: registry.example.com
       http:
         paths:
         - path: /
           pathType: Prefix
           backend:
             service:
               name: registry
               port:
                 number: 80
   EOF
   ```
   {% endraw %}

5. Разрешаем доступ в Container Registry без TLS для docker:

   Через меню Docker Desktop -> Settings -> Docker Engine добавим новый ключ в конфигурацию:
   
   ```json
   {
   "insecure-registries": ["registry.example.com:80"]
   }
   ```
   
   Перезапустим Docker Desktop через меню, открывающееся правым кликом по иконке Docker Desktop в трее.

6. Разрешаем доступ в Container Registry без TLS для werf:

   В терминале где будет запускаться werf установим переменную окружения `WERF_INSECURE_REGISTRY=1`. Для bash:

   ```shell
   export WERF_INSECURE_REGISTRY=1
   ```

   Чтобы опция автоматически устанавливалась в новых bash-сессиях, добавим её в `.bashrc`:

   ```shell
   echo export WERF_INSECURE_REGISTRY=1 | tee -a ~/.bashrc
   ```

7. Мы будем использовать домены `vote.quickstart-application.example.com` и `result.quickstart-application.example.com` для доступа к приложению и домен `registry.example.com` для доступа к Container Registry.

   Обновим файл hosts. Выполните команду в терминале:

   ```shell
   echo "$(minikube ip) vote.quickstart-application.example.com result.quickstart-application.example.com registry.example.com" | sudo tee -a /etc/hosts
   ```

8. Также делаем доступ к домену `registry.example.com` из minikube node:

   ```shell
   minikube ssh -- "echo $(minikube ip) registry.example.com | sudo tee -a /etc/hosts"
   ```

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">Linux — minikube</a>
<div class="details__content" markdown="1">
1. Установите [minikube](https://github.com/kubernetes/minikube#installation).
2. Запустите minikube:

   {% raw %}
   ```shell
   minikube start --driver=docker --insecure-registry registry.example.com:80
   ```
   {% endraw %}

   **ВАЖНО.** С параметром `--insecure-registry` мы подготавливаем такое окружение, которое сможет работать с Container Registry без TLS. В нашем случае для упрощения настройка TLS отсутствует.

3. Установка NGINX Ingress Controller:

   {% raw %}
   ```shell
   minikube addons enable ingress
   ```
   {% endraw %}

4. Установка Container Registry для хранения образов:

   {% raw %}
   ```shell
   minikube addons enable registry
   ```
   {% endraw %}

   Создадим Ingress для доступа к Container Registry:

   {% raw %}
   ```shell
   kubectl apply -f - << EOF
   ---
   apiVersion: networking.k8s.io/v1
   kind: Ingress
   metadata:
     name: registry
     namespace: kube-system
     annotations:
       nginx.ingress.kubernetes.io/proxy-body-size: "0"
   spec:
     rules:
     - host: registry.example.com
       http:
         paths:
         - path: /
           pathType: Prefix
           backend:
             service:
               name: registry
               port:
                 number: 80
   EOF
   ```
   {% endraw %}
   
5. Разрешаем доступ в Container Registry без TLS для docker:

   В файл, по умолчанию находящийся в `/etc/docker/daemon.json`, добавим новый ключ:

   ```json
   {
   "insecure-registries": ["registry.example.com:80"]
   }
   ```
   
   Перезапустим Docker:
   
   ```shell
   sudo systemctl restart docker
   ```

6. Разрешаем доступ в Container Registry без TLS для werf:

   В терминале где будет запускаться werf установим переменную окружения `WERF_INSECURE_REGISTRY=1`.

   Для bash:

   ```shell
   export WERF_INSECURE_REGISTRY=1
   ```

   Чтобы опция автоматически устанавливалась в новых bash-сессиях, добавим её в `.bashrc`:

   ```shell
   echo export WERF_INSECURE_REGISTRY=1 | tee -a ~/.bashrc
   ```

7. Мы будем использовать домены `vote.quickstart-application.example.com` и `result.quickstart-application.example.com` для доступа к приложению и домен `registry.example.com` для доступа к Container Registry.

   Обновим файл hosts. Выполните команду в терминале:

   ```shell
   echo "$(minikube ip) vote.quickstart-application.example.com result.quickstart-application.example.com registry.example.com" | sudo tee -a /etc/hosts
   ```

8. Также делаем доступ к домену `registry.example.com` из minikube node:

   ```shell
   minikube ssh -- "echo $(minikube ip) registry.example.com | sudo tee -a /etc/hosts"
   ```

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

2. Запустите команду converge, которяа использует Container Registry для хранения образов:

   {% raw %}
   ```shell
   werf converge --repo registry.example.com:80/quickstart-application
   ```
   {% endraw %}

_Примечание: для подключения к кластеру Kubernetes `werf` использует те же настройки, что и `kubectl`: файл `~/.kube/config` и переменную среды `KUBECONFIG`. Также поддерживаются флаги `--kube-config` и `--kube-config-base64` - с их помощью можно указывать кастомные файлы kubeconfig._

## Проверьте результаты

После успешного завершения команды `converge` можно считать, что наше приложение развернуто и работает.

Как вы помните, наше приложение представляет собой простую голосовалку. Давайте его проверим!

1. Чтобы принять участие в голосовании, перейдите по ссылке: [vote.quickstart-application.example.com](http://vote.quickstart-application.example.com)

2. Чтобы увидеть результаты голосования, перейдите по ссылке: [result.quickstart-application.example.com](http://result.quickstart-application.example.com)

## Принципы работы

Чтобы развернуть приложение с помощью `werf`, необходимо описать желаемое состояние в Git (как описано во [введении](/introduction.html)).

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

  ![Схема взаимодействия компонентов]({{ "images/quickstart-architecture.svg" | true_relative_url }})

   - Фронтенд-приложение на Python или ASP.NET Core позволяет пользователю проголосовать за один из двух вариантов;
   - Очередь на базе Redis или NATS получает новые голоса;
   - Worker на основе .NET Core, Java или .NET Core 2.1 собирает голоса и сохраняет их в...
   - Базу данных Postgres или TiDB в томе Docker;
   - Веб-приложение на Node.js или ASP.NET Core SignalR в реальном времени показывает результаты голосования.

## Что дальше?

Рекомендуем ознакомиться со статьей ["Использование werf с системами CI/CD"](using_with_ci_cd_systems.html) или обратиться к соответствующим [руководствам](/guides.html).
