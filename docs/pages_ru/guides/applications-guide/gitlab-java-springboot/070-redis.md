---
title: Подключаем redis
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-java-springboot/070-redis.html
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

В этой главе мы настроим в нашем базовом приложении работу простейшей in-memory базой данных, например, redis или memcached. Для примера возьмём первый вариант — это означает, что база данных будет stateless.

{% offtopic title="А как быть, если база данных должна сохранять данные на диске?" %}
Этот вопрос мы разберём в следующей главе на примере [PostgreSQL](080-database.html). В рамках текущей главы разберёмся с общими вопросами: как базу данных в принципе завести в кластер, сконфигурировать и подключиться к ней из приложения.
{% endofftopic %}

В простейшем случае нет необходимости вносить изменения в сборку — уже собранные образы есть на DockerHub. Надо просто выбрать правильный образ, корректно сконфигурировать его в своей инфраструктуре, а потом подключиться к базе данных из Rails приложения.

## Сконфигурировать Redis в Kubernetes

Для того, чтобы сконфигурировать Redis в кластере — необходимо прописать объекты с помощью Helm. Мы можем сделать это самостоятельно, но рассмотрим вариант с подключением внешнего чарта. В любом случае, нам нужно будет указать: имя сервиса, порт, логин и пароль — и разобраться, как эти параметры пробросить в подключённый внешний чарт. 

Нам необходимо будет:

1. Указать Redis как зависимый сабчарт в `requirements.yaml`;
2. Сконфигурировать в werf работу с зависимостями;
3. Сконфигурировать подключённый сабчарт;
4. Убедиться, что создаётся master-slave кластер Redis.

Пропишем сабчарт с Redis:

{% snippetcut name=".helm/requirements.yaml" url="#" %}
{% raw %}
```yaml
dependencies:
- name: redis
  version: 9.3.2
  repository: https://kubernetes-charts.storage.googleapis.com/
  condition: redis.enabled
```
{% endraw %}
{% endsnippetcut %}

Для того чтобы werf при деплое загрузил необходимые нам сабчарты - нужно прописать в `.gitlab-ci.yml` работу с зависимостями:

{% snippetcut name=".gitlab-ci.yml" url="#" %}
{% raw %}
```yaml
.base_deploy:
  stage: deploy
  script:
    - werf helm repo init
    - werf helm dependency update
    - werf deploy
```
{% endraw %}
{% endsnippetcut %}

Для того, чтобы подключённые сабчарты заработали — нужно указать настройки в `values.yaml`:

{% snippetcut name=".helm/values.yaml" url="#" %}
{% raw %}
```yaml
redis:
  enabled: true
```
{% endraw %}
{% endsnippetcut %}

А также сконфигурировать имя сервиса, порт, логин и пароль, согласно [документации](https://github.com/bitnami/charts/tree/master/bitnami/redis/#parameters) нашего сабчарта:

{% snippetcut name=".helm/values.yaml" url="#" %}
{% raw %}
```yaml
redis:
  fullnameOverride: guided-redis
  nameOverride: guided-redis
```
{% endraw %}
{% endsnippetcut %}

{% snippetcut name="secret-values.yaml (расшифрованный)" url="#" %}
{% raw %}
```yaml
redis:
  password: "LYcj6c09D9M4htgGh64vXLxn95P4Wt"
```
{% endraw %}
{% endsnippetcut %}

Сконфигурировать логин и порт для подключения у этого сабчарта невозможно, но если изучить исходный код — можно найти использующиеся в сабчарте значения. Пропишем нужные значения с понятными нам ключами — они понадобятся нам позже, когда мы будем конфигурировать приложение.

{% snippetcut name=".helm/values.yaml" url="#" %}
{% raw %}
```yaml
redis:
   login:
      _default: ____________
   port:
      _default: 6379
```
{% endraw %}
{% endsnippetcut %}


{% offtopic title="Как быть, если найти параметры не получается?" %}

Некоторые сервисы вообще не требуют аутентификации, в частности, Redis зачастую используется без неё.

{% endofftopic %}

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

## Подключение Rails приложения к базе Redis

В нашем приложении - мы будем подключаться к master узлу Redis. Нам нужно, чтобы при выкате в любое окружение приложение подключалось к правильному Redis.

В нашем приложении мы будем использовать Redis как хранилище сессий, указываем нужные dependency:

{% snippetcut name="pom.xml" url="#" %}
{% raw %}
```xml
    <dependency>
      <groupId>org.springframework.boot</groupId>
      <artifactId>spring-boot-starter-data-redis-reactive</artifactId>
    </dependency>
```
{% endraw %}
{% endsnippetcut %}

Сопоставим переменные java, используемые для подключения к redis и переменные окружения контейнера. 
Как и в случае с работой с файлами выше пропишем в application.properties:

TODO: вот это не правда, надо испльзовать те переименные, что прописаны в стсатье

{% snippetcut name="application.properties" url="#" %}
{% raw %}
```yaml
spring.redis.host=${REDISHOST}
spring.redis.port=${REDISPORT}
```
{% endraw %}
{% endsnippetcut %}

Для подключения к базе данных нам, очевидно, нужно знать: хост, порт, логин, пароль. В коде приложения мы используем несколько переменных окружения: `REDIS_HOST`, `REDIS_PORT`, `REDIS_LOGIN`, `REDIS_PASSWORD`. Мы уже сконфигурировали часть значений в `values.yaml` для подключаемого сабчарта. Можно воспользоваться теми же значениями и дополнить их.

Будем **конфигурировать хост** через `values.yaml`:

{% snippetcut name=".helm/templates/deployment.yaml" url="#" %}
{% raw %}
```yaml
- name: REDIS_HOST
  value: "{{ pluck .Values.global.env .Values.redis.host | first | default .Values.redis.host_default | quote }}"
```
{% endraw %}
{% endsnippetcut %}

{% snippetcut name=".helm/values.yaml" url="#" %}
{% raw %}
```yaml
redis:
  host:
    _default: guided-redis-master
```
{% endraw %}
{% endsnippetcut %}

{% offtopic title="А зачем такие сложности, может просто прописать значения в шаблоне?" %}

Казалось бы, можно написать примерно так:

{% snippetcut name=".helm/templates/deployment.yaml" url="#" %}
{% raw %}
```yaml
- name: REDIS_HOST
  value: "{{ .Chart.Name }}-{{ .Values.global.env }}-redis-master"
```
{% endraw %}
{% endsnippetcut %}

На практике иногда возникает необходимость переехать в другую базу данных или кастомизировать что-то — и в этих случаях в разы удобнее работать через `values.yaml`. Причём значений для разных окружений мы не прописываем, а ограничиваемся дефолтным значением:

{% snippetcut name="values.yaml" url="#" %}
{% raw %}
```yaml 
redis:
   host:
      _default: redis
```
{% endraw %}
{% endsnippetcut %}

И под конкретные окружения значения прописываем только если это действительно нужно.
{% endofftopic %}

**Конфигурируем логин и порт** через `values.yaml`, просто прописывая значения:

{% snippetcut name=".helm/templates/deployment.yaml" url="#" %}
{% raw %}
```yaml
- name: REDIS_LOGIN
  value: "{{ pluck .Values.global.env .Values.redis.login | first | default .Values.redis.login_default | quote }}"
- name: REDIS_PORT
  value: "{{ pluck .Values.global.env .Values.redis.port | first | default .Values.redis.port_default | quote }}"
```
{% endraw %}
{% endsnippetcut %}

Мы уже **сконфигурировали пароль** — используем прописанное ранее значение:

{% snippetcut name=".helm/templates/deployment.yaml" url="#" %}
{% raw %}
```yaml
- name: REDIS_PASSWORD
  value: "{{ .Values.redis.password | quote }}"
```
{% endraw %}
{% endsnippetcut %}

<div>
    <a href="080-database.html" class="nav-btn">Далее: Подключение базы данных</a>
</div>
