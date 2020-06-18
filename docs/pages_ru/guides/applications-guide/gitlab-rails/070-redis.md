---
title: Подключаем redis
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-rails/070-redis.html
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

TODO: чёт я нихуя не понял, а что с хранением данных?

В этой главе мы настроим в нашем базовом приложении работу простейшей базой данных, например, redis или memcached. Для примера возьмём первый вариант.

В простейшем случае нет необходимости вносить изменения в сборку — уже собранные образы есть на DockerHub. Надо просто выбрать правильный образ, корректно сконфигурировать его в своей инфраструктуре, а потом подключиться к базе данных из Rails приложения.

## Сконфигурировать Redis в Kubernetes

Есть два способа подключить: прописать helm-чарт самостоятельно или подключить внешний чарт. Мы рассмотрим второй вариант, подключим redis как внешний subchart.

Для этого нужно:

1. прописать изменения в yaml файлы;
2. указать Redis конфиги;
3. подсказать werf, что ему нужно подтягивать subchart.

Пропишем helm-зависимости:

{% snippetcut name=".helm/requirements.yaml" url="gitlab-rails-files/examples/example_4/.helm/requirements.yaml" %}
```yaml
dependencies:
- name: redis
  version: 9.3.2
  repository: https://kubernetes-charts.storage.googleapis.com/
  condition: redis.enabled
```
{% endsnippetcut %}

Для того чтобы werf при деплое загрузил необходимые нам сабчарты - нужно прописать в `.gitlab-ci.yml` работу с зависимостями

{% snippetcut name=".gitlab-ci.yml" url="gitlab-rails-files/examples/example_4/.gitlab-ci.yml#L24" %}
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

{% snippetcut name=".helm/values.yaml" url="gitlab-rails-files/examples/example_4/.helm/values.yaml#L3" %}
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
  name: example-4-stage-redis-master

# Source: example_4/charts/redis/templates/redis-slave-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: example-4-stage-redis-slave
```

Знание этих Service нужно нам, чтобы потом к ним подключаться.

## Подключение Rails приложения к базе redis

В нашем приложении - мы будем подключаться к master узлу Redis. Нам нужно, чтобы при выкате в любое окружение приложение подключалось к правильному Redis. Рассмотрим настройки подключение к redis из нашего приложения на примере стандартного cable (`config/cable.yml`) и стандартного гема `gem 'redis', '~> 4.0'`.

{% snippetcut name="config/cable.yml" url="gitlab-rails-files/examples/example_4/config/cable.yml#L9" %}
```yaml
production:
  adapter: redis
  url: <%= ENV.fetch("REDIS_URL") { "redis://localhost:6379/1" } %>
  channel_prefix: example_2_production
```
{% endsnippetcut %}

В данном файле мы видим что адрес подключения берется из переменной окружения `REDIS_URL` и если такая переменная не задана - подставляется значение по умолчанию `redis://localhost:6379/1`. Пробросим в `REDIS_URL` реальное название Service:

{% snippetcut name=".helm/templates/deployment.yaml" url="gitlab-rails-files/examples/example_4/.helm/templates/deployment.yaml#L29" %}
```yaml
- name: REDIS_URL
  value: "redis://{{ .Chart.Name }}-{{ .Values.global.env }}-redis-master:6379/1"
```
{% endsnippetcut %}

Таким образом, для каждого стенда будет прописываться корректное значение `REDIS_URL`, например `redis://example-2-stage-redis-master:6379/1` — для stage окружения.

<div>
    <a href="080-database.html" class="nav-btn">Далее: Подключение базы данных</a>
</div>
