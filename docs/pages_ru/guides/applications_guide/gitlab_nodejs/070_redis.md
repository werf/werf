---
title: Подключаем redis
sidebar: applications_guide
guide_code: gitlab_nodejs
permalink: documentation/guides/applications_guide/gitlab_nodejs/070_redis.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/requirements.yaml
- .helm/values.yaml
- .helm/secret-values.yaml
- src/js/index.js
- .gitlab-ci.yml
{% endfilesused %}

В этой главе мы настроим в нашем базовом приложении работу с простейшей in-memory базой данных, например, redis или memcached. Для примера возьмём первый вариант — это означает, что база данных будет stateless.

{% offtopic title="А как быть, если база данных должна сохранять данные на диске?" %}
Этот вопрос мы разберём в следующей главе на примере [PostgreSQL](080_database.html). В рамках текущей главы разберёмся с общими вопросами: как базу данных в принципе завести в кластер, сконфигурировать и подключиться к ней из приложения.
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

{% snippetcut name=".helm/requirements.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/070-redis/.helm/requirements.yaml" %}
{% raw %}
```yaml
dependencies:
<...>
- name: redis
  version: 9.3.2
  repository: https://kubernetes-charts.storage.googleapis.com/
  condition: redis.enabled
```
{% endraw %}
{% endsnippetcut %}

Для того чтобы werf при деплое загрузил необходимые нам сабчарты - нужно прописать в `.gitlab-ci.yml` работу с зависимостями:

{% snippetcut name=".gitlab-ci.yml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/070-redis/.gitlab-ci.yml" %}
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

А также сконфигурировать имя сервиса, порт, логин и пароль, согласно [документации](https://github.com/bitnami/charts/tree/master/bitnami/redis/#parameters) нашего сабчарта:

{% snippetcut name=".helm/values.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/070-redis/.helm/values.yaml" %}
{% raw %}
```yaml
redis:
  fullnameOverride: guided-redis
  nameOverride: guided-redis
```
{% endraw %}
{% endsnippetcut %}


{% offtopic title="А ключ redis он откуда такой?" %}
Этот ключ должен совпадать с именем сабчарта-зависимости в файле `requirements.yaml` — тогда настройки будут пробрасываться в сабчарт.
{% endofftopic %}
{% snippetcut name="secret-values.yaml (расшифрованный)" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/070-redis/.helm/secret-values.yaml" %}
{% raw %}
```yaml
redis:
  password: "LYcj6c09D9M4htgGh64vXLxn95P4Wt"
```
{% endraw %}
{% endsnippetcut %}

Сконфигурировать логин и порт для подключения у этого сабчарта невозможно, но если изучить исходный код — можно найти использующиеся в сабчарте значения. Пропишем нужные значения с понятными нам ключами — они понадобятся нам позже, когда мы будем конфигурировать приложение.

{% snippetcut name=".helm/values.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/070-redis/.helm/values.yaml" %}
{% raw %}
```yaml
redis:
   _port:
      _default: 6379
```
{% endraw %}
{% endsnippetcut %}

{% offtopic title="Почему мы пишем эти ключи со знака _ и вообще легально ли это?" %}
Когда мы пишем дополнительные ключи по соседству с ключами, пробрасывающимися в сабчарт, мы рискуем случайно "зацепить" лишнее. Поэтому нужно быть внимательным, сверяться с [документацией сабчарта](https://github.com/bitnami/charts/tree/master/bitnami/redis/#parameters) и не использовать пересекающиеся ключи.

Для надёжности — введём соглашение на использование знака подчёркивания `_` в начале таких ключей.
{% endofftopic %}


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

Знание этих Service нужно нам, чтобы потом к ним подключаться.

## Подключение NodeJS приложения к базе Redis

В нашем приложении - мы будем подключаться к master узлу Redis. Нам нужно, чтобы при выкате в любое окружение приложение подключалось к правильному Redis.

В нашем приложении мы будем использовать Redis как хранилище сессий.

{% snippetcut name="backend/src/server/server.js" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/070-redis/backend/src/server/server.js" %}
{% raw %}
```js
const REDIS_URI = "redis://"+ process.env.REDIS_HOST+":"+ process.env.REDIS_PORT || "redis://127.0.0.1:6379";
const SESSION_TTL = process.env.SESSION_TTL || 3600;
const COOKIE_SECRET = process.env.COOKIE_SECRET || "supersecret";
// Redis connect
const expSession = require("express-session");
const redis = require("redis");
let redisClient = redis.createClient(REDIS_URI);
let redisStore = require("connect-redis")(expSession);

var session = expSession({
  store: new redisStore({ client: redisClient, ttl: SESSION_TTL }),
  secret: "keyboard cat",
  resave: false,
  saveUninitialized: false,
});
var sharedsession = require("express-socket.io-session");
app.use(session);
```
{% endraw %}
{% endsnippetcut %}

Для подключения к базе данных нам, очевидно, нужно знать: хост, порт, логин, пароль. В коде приложения мы используем несколько переменных окружения: `REDIS_HOST`, `REDIS_PORT`, `REDIS_LOGIN`, `REDIS_PASSWORD`,`SESSION_TTL`, `COOKIE_SECRET`. Мы уже сконфигурировали часть значений в `values.yaml` для подключаемого сабчарта. Можно воспользоваться теми же значениями и дополнить их.

Будем **конфигурировать хост** через `values.yaml`:

{% snippetcut name=".helm/templates/deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/070-redis/.helm/templates/deployment.yaml" %}
{% raw %}
```yaml
- name: REDIS_HOST
  value: "{{ pluck .Values.global.env .Values.redis.host | first | default .Values.redis.host._default | quote }}"
```
{% endraw %}
{% endsnippetcut %}

{% snippetcut name=".helm/values.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/070-redis/.helm/values.yaml" %}
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

{% snippetcut name=".helm/templates/deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/070-redis/.helm/templates/deployment.yaml" %}
{% raw %}
```yaml
- name: REDIS_HOST
  value: "{{ .Chart.Name }}-{{ .Values.global.env }}-redis-master"
```
{% endraw %}
{% endsnippetcut %}

На практике иногда возникает необходимость переехать в другую базу данных или кастомизировать что-то — и в этих случаях в разы удобнее работать через `values.yaml`. Причём значений для разных окружений мы не прописываем, а ограничиваемся дефолтным значением:

{% snippetcut name="values.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/070-redis/.helm/values.yaml" %}
{% raw %}
```yaml
redis:
   host:
      _default: guided-redis-master
```
{% endraw %}
{% endsnippetcut %}

И под конкретные окружения значения прописываем только если это действительно нужно.
{% endofftopic %}

**Конфигурируем логин и порт** через `values.yaml`, просто прописывая значения:

{% snippetcut name=".helm/templates/deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/070-redis/.helm/templates/deployment.yaml" %}
{% raw %}
```yaml
- name: REDIS_PORT
  value: "{{ pluck .Values.global.env .Values.redis._port | first | default .Values.redis._port._default | quote }}"
```
{% endraw %}
{% endsnippetcut %}

Мы уже **сконфигурировали пароль** — используем прописанное ранее значение:

{% snippetcut name=".helm/templates/deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/070-redis/.helm/templates/deployment.yaml" %}
{% raw %}
```yaml
- name: REDIS_PASSWORD
  value: "{{ .Values.redis.password | quote }}"
```
{% endraw %}
{% endsnippetcut %}

Также нам нужно **сконфигурировать переменные, необходимые приложению** для работы с Redis:

{% snippetcut name=".helm/templates/deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/070-redis/.helm/templates/deployment.yaml" %}
{% raw %}
```yaml
- name: SESSION_TTL
  value: "{{ pluck .Values.global.env .Values.app.redis.session_ttl | first | default .Values.redis.session_ttl_default | quote }}"
- name: COOKIE_SECRET
  value: "{{ pluck .Values.global.env .Values.app.redis.cookie_secret | first | default .Values.app.redis.cookie_secret_default | quote }}"
```
{% endraw %}
{% endsnippetcut %}


{% snippetcut name="values.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/070-redis/.helm/values.yaml" %}
{% raw %}
```yaml
  redis:
    session_ttl:
        _default: "3600"
    cookie_secret:
        _default: "supersecret"
```
{% endraw %}
{% endsnippetcut %}

<div>
    <a href="080_database.html" class="nav-btn">Далее: Подключение базы данных</a>
</div>
