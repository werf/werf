---
title: Подключаем redis
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-python-django/070-redis.html
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

Допустим к нашему приложению нужно подключить простейшую базу данных, например, redis или memcached. Возьмем первый вариант.

В простейшем случае нет необходимости вносить изменения в сборку — всё уже собрано для нас. Надо просто подключить нужный образ, а потом в вашем django приложении корректно обратиться к этому приложению.

Все данные для подключения у нас будут передаваться через переменные окружения.

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
  name: example-2-stage-redis-master

# Source: example-2/charts/redis/templates/redis-slave-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: example-2-stage-redis-slave
```

## Подключение Django приложения к базе redis

В нашем приложении - мы будем  подключаться к мастер узлу редиса. Нам нужно, чтобы при выкате в любое окружение приложение подключалось к правильному редису.

Рассмотрим настройки подключения к redis из нашего приложения.

```yaml
CACHES = {
    "default": {
        "BACKEND": "django_redis.cache.RedisCache",
        "LOCATION": os.environ.get('REDIS_URL', default="redis://127.0.0.1:6379/1"),
        "OPTIONS": {
            "CLIENT_CLASS": "django_redis.client.DefaultClient"
        },
        "KEY_PREFIX": os.environ.get('REDIS_KEY_PREFIX', default="example")
    }
}
```

В данном файле мы видим что адрес подключения берется из переменной окружения `REDIS_URL` и если такая переменная не задана - подставляется значение по умолчанию `redis://127.0.0.1:6379/1`


Чтобы добавить redis в качестве кеша, нам необходимо установить пакет ``, и добавить в файл настроек `settings.py` следующие строки:

Для подключения нашего приложения к redis нам необходимо установить дополнительный пакет `pip install django-redis` и указать переменную окружения `REDIS_URL` при деплое нашего приложения в файле с описанием деплоймента.


```
- name: REDIS_URL
  value: "redis://{{ .Chart.Name }}-{{ .Values.global.env }}-redis-master:6379/1"
```


В итоге, при деплое нашего приложения преобразуется например в строку

`redis://example-2-stage-redis-master:6379/1` для stage окружения

Более подробно про подключение к редису можно почитать в [документации](https://github.com/jazzband/django-redis)

<div>
    <a href="080-database.html" class="nav-btn">Далее: Подключение базы данных</a>
</div>
