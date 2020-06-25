---
title: Подключаем redis
sidebar: applications-guide
permalink: documentation/guides/applications-guide/template/070-redis.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/requirements.yaml
- .helm/values.yaml
- config/cable.yml
- .gitlab-ci.yml
{% endfilesused %}

В этой главе мы настроим в нашем базовом приложении работу простейшей in-memory базой данных, например, redis или memcached. Для примера возьмём первый вариант.

{% offtopic title="А как быть, если база данных должна сохранять данные на диске?" %}
Этот вопрос мы разберём в следующей главе на примере [PostgreSQL](080-database.html). В рамках текущей главы разберёмся с общими вопросами: как базу данных в принципе завести в кластер, сконфигурировать и подключиться к ней из приложения.
{% endofftopic %}


В простейшем случае нет необходимости вносить изменения в сборку — уже собранные образы есть на DockerHub. Надо просто выбрать правильный образ, корректно сконфигурировать его в своей инфраструктуре, а потом подключиться к базе данных из Rails приложения.

## Сконфигурировать Redis в Kubernetes

Есть два способа подключить: прописать helm-чарт самостоятельно или подключить внешний чарт. Мы рассмотрим второй вариант, подключим redis как внешний subchart.

Todo
нам надо
- Имя сервиса
- Имя базы
- Пароль
- ...

Для этого нужно:

1. прописать изменения в yaml файлы;
2. указать Redis конфиги;
3. подсказать werf, что ему нужно подтягивать subchart.

Пропишем helm-зависимости:

{% snippetcut name=".helm/requirements.yaml" url="template-files/examples/example_4/.helm/requirements.yaml" %}
```yaml
dependencies:
- name: redis
  version: 9.3.2
  repository: https://kubernetes-charts.storage.googleapis.com/
  condition: redis.enabled
```
{% endsnippetcut %}

Для того чтобы werf при деплое загрузил необходимые нам сабчарты - нужно прописать в `.gitlab-ci.yml` работу с зависимостями

{% snippetcut name=".gitlab-ci.yml" url="template-files/examples/example_4/.gitlab-ci.yml#L24" %}
```yaml
.base_deploy:
  stage: deploy
  script:
    - werf helm repo init
    - werf helm dependency update
    - werf deploy
```
{% endsnippetcut %}

Для того, чтобы подключённые сабчарты заработали — нужно указать настройки в `values.yaml`:

{% snippetcut name=".helm/values.yaml" url="template-files/examples/example_4/.helm/values.yaml#L3" %}
```yaml
redis:
  enabled: true
```
{% endsnippetcut %}

При использовании сабчарта по умолчанию создается master-slave кластер redis.

Если посмотреть на рендер (`werf helm render`) нашего приложения с включенным сабчартом для redis, то можем увидеть какие будут созданы объекты Service:

```yaml
# Source: example_4/charts/redis/templates/redis-master-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: guided-redis-master

# Source: example_4/charts/redis/templates/redis-slave-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: guided-redis-slave
```

Знание этих Service нужно нам, чтобы потом к ним подключаться.

## Подключение Rails приложения к базе Redis

В нашем приложении - мы будем подключаться к master узлу Redis. Нам нужно, чтобы при выкате в любое окружение приложение подключалось к правильному Redis.

В нашем приложении мы будем использовать Redis как хранилище сессий.

{% snippetcut name="____________" url="____________" %}
```____________
____________
```
{% endsnippetcut %}

Для подключения к базе данных нам, очевидно, нужно знать: хост, порт, логин, пароль. В коде приложения мы используем несколько переменных окружения: `REDIS_HOST`, `REDIS_PORT`, `REDIS_LOGIN`, `REDIS_PASSWORD`.  

Будем **конфигурировать хост** через `values.yaml`:

{% snippetcut name=".helm/templates/deployment.yaml" url="____________" %}
```yaml
- name: REDIS_HOST
  value: "{{ pluck .Values.global.env .Values.redis.host | first | default .Values.redis.host_default | quote }}"
```
{% endsnippetcut %}

{% offtopic title="А зачем такие сложности, может просто прописать значения в шаблоне?" %}

Казалось бы, можно написать примерно так:

```yaml
- name: REDIS_HOST
  value: "{{ .Chart.Name }}-{{ .Values.global.env }}-redis-master"
```

На практике иногда возникает необходимость переехать в другую базу данных или кастомизировать что-то — и в этих случаях в разы удобнее работать через `values.yaml`. Причём значений для разных окружений мы не прописываем, а ограничиваемся дефолтным значением:

```yaml 
redis:
   host:
      _default: redis
```

И под конкретные окружения значения прописываем только если это действительно нужно.
{% endofftopic %}

**Конфигурируем логин и порт** через `values.yaml`, просто прописывая значения:

{% snippetcut name=".helm/templates/deployment.yaml" url="____________" %}
```yaml
- name: REDIS_LOGIN
  value: "{{ pluck .Values.global.env .Values.redis.login | first | default .Values.redis.login_default | quote }}"
- name: REDIS_PORT
  value: "{{ pluck .Values.global.env .Values.redis.port | first | default .Values.redis.port_default | quote }}"
```
{% endsnippetcut %}

{% snippetcut name="values.yaml" url="____________" %}
```yaml
redis:
   login:
      _default: ____________
   port:
      _default: ____________
```
{% endsnippetcut %}

TODO: Конфигурируем пароль ХУЙ ЗНАЕТ КАК ВООБЩЕ

{% snippetcut name=".helm/templates/deployment.yaml" url="____________" %}
```yaml
- name: REDIS_PASSWORD
  value: "{{ pluck .Values.global.env .Values.redis.password | first | default .Values.redis.password_default | quote }}"
```
{% endsnippetcut %}

{% snippetcut name="secret-values.yaml" url="____________" %}
```yaml
redis:
  password:
    _default: 100067e35229a23c5070ad5407b7406a7d58d4e54ecfa7b58a1072bc6c34cd5d443e
```
{% endsnippetcut %}

<div>
    <a href="080-database.html" class="nav-btn">Далее: Подключение базы данных</a>
</div>
