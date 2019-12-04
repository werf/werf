---
title: Деплой в Kubernetes
sidebar: documentation
permalink: documentation/guides/deploy_into_kubernetes.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

## Обзор задачи

Будет рассмотрено как деплоить приложение в Kubernetes с помощью Werf.

Для деплоя приложений в Kubernetes Werf использует [Helm](https://helm.sh) (с некоторыми изменениями и дополнениями). В статье мы создадим простое web-приложение, соберем все необходимые для него образы, создадим Helm-шаблоны и запустим приложение в кластере Kubernetes.

## Требования

 * Работающий кластер Kubernetes. Для выполнения примера вы можете использовать как обычный Kubernetes кластер, так и Minikube. Если вы решили использовать Minikube, прочитайте [статью о настройке Minikube]({{ site.baseurl }}/documentation/reference/development_and_debug/setup_minikube.html), чтобы запустить Minikube и Docker Registry.
 * Работающий Docker Registry.
   * Доступ от хостов Kubernetes с правами на push образов в Docker Registry.
   * Доступ от хостов Kubernetes с правами на pull образов в Docker Registry.
 * Установленные [зависимости Werf]({{ site.baseurl }}/documentation/guides/installation.html#install-dependencies).
 * Установленный [Multiwerf](https://github.com/flant/multiwerf).
 * Установленный и сконфигурированный `kubectl` для доступа в кластер Kubernetes (<https://kubernetes.io/docs/tasks/tools/install-kubectl/>).

**Внимание!** Далее в качестве адреса репозитория будет использоваться значение `:minikube` . Если вы используете собственный кластер Kubernetes и Docker Registry, то указывайте репозиторий проекта в Docker Registry вместо аргумента `:minikube`.

### Выбор версии Werf

Перед началом работы необходимо выбрать версию Werf. Для выбора актуальной версии Werf в канале beta, релиза 1.0, выполним следующую команду:

```shell
source <(multiwerf use 1.0 beta)
```

## Архитектура приложения

Пример представляет собой простейшее web-приложение, для запуска которого нам нужен только web-сервер.

Архитектура подобных приложений в Kubernetes выглядит, как правило, следующим образом:

     .----------------------.
     | backend (Deployment) |
     '----------------------'
                |
                |
      .--------------------.
      | frontend (Ingress) |
      '--------------------'

Здесь `backend` — web-сервер с приложением, `frontend` — прокси-сервер, который выступает точкой входа и перенаправления внешнего трафика в приложение.

## Файлы приложения

Werf ожидает, что все файлы, необходимые для сборки и развертывания приложения, находятся в папке приложения (папке проекта) вместе с исходным кодом, если он имеется.   

Создадим пустую директорию проекта и перейдём в неё для выполнения следующих шагов:

```shell
mkdir myapp
cd myapp
```

## Подготовка образа

Необходимо подготовить образ приложения с web-сервером внутри. Для этого создадим файл `werf.yaml` в папке приложения со следующим содержимым:

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

Web-приложение состоит из единственной статической HTML-страницы, которая задаётся в инструкциях и создаётся при сборке образа. Содержимое этой страницы будет отдавать Python HTTP-сервер.

Соберём образ приложения и загрузим его в Docker Registry:

```shell
werf build-and-publish --stages-storage :local --tag-custom myapp --images-repo :minikube
```

Название собранного образа приложения состоит из адреса Docker Registry (`REPO`) и тэга (`TAG`). 
При указании `:minikube` в качестве адреса Docker Registry Werf использует адрес `werf-registry.kube-system.svc.cluster.local:5000/myapp`. Так как в качестве тега был указан `myapp`, Werf загрузит в Docker Registry образ `werf-registry.kube-system.svc.cluster.local:5000/myapp:myapp`.

## Подготовка конфигурации деплоя

Werf использует встроенный [Helm](helm.sh) *для применения* конфигурации в Kubernetes. 
Для *описания* объектов Kubernetes Werf использует конфигурационные файлы Helm: шаблоны и файлы с параметрами (например, `values.yaml`). 
Помимо этого, Werf поддерживает дополнительные файлы, такие как — файлы секретами и с секретными значениями (например `secret-values.yaml`), а также дополнительные Go-шаблоны для интеграции собранных образов.

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

В конфигурации описывается создание Deployment `myapp-backend` (конструкция {% raw %}`{{ .Chart.Name }}-backend`{% endraw %} будет преобразована в `myapp-backend`) с четырьмя репликами.

Конструкция {% raw %}`{{ werf_container_image . | indent 8 }}`{% endraw %} — это использование функции Werf, которая:
* всегда возвращает поле `image:` для ресурса Kubernetes с именем образа, учитывая используемую схему тэгирования (в нашем случае — `werf-registry.kube-system.svc.cluster.local:5000/myapp:latest`)
* функция может возвращать дополнительные поля, такие как `imagePullPolicy`, на основании схемы тэгирования, заложенной логики и некоторых внешних условий.

Функция `werf_container_image` позволяет удобно указывать имя образа в объекте Kubernetes исходя из **описанной** (в `werf.yaml`) конфигурации. Как использовать эту функцию в случае если в конфигурации описано несколько образов, читай подробнее [в соответствующей статье]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_image).

Конструкция {% raw %}`{{ werf_container_env . | indent 8 }}`{% endraw %} — это использование другой внутренней функции Werf, которая *может* возвращать содержимое секции `env:` соответствующего контейнера объекта Kubernetes. Использование функции `werf_container_image` при описании объекта приводит к тому, что Kubernetes будет перезапускать pod'ы Deployment'а только если соответствующий образ изменился (смотри [подробнее]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_env)).

Наконец, создается сервис `myapp-backend`, для доступа к pod'ам Deployment'а `myapp-backend`.

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

Эта конфигурация описывает Ingress, и настраивает NGINX прокси-сервер перенаправлять трафик для хоста `myapp.local` на backend-сервер `myapp-backend`.

## Деплой

Если используется `minikube`, то перед деплоем необходимо включить ingress-модуль:

```shell
minikube addons enable ingress
```

Наконец, запустим деплой:

```shell
werf deploy --stages-storage :local --images-repo :minikube --tag-custom myapp --env dev
```

После запуска команды, werf создаст соответствующие ресурсы в Kubernetes и будет отслеживать статус Deployment `myapp-backend` до его готовности (готовности всех Pod'ов) либо ошибки.

Для того, чтобы сформировать правильные имя Helm-релиза и namespace, требуется указать [окружение]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#environment) с помощью параметра `--env`.

В результате будет создан helm-релиз с именем `myapp-dev`. Название Helm-релиза состоит из [имени проекта]({{ site.baseurl }}/documentation/configuration/introduction.html#meta-configuration-doc) `myapp` (указанного в `werf.yaml`) и переданного названия окружения — `dev`. Более подробно про формирование имен Helm-релизов можно посмотреть в [документации]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#release-name).

При создании объектов Kubernetes будет использоваться namespace `myapp-dev`. Имя этого namespace также состоит из [имени проекта]({{ site.baseurl }}/documentation/configuration/introduction.html#meta-configuration-doc) `myapp` (указанного в `werf.yaml`) и переданного названия окружения — `dev`. Более подробно про формирование namespace в Kubernetes можно посмотреть в [документации]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#kubernetes-namespace).

## Проверка работы приложения

Самое время узнать ip-адрес вашего кластера. Если вы используете minikube, то узнать ip-адрес можно следующим способом (обычно это `192.168.99.100`):
```shell
minikube ip
```

Убедитесь, что имя `myapp.local` разрешается в полученный IP-адрес вашего кластера на вашей машине. Например, добавьте соответствующую запись в файл `/etc/hosts`:

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

Более подробно об особенностях и возможностях деплоя приложений с помощью Werf, например, об использовании секретов [читайте в руководстве]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html).
