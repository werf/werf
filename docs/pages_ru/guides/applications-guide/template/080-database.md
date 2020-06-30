---
title: Подключение базы данных
sidebar: applications-guide
permalink: documentation/guides/applications-guide/template/080-database.html
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
{% raw %}
```yaml
dependencies:
- name: postgresql
  version: 8.0.0
  repository: https://kubernetes-charts.storage.googleapis.com/
  condition: postgresql.enabled
```
{% endraw %}
{% endsnippetcut %}

Для того чтобы werf при деплое загрузил необходимые нам сабчарты - нужно прописать в `.gitlab-ci.yml` работу с зависимостями

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
postgresql:
  enabled: true
  postgresqlDatabase: guided-database
  postgresqlUsername: guide-username
  postgresqlHost: postgres
  imageTag: "12"
  persistence:
    enabled: true
```
{% endraw %}
{% endsnippetcut %}

Пароль от базы данных мы тоже конфигурируем, но хранить их нужно в секретных переменных. Для этого стоит использовать [механизм секретных переменных](#######TODO). *Вопрос работы с секретными переменными рассматривался подробнее, [когда мы делали базовое приложение](020-basic.html#secret-values-yaml)*.

{% snippetcut name=".helm/secret-values.yaml (зашифрованный)" url="#" %}
{% raw %}
```yaml
postgresql:
  postgresqlPassword: 1000b925471a9491456633bf605d7d3f74c3d5028f2b1e605b9cf39ba33962a4374c51f78637b20ce7f7cd27ccae2a3b5bcf
```
{% endraw %}
{% endsnippetcut %}

После описанных изменений деплой в любое из окружений должно привести к созданию PostgreSQL.

<a name="appconnect" />

## Подключение приложения к базе PostgreSQL

Для подключения ____________ приложения к PostgreSQL необходимо установить зависимость ____________ и сконфигурировать:

{% snippetcut name="____________" url="#" %}
{% raw %}
```____________
____________
____________
____________
```
{% endraw %}
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
{% raw %}
```yaml
postgre:
   login:
      _default: ____________
   port:
      _default: ____________
```
{% endraw %}
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
{% raw %}
```yaml
postgre:
  password:
    _default: 100067e35229a23c5070ad5407b7406a7d58d4e54ecfa7b58a1072bc6c34cd5d443e
```
{% endraw %}
{% endsnippetcut %}

{% endofftopic %}

<a name="migrations" />

## Выполнение миграций

Работа реальных приложений почти немыслима без выполнения миграций. С точки зрения Kubernetes миграции выполняются созданием объекта Job, который разово запускает под с необходимыми контейнерами. Запуск миграций мы пропишем после каждого деплоя приложения.

{% offtopic title="Как конфигурируем сам Job?" %}

Подробнее о конфигурировании объекта Job можно почитать [в документации Kubernetes](https://kubernetes.io/docs/concepts/workloads/controllers/job/).

Также мы воспользуемся аннотациями Helm [`helm.sh/hook` и `helm.sh/weight`](https://helm.sh/docs/topics/charts_hooks/), чтобы Job выполнялся после того, как применится новая конфигурация.

{% snippetcut name=".helm/templates/job.yaml" url="#" %}
{% raw %}
```yaml
    "helm.sh/hook": post-install,post-upgrade
    "helm.sh/weight": "2"
```
{% endraw %}
{% endsnippetcut %}

{% endofftopic %}

Так как состояние кластера постоянно меняется — мы не можем быть уверены, что на момент запуска миграций база работает и доступна. Поэтому в Job мы добавляем `initContainer`, который не даёт запуститься скрипту миграции, пока не станет доступна база данных.

{% snippetcut name=".helm/templates/job.yaml" url="#" %}
TODO: выправить этот скрипт, он должен использовать наши реальные конфиги
{% raw %}
```yaml
      initContainers:
      - name: wait-postgres
        image: postgres:12
        command:
          - "sh"
          - "-c"
          - "until pg_isready -h {{ pluck .Values.global.env .Values.db.host | first | default .Values.db.host._default }} -U {{ .Values.postgresql.postgresqlUsername }}; do sleep 2; done;"
```
{% endraw %}
{% endsnippetcut %}

И, непосредственно запускаем миграции. При запуске миграций мы используем тот же самый образ что и в Deployment приложения.

{% snippetcut name=".helm/templates/job.yaml" url="#" %}
{% raw %}
```yaml
      - name: migration
        command: [____________]
{{ tuple "basicapp" . | include "werf_container_image" | indent 8 }}
        env:
{{- include "database_envs" . | indent 10 }}
{{ tuple "basicapp" . | include "werf_container_env" | indent 10 }}
```
{% endraw %}
{% endsnippetcut %}

<div>
    <a href="090-unittesting.html" class="nav-btn">Далее: Юнит-тесты и Линтеры</a>
</div>
