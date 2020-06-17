---
title: Подключаем redis
sidebar: applications-guide
permalink: documentation/guides/applications-guide/template/070-redis.html
layout: guide
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/requirements.yaml
- .helm/values.yaml
- config/cable.yml
- .gitlab-ci.yml
{% endfilesused %}

Допустим к нашему приложению нужно подключить простейшую базу данных, например, redis или memcached. Возьмем первый вариант.

В простейшем случае нет необходимости вносить изменения в сборку — всё уже собрано для нас. Надо просто подключить нужный образ, а потом в вашем Nodejs приложении корректно обратиться к этому приложению.

<a name="redis-to-kubernetes" />

## Завести Redis в Kubernetes

Есть два способа подключить: прописать helm-чарт самостоятельно или подключить внешний. Мы рассмотрим второй вариант.

Подключим redis как внешний subchart.

Для этого нужно:

1. прописать изменения в yaml файлы;
2. указать редису конфиги
3. подсказать werf, что ему нужно подтягивать subchart.

Добавим в файл `.helm/requirements.yaml` следующие изменения:

```yaml
dependencies:
- name: redis
  version: 9.3.2
  repository: https://kubernetes-charts.storage.googleapis.com/
  condition: redis.enabled
```

Для того чтобы werf при деплое загрузил необходимые нам сабчарты - нужно добавить команды в `.gitlab-ci`

```yaml
.base_deploy:
  stage: deploy
  script:
    - werf helm repo init
    - werf helm dependency update
    - werf deploy
```

Опишем параметры для redis в файле `.helm/values.yaml`

```yaml
redis:
  enabled: true
```

При использовании сабчарта по умолчанию создается master-slave кластер redis.

Если посмотреть на рендер (`werf helm render`) нашего приложения с включенным сабчартом для redis, то можем увидеть какие будут созданы сервисы:

```yaml
# Source: example-2/charts/redis/templates/redis-master-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: chat-stage-redis-master

# Source: example-2/charts/redis/templates/redis-slave-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: chat-stage-redis-slave
```

<a name="redis-to-app" />

## Подключение приложения к Redis

В нашем приложении - мы будем  подключаться к мастер узлу редиса. Нам нужно, чтобы при выкате в любое окружение приложение подключалось к правильному редису.

В src/js/index.js мы добавляем:


```js
const REDIS_URI = process.env.SESSION_REDIS || "redis://127.0.0.1:6379";
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

Добавляем в values.yaml

values.yaml

```yaml
app:
  redis:
     host:
       master:
         stage: chat-stage-redis-master
       slave:
         stage: chat-stage-redis-slave
```


secret-values.yaml


```yaml
app:
 redis:
    password:
      _default: 100067e35229a23c5070ad5407b7406a7d58d4e54ecfa7b58a1072bc6c34cd5d443e
```


И наконец добавляем подстановку переменных в сам манифест с нашим приложением: 



```yaml
      - name: SESSION_REDIS
        value: "redis://root:{{ pluck .Values.global.env .Values.app.redis.password | first | default .Values.app.redis.password._default }}@{{ pluck .Values.global.env .Values.app.redis.host.master | first }}:6379"
```



В данном случае Redis подключается как хранилище для сессий.

