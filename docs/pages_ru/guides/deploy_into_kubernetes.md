---
title: Деплой в Kubernetes
sidebar: documentation
permalink: documentation/guides/deploy_into_kubernetes.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

## Обзор задачи

В статье рассматривается выкат приложения в Kubernetes с помощью werf. 
Для деплоя приложений в Kubernetes werf использует [Helm](https://helm.sh) (с некоторыми изменениями и дополнениями). 
В статье мы создадим простое web-приложение, соберем все необходимые для него образы, создадим Helm-шаблоны и выкатим приложение в кластер Kubernetes.

## Требования

 * Работающий кластер Kubernetes. Для выполнения примера вы можете использовать как обычный Kubernetes-кластер, так и Minikube. Если вы решили использовать Minikube, прочитайте [статью о настройке Minikube]({{ site.baseurl }}/documentation/reference/development_and_debug/setup_minikube.html), чтобы запустить Minikube и Docker registry.
 * Работающий Docker registry.
   * Доступ от хостов Kubernetes с правами на push образов в Docker registry.
   * Доступ от хостов Kubernetes с правами на pull образов в Docker registry.
 * Установленные [зависимости werf]({{ site.baseurl }}/documentation/guides/installation.html#install-dependencies).
 * Установленный [multiwerf](https://github.com/flant/multiwerf).
 * Установленный и сконфигурированный `kubectl` для доступа в кластер Kubernetes (<https://kubernetes.io/docs/tasks/tools/install-kubectl/>).

**Внимание!** Далее в качестве адреса репозитория будет использоваться значение `:minikube` . Если вы используете собственный кластер Kubernetes и Docker registry, то указывайте репозиторий проекта в Docker registry вместо аргумента `:minikube`.

### Выбор версии werf

Перед началом работы необходимо выбрать версию werf. Для выбора актуальной версии werf в канале stable, релиза 1.0, выполним следующую команду:

```shell
. $(multiwerf use 1.0 stable --as-file)
```

## Архитектура приложения

Пример представляет собой простейшее web-приложение, для запуска которого нам нужен только web-сервер. 
Архитектуру приложения в Kubernetes можно представить следующим образом:

     .----------------------.
     | backend (Deployment) |
     '----------------------'
                |
                |
      .--------------------.
      | frontend (Ingress) |
      '--------------------'

Здесь `backend` — web-сервер с приложением, `frontend` — прокси-сервер, который выступает точкой входа и используется для перенаправления внешнего трафика.

## Файлы приложения

Создадим пустую директорию проекта и перейдём в неё для выполнения следующих шагов:

```shell
mkdir myapp
cd myapp
```

werf ожидает, что все файлы, необходимые для сборки и развертывания приложения, находятся в папке приложения (папке проекта) вместе с исходным кодом, если он имеется. В нашем примере в этой директорию будут храниться только конфигурации.

## Подготовка образа

Необходимо подготовить образ приложения с web-сервером внутри. 
Для этого создадим файл `werf.yaml` в папке приложения со следующим содержимым:

```yaml
project: myapp
configVersion: 1
---

image: ~
from: python:alpine
ansible:
  install:
  - file:
      path: /app
      state: directory
      mode: 0755
  - name: Prepare main page
    copy:
      content:
        <!DOCTYPE html>
        <html>
          <body>
            <h2>Congratulations!</h2>
            <img src="https://flant.com/images/logo_en.png" style="max-height:100%;" height="76">
          </body>
        </html>
      dest: /app/index.html
```

Web-приложение состоит из единственной статической HTML-страницы, которая задаётся в инструкциях и добавляется при сборке образа. 
Содержимое этой страницы будет отдавать Python HTTP-сервер.

Соберём образ приложения и загрузим его в Docker registry:

```shell
werf build-and-publish --stages-storage :local --tag-custom myapp --images-repo :minikube
```

Название собранного образа приложения состоит из адреса Docker registry (`REPO`) и тега (`TAG`). 
При указании `:minikube` в качестве адреса Docker registry werf использует адрес `werf-registry.kube-system.svc.cluster.local:5000/myapp`. 
Так как в качестве тега был указан `myapp`, werf загрузит в Docker registry следующий образ `werf-registry.kube-system.svc.cluster.local:5000/myapp:myapp`.

## Подготовка конфигурации деплоя

werf использует встроенный [Helm](helm.sh) *для применения* конфигурации в Kubernetes. 
Для *описания* объектов Kubernetes werf использует конфигурационные файлы Helm: шаблоны и файлы с параметрами (например, `values.yaml`). 
Помимо этого, werf поддерживает дополнительные файлы, такие как — файлы секретами и с секретными значениями (например `secret-values.yaml`), а также дополнительные Go-шаблоны для интеграции собранных образов.

### Backend

Создадим файл конфигурации backend `.helm/templates/010-backend.yaml` (далее мы рассмотрим его подробнее):

{% raw %}
```yaml
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}-backend
spec:
  replicas: 4
  template:
    metadata:
      labels:
        service: {{ .Chart.Name }}-backend
    spec:
      containers:
      - name: backend
        workingDir: /app
        command: [ "python3", "-m", "http.server", "8080" ]
{{ werf_container_image . | indent 8 }}
        livenessProbe:
          httpGet:
            path: /
            port: 8080
            scheme: HTTP
        readinessProbe:
          httpGet:
            path: /
            port: 8080
            scheme: HTTP
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        env:
{{ werf_container_env . | indent 8 }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Chart.Name }}-backend
spec:
  clusterIP: None
  selector:
    service: {{ .Chart.Name }}-backend
  ports:
  - name: http
    port: 8080
    protocol: TCP
```
{% endraw %}

В конфигурации описываются Deployment `myapp-backend` и сервис для доступа к pod'ам. Особое внимание стоит уделить функциям werf `werf_container_image` и `werf_container_env`. 

Функция `werf_container_image` позволяет добавить поле `image` с корректным именем образа в конфигурацию, используя контекст и опциональное имя из `werf.yaml` в качестве параметров. 
В нашем случае образ безымянный (`~`), поэтому функция принимает только контекст без имени. 
Функция добавит тег на основе используемой схемы тегирования. 
В зависимости от схемы тегирования функция также может вернуть поле `imagePullPolicy` со определённым значением.
Более подробно о функции можно прочитать [в соответствующей статье]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_image).

Функция `werf_container_env` позволяет сгенерировать служебные значения в секции `env`, которые влияют на перезапуск pod'ов при изменении образа.
В качестве аргументов также принимаются контекст и имя образа из `werf.yaml`.
Подробнее про логику работы и особенности можно прочитать в [статье]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_env)).

### Frontend

Создадим файл конфигурации frontend `.helm/templates/090-frontend.yaml` со следующим содержимым:

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: myapp-frontend
  annotations:
    kubernetes.io/ingress.class: "nginx"
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
spec:
  rules:
  - host: myapp.local
    http:
      paths:
      - path: /
        backend:
          serviceName: myapp-backend
          servicePort: 8080
```

Эта конфигурация описывает Ingress и настраивает NGINX прокси-сервер. 
По описанному правилу трафик для хоста `myapp.local` будет перенаправляться на backend-сервер `myapp-backend`.

## Деплой

При использовании `minikube` перед деплоем необходимо включить ingress-модуль:

```shell
minikube addons enable ingress
```

Наконец, запустим деплой:

```shell
werf deploy --stages-storage :local --images-repo :minikube --tag-custom myapp --env dev
```

После запуска команды, werf создаст соответствующие ресурсы в Kubernetes и будет отслеживать статус Deployment `myapp-backend` до его готовности (готовности всех pod'ов) либо ошибки.

Для того чтобы сформировать правильные имя Helm-релиза и namespace требуется указать [окружение]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#environment) с помощью параметра `--env`.

В результате будет создан Helm-релиз с именем `myapp-dev`. 
Название Helm-релиза состоит из [имени проекта]({{ site.baseurl }}/documentation/configuration/introduction.html#meta-configuration-doc) `myapp` (указанного в `werf.yaml`) и переданного названия окружения — `dev`. 
Более подробно про формирование имен Helm-релизов можно посмотреть в [документации]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#release-name).

При создании объектов Kubernetes будет использоваться namespace `myapp-dev`. 
Имя namespace также по умолчанию состоит из [имени проекта]({{ site.baseurl }}/documentation/configuration/introduction.html#meta-configuration-doc) `myapp` (указанного в `werf.yaml`) и переданного названия окружения — `dev`. 
Более подробно про формирование namespace в Kubernetes можно посмотреть в [документации]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#kubernetes-namespace).

## Проверка работы приложения

Самое время узнать ip-адрес кластера. Если используется minikube, то узнать ip-адрес можно следующим способом (обычно это `192.168.99.100`):
```shell
minikube ip
```

Убедитесь, что имя `myapp.local` разрешается в полученный IP-адрес кластера на вашей машине. Например, добавьте соответствующую запись в файл `/etc/hosts`:

```shell
192.168.99.100 myapp.local
```

Проверим работу приложения, открыв адрес `http://myapp.local`.

## Удаление приложение из кластера

Для полного удаления из кластера развернутого приложения запустим следующую команду:

```shell
werf dismiss --env dev --with-namespace
```

## Читайте также

Более подробно об особенностях и возможностях деплоя приложений с помощью werf, например, об использовании секретов [читайте в руководстве]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html).
