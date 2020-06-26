---
title: Подключаем redis
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-nodejs/070-redis.html
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

Для этого нужно:

1. Указать Redis как зависимый сабчарт в `requirements.yaml`;
2. Добавить в `gitlab-ci.yml` инициализацию и скачивание сабчартов с помощью werf;
3. Добавить в манифест с приложением переменные для подключения к Redis;

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

Однако не редко возникают ситуации когда нам необходима не совсем стандартная конфигурация сабчарта, либо же сильно видоизмененная. Все официальные чарты пользуются файлом `values.yaml` для своей настройки. Но до момента скачивания сабчарта мы не имеем к нему доступа.

Потому в werf существует механизм передачи параметров из нашего основного `values.yaml` непосредственно внутрь используемых нами сабчартов, позволяя нам хранить всю необходимую конфигурацию в одном месте.

Рассмотрим на примере нашего Redis сабчарта. Имена сервисов которые мы получили при рендере выглядят некорректно и длинно, т.к. сабчарт генерирует имена своих объектов на основании имени релиза.

Согласно [документации](https://github.com/bitnami/charts/tree/master/bitnami/redis/#parameters) нашего сабчарта, мы можем изменить генерацию имен по умолчанию добавив параметры `nameOverride` и `fullnameOverride`.

Потому добавляем эти параметры в наш `values.yaml`

{% snippetcut name=".gitlab-ci.yml" url="template-files/examples/example_4/.helm/values.yaml#L24" %}
```yaml
redis:
  fullnameOverride: redis
  nameOverride: redis
```
Первой строкой мы указали сабчарт в который мы передаем параметры (его имя должно совпадать с именем указанным в `requirements.yaml`), а затем и сами параметры точно так же как бы мы это делали в `values.yaml` самого сабчарта.

Теперь имена сервисов Redis после рендера будут выглядеть так:

```yaml
# Source: example_4/charts/redis/templates/redis-master-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: redis-master

# Source: example_4/charts/redis/templates/redis-slave-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: redis-slave
```

## Подключение NodeJS приложения к базе Redis

В нашем приложении - мы будем подключаться к master узлу Redis. Нам нужно, чтобы при выкате в любое окружение приложение подключалось к правильному Redis.

В нашем приложении мы будем использовать Redis как хранилище сессий.

{% snippetcut name="src/js/index.js" url="____________" %}
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
{% endsnippetcut %}

Для подключения к базе данных нам, очевидно, нужно знать: хост, порт, логин, пароль. В коде приложения мы используем несколько переменных окружения: `REDIS_HOST`, `REDIS_PORT`, `SESSION_TTL`, `COOKIE_SECRET`.  

Вы скорее всего заметили что мы не указали переменные для логина и пароля. В данном случае для подключения к редису они нам не требуются.

{% offtopic title="А почему они не требуются?" %}

На это у нас есть несколько причин:
1. К сожалению данный сабчарт не поддерживает использование отдельного логина.
2. В этом случае ради примера нам достаточно указать стандартную конфигурацию Redis без аутентификации, т.к. при установке из сабчарта он остается доступным только внутри кластера Кубернетес.

Тем не менее для Redis находящегося за пределами кластера, либо любой другой базы схожего типа (Memcached, Tarantool) вы можете прописать данные для аутентификации как обычные переменные.

{% endofftopic %}

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
  value: "redis-master"
```

На практике иногда возникает необходимость переехать в другую базу данных или кастомизировать что-то — и в этих случаях в разы удобнее работать через `values.yaml`. Причём значений для разных окружений мы не прописываем, а ограничиваемся дефолтным значением:

```yaml 
app:
  redis:
    host:
      _default: redis-master
```

И под конкретные окружения значения прописываем только если это действительно нужно.
{% endofftopic %}

**Конфигурируем порт и остальные параметры** через `values.yaml`, просто прописывая значения:

{% snippetcut name=".helm/templates/deployment.yaml" url="____________" %}
```yaml
- name: REDIS_PORT
  value: "{{ pluck .Values.global.env .Values.app.redis.port | first | default .Values.app.redis.port_default | quote }}"
- name: SESSION_TTL
  value: "{{ pluck .Values.global.env .Values.app.redis.session_ttl | first | default .Values.redis.session_ttl_default | quote }}"
- name: COOKIE_SECRET
  value: "{{ pluck .Values.global.env .Values.app.redis.cookie_secret | first | default .Values.app.redis.cookie_secret_default | quote }}"
```
{% endsnippetcut %}

{% snippetcut name="values.yaml" url="____________" %}
```yaml
app:
  redis:
    port:
        _default: ____________
    session_ttl:
        _default: ____________
    cookie_secret:
        _default: ____________
```
{% endsnippetcut %}

<div>
    <a href="080-database.html" class="nav-btn">Далее: Подключение базы данных</a>
</div>
