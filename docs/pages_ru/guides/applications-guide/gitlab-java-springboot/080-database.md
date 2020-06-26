---
title: Подключение базы данных
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-java-springboot/080-database.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/templates/job.yaml
- .helm/templates/_envs.tpl
- .helm/requirements.yaml
- .helm/values.yaml
- .helm/secret-values.yaml
- config/database.yml
- .gitlab-ci.yml
{% endfilesused %}

В этой главе мы настроим в нашем базовом приложении продвинутую работу с базой данных, включающую в себя вопросы выполнения миграций. В качестве базы данных возьмём PostgreSQL.

Мы не будем вносить изменения в сборку — будем использовать образы с DockerHub и конфигурировать их и инфраструктуру.

<a name="kubeconfig" />

## Сконфигурировать PostgreSQL в Kubernetes

TODO: нам сначала надо определиться с именем базы, логином, паролем, портом, МЕСТОМ ХРАНЕНИЯ, а потом уже бросаться вот это всё прописывать в конфигах и выбирать какой чарт подключать

Есть два способа подключить: прописать helm-чарт самостоятельно или подключить внешний чарт. Мы рассмотрим второй вариант, подключим PostgreSQL как внешний subchart.

Для этого нужно:

1. прописать изменения в yaml файлы;
2. указать Redis конфиги;
3. подсказать werf, что ему нужно подтягивать subchart.

Пропишем helm-зависимости:

{% snippetcut name=".helm/requirements.yaml" url="#" %}
```yaml
dependencies:
- name: postgresql
  version: 8.0.0
  repository: https://kubernetes-charts.storage.googleapis.com/
  condition: postgresql.enabled
```
{% endsnippetcut %}

Для того чтобы werf при деплое загрузил необходимые нам сабчарты - нужно прописать в `.gitlab-ci.yml` работу с зависимостями

{% snippetcut name=".gitlab-ci.yml" url="#" %}
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

{% snippetcut name=".helm/values.yaml" url="#" %}
```yaml
postgresql:
  enabled: true
  postgresqlDatabase: guided-database
  postgresqlUsername: guide-username
  postgresqlHost: postgres
  imageTag: "12"
  persistence:
    enabled: true
```
{% endsnippetcut %}

Пароль от базы данных мы тоже конфигурируем, но хранить их нужно в секретных переменных. Для этого стоит использовать [механизм секретных переменных](#######TODO). *Вопрос работы с секретными переменными рассматривался подробнее, [когда мы делали базовое приложение](020-basic.html#secret-values-yaml)*.

{% snippetcut name=".helm/secret-values.yaml (зашифрованный)" url="#" %}
```yaml
postgresql:
  postgresqlPassword: 1000b925471a9491456633bf605d7d3f74c3d5028f2b1e605b9cf39ba33962a4374c51f78637b20ce7f7cd27ccae2a3b5bcf
```
{% endsnippetcut %}

После описанных изменений деплой в любое из окружений должно привести к созданию PostgreSQL.

<a name="appconnect" />

## Подключение приложения к базе PostgreSQL

Для подключения Spring приложения к PostgreSQL необходимо установить зависимость ____________ и сконфигурировать:

{% snippetcut name="____________" url="#" %}
```____________
____________
____________
____________
```
{% endsnippetcut %}

Для подключения к базе данных нам, очевидно, нужно знать: хост, порт, имя базы данных, логин, пароль. В коде приложения мы используем несколько переменных окружения: `POSTGRESQL_HOST`, `POSTGRESQL_PORT`, `POSTGRESQL_DATABASE`, `POSTGRESQL_LOGIN`, `POSTGRESQL_PASSWORD`

Настраиваем эти переменные окружения по аналогии с тем, как [настраивали Redis](070-redis.md), но вынесем все переменные в блок `database_envs` в отдельном файле `_envs.tpl`. Это позволит переиспользовать один и тот же блок в Pod-е с базой данных и в Job с миграциями.

{% offtopic title="Как работает вынос части шаблона в блок?" %}

{% snippetcut name=".helm/templates/_envs.tpl" url="#" %}
{% raw %}
```yaml
{{- define "database_envs" }}
- name: POSTGRESQL_HOST
  value: "{{ pluck .Values.global.env .Values.postgre.host | first | default .Values.postgre.host_default | quote }}"
...
{{- end }}
```
{% endraw %}
{% endsnippetcut %}

Вставляя этот блок — не забываем добавить отступы с помощью функции `indent`:

{% snippetcut name=".helm/templates/deployment.yaml" url="#" %}
{% raw %}
```yaml
{{- include "database_envs" . | indent 8 }}
```
{% endraw %}
{% endsnippetcut %}

{% endofftopic %}


{% offtopic title="Какие значения прописываем в переменные окружения?" %}
Будем **конфигурировать хост** через `values.yaml`:

{% snippetcut name=".helm/templates/_envs.tpl" url="#" %}
{% raw %}
```yaml
- name: POSTGRESQL_HOST
  value: "{{ pluck .Values.global.env .Values.postgre.host | first | default .Values.postgre.host_default | quote }}"
```
{% endraw %}
{% endsnippetcut %}

**Конфигурируем логин и порт** через `values.yaml`, просто прописывая значения:

{% snippetcut name=".helm/templates/deployment.yaml" url="#" %}
{% raw %}
```yaml
- name: POSTGRESQL_LOGIN
  value: "{{ pluck .Values.global.env .Values.postgre.login | first | default .Values.postgre.login_default | quote }}"
- name: POSTGRESQL_PORT
  value: "{{ pluck .Values.global.env .Values.postgre.port | first | default .Values.postgre.port_default | quote }}"
```
{% endraw %}
{% endsnippetcut %}

{% snippetcut name="values.yaml" url="#" %}
```yaml
postgre:
   login:
      _default: ____________
   port:
      _default: ____________
```
{% endsnippetcut %}

TODO: Конфигурируем пароль НЕ ПОНЯТНО КАК ВООБЩЕ

{% snippetcut name=".helm/templates/deployment.yaml" url="#" %}
{% raw %}
```yaml
- name: POSTGRESQL_PASSWORD
  value: "{{ pluck .Values.global.env .Values.postgre.password | first | default .Values.postgre.password_default | quote }}"
```
{% endraw %}
{% endsnippetcut %}

{% snippetcut name="secret-values.yaml" url="#" %}
```yaml
postgre:
  password:
    _default: 100067e35229a23c5070ad5407b7406a7d58d4e54ecfa7b58a1072bc6c34cd5d443e
```
{% endsnippetcut %}

{% endofftopic %}

<a name="migrations" />

## Выполнение миграций

Обычной практикой при использовании kubernetes является выполнение миграций в БД с помощью Job в кубернетес. Это единоразовое выполнение команды в контейнере с определенными параметрами.

Однако в нашем случае spring сам выполняет миграции в соответствии с правилами, которые описаны в нашем коде. Подробно о том как это делается и настраивается прописано в [документации](https://docs.spring.io/spring-boot/docs/2.1.1.RELEASE/reference/html/howto-database-initialization.html). Там рассматривается несколько инструментов для выполнения миграций и работы с бд вообще - Spring Batch Database, Flyway и Liquibase. С точки зрения helm-чартов или сборки совсем ничего не поменятся. Все доступы у нас уже прописаны.

<a name="database-fixtures" />

## Накатка фикстур

Аналогично предыдущему пункту про выполнение миграций, накатка фикстур производится фреймворком самостоятельно используя вышеназванные инструменты для миграций. Для накатывания фикстур в зависимости от стенда (dev-выкатываем, production - никогда не накатываем фикстуры) мы опять же можем передавать переменную окружения в приложение. Например, для Flyway это будет `SPRING_FLYWAY_LOCATIONS`, для Spring Batch Database нужно присвоить Java-переменной `spring.batch.initialize-schema` значение переменной из environment. В ней мы в зависимости от окружения проставляем либо `always` либо `never`. Реализации разные, но подход один - обязательно нужно принимать эту переменную извне. Документация та же что и в предыдущем пункте.


<div>
    <a href="090-unittesting.html" class="nav-btn">Далее: Юнит-тесты и Линтеры</a>
</div>
