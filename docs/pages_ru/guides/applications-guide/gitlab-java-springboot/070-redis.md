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

В простейшем случае нет необходимости вносить изменения в сборку — всё уже собрано для нас. Надо просто подключить нужный образ, а потом в вашем Java-приложении корректно обратиться к этому приложению.

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
  name: example-2-stage-redis-master

# Source: example-2/charts/redis/templates/redis-slave-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: example-2-stage-redis-slave
```

<a name="redis-to-app" />

## Подключение приложения к Redis

В нашем приложении - мы будем  подключаться к мастер узлу редиса. Нам нужно, чтобы при выкате в любое окружение приложение подключалось к правильному редису.

Рассмотрим настройки подключения к redis из нашего приложения.

Чтобы начать использовать redis - указаваем в pom.xml нужные denendency для работы с redis

```xml
    <dependency>
      <groupId>org.springframework.boot</groupId>
      <artifactId>spring-boot-starter-data-redis-reactive</artifactId>
    </dependency>
```

[pom.xml](gitlab-java-springboot-files/03-demo-db/pom.xml:32-35)

[werf.yaml](gitlab-java-springboot-files/03-demo-db/werf.yaml) остается неизменным - будем пользоваться тем, что получили в главе оптимизация сборки - чтобы не отвлекаться на сборку ассетов.
Сопоставим переменные java, используемые для подключения к redis и переменные окружения контейнера. 
Как и в случае с работой с файлами выше пропишем в application.properties:

```yaml
spring.redis.host=${REDISHOST}
spring.redis.port=${REDISPORT}
```

[application.properties](gitlab-java-springboot-files/03-demo-db/src/main/resources/application.properties:1-2)

и пеередаем в секции env deployment-а:

```yaml
       env:
{{ tuple "hello" . | include "werf_container_env" | indent 8 }}
        - name: REDISHOST
          value: {{ pluck .Values.global.env .Values.redis.host | first | default .Values.redis.host_default | quote }}
        - name: REDISPORT
          value: {{ pluck .Values.global.env .Values.redis.port | first | default .Values.redis.port_default | quote }}
```

[deployment.yaml](gitlab-java-springboot-files/03-demo-db/.helm/templates/10-deployment.yaml)

Разумеется, требуется и прописать значения для этих переменных в [values.yaml](03-demo-db/.helm/values.yaml).


