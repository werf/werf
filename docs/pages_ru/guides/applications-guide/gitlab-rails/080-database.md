---
title: Подключение базы данных
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-rails/080-database.html
author: alexey.chazov <alexey.chazov@flant.com>
layout: guide
toc: false
author_team: "bravo"
author_name: "alexey.chazov"
ci: "gitlab"
language: "ruby"
framework: "rails"
is_compiled: 0
package_managers_possible:
 - bundler
package_managers_chosen: "bundler"
unit_tests_possible:
 - Rspec
unit_tests_chosen: "Rspec"
assets_generator_possible:
 - webpack
 - gulp
assets_generator_chosen: "webpack"
---

TODO: а где всё хранится? Почему не слова о сторейджах?

В этой главе мы настроим в нашем базовом приложении продвинутую работу с базой данных, включающую в себя вопросы выполнения миграций. В качестве базы данных возьмём PostgreSQL.

Мы не будем вносить изменения в сборку — будем использовать образы с DockerHub и конфигурировать их и инфраструктуру.

<a name="kubeconfig" />

## Сконфигурировать PostgreSQL в Kubernetes

Есть два способа подключить: прописать helm-чарт самостоятельно или подключить внешний чарт. Мы рассмотрим второй вариант, подключим PostgreSQL как внешний subchart.

Для этого нужно:

1. прописать изменения в yaml файлы;
2. указать Redis конфиги;
3. подсказать werf, что ему нужно подтягивать subchart.

Пропишем helm-зависимости:

{% snippetcut name=".helm/requirements.yaml" url="gitlab-rails-files/examples/example_3/.helm/requirements.yaml" %}
```yaml
dependencies:
- name: postgresql
  version: 8.0.0
  repository: https://kubernetes-charts.storage.googleapis.com/
  condition: postgresql.enabled
```
{% endsnippetcut %}

Для того чтобы werf при деплое загрузил необходимые нам сабчарты - нужно прописать в `.gitlab-ci.yml` работу с зависимостями

{% snippetcut name=".gitlab-ci.yml" url="gitlab-rails-files/examples/example_3/.gitlab-ci.yml#L24" %}
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

{% snippetcut name=".helm/values.yaml" url="gitlab-rails-files/examples/example_3/.helm/values.yaml#L4" %}
```yaml
postgresql:
  enabled: true
  postgresqlDatabase: hello_world
  postgresqlUsername: hello_world_user
  postgresqlHost: postgres
  imageTag: "12"
  persistence:
    enabled: true
```
{% endsnippetcut %}

Пароль от базы данных мы тоже конфигурируем, но хранить их нужно в секретных переменных. Для этого стоит использовать [механизм секретных переменных](#######TODO). *Вопрос работы с секретными переменными рассматривался подробнее, [когда мы делали базовое приложение](020-basic.html#secret-values-yaml)*.

{% snippetcut name=".helm/secret-values.yaml (зашифрованный)" url="gitlab-rails-files/examples/example_3/.helm/secret-values.yaml#L3" %}
```yaml
postgresql:
  postgresqlPassword: 1000b925471a9491456633bf605d7d3f74c3d5028f2b1e605b9cf39ba33962a4374c51f78637b20ce7f7cd27ccae2a3b5bcf
```
{% endsnippetcut %}

После описанных изменений деплой в любое из окружений должно привести к созданию PostgreSQL.

<a name="appconnect" />

## Подключение Rails приложения к базе postgresql

Для подключения Rails приложения к базе необходимо в Gemfile прописать нужный gem (`TODO: какой???`) и правильно сконфигурировать Rails приложение.

Внесем изменения в файл настроек подключения к базе данных:

{% snippetcut name="config/database.yml" url="gitlab-rails-files/examples/example_3/config/database.yml#L17" %}
```yaml
default: &default
  adapter: postgresql
  encoding: unicode
  pool: <%= ENV.fetch("RAILS_MAX_THREADS") { 5 } %>
  host: <%= ENV['DATABASE_HOST'] %>
  username: <%= ENV['DATABASE_USERNAME'] %>
  password: <%= ENV['DATABASE_PASSWORD'] %>
  database: <%= ENV['DATABASE_NAME'] %>

development:
  <<: *default
```
{% endsnippetcut %}

Для того, чтобы переменные окружения попали в контейнер — пропишем их в Pod-е. Для удобства и переиспользуемости — объявим эти переменные в блок `database_envs` в отдельном файле `_envs.tpl` и потом просто подключим в нужном контейнере.

{% snippetcut name=".helm/templates/_envs.tpl" url="gitlab-rails-files/examples/example_3/.helm/templates/_envs.tpl#L10" %}
{% raw %}
```yaml
{{- define "database_envs" }}
- name: DATABASE_HOST
  value: {{ .Chart.Name }}-{{ .Values.global.env }}-postgresql
- name: DATABASE_NAME
  value: {{ .Values.postgresql.postgresqlDatabase }}
- name: DATABASE_USERNAME
  value: {{ .Values.postgresql.postgresqlUsername }}
- name: DATABASE_PASSWORD
  value: {{ .Values.postgresql.postgresqlPassword }}
{{- end }}
```
{% endraw %}
{% endsnippetcut %}

Вставляя этот блок — не забываем добавить отступы с помощью функции `indent`

{% snippetcut name=".helm/templates/deployment.yaml" url="gitlab-rails-files/examples/example_3/.helm/templates/deployment.yaml#L24" %}
{% raw %}
```yaml
{{- include "database_envs" . | indent 8 }}
```
{% endraw %}
{% endsnippetcut %}

Остальные значения подставляются из файлов `values.yaml` и `secret-values.yaml`

<a name="migrations" />

## Выполнение миграций

Работа реальных приложений почти немыслима без выполнения миграций. С точки зрения Kubernetes миграции выполняются созданием объекта Job, который разово запускает под с необходимыми контейнерами. Запуск миграций мы пропишем после каждого деплоя приложения.

TODO: разве "после деплоя, но до доступности"????

TODO: вот этот пиздец внизу надо нарезать и рассказать по кускам.

{% snippetcut name=".helm/templates/job.yaml" url="gitlab-rails-files/examples/example_3/.helm/templates/job.yaml" %}
{% raw %}
```yaml
---
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ $.Values.global.werf.name }}-migrate-db
  annotations:
    "helm.sh/hook": post-install,post-upgrade
    "helm.sh/hook-weight": "2"
spec:
  backoffLimit: 0
  template:
    metadata:
      name: {{ $.Values.global.werf.name }}-migrate-db
    spec:
      initContainers:
      - name: wait-postgres
        image: postgres:12
        command:
          - "sh"
          - "-c"
          - "until pg_isready -h {{ .Chart.Name }}-{{ .Values.global.env }}-postgresql -U {{ .Values.postgresql.postgresqlUsername }}; do sleep 2; done;"
      containers:
      - name: rails
{{ tuple "rails" . | include "werf_container_image" | indent 8 }}
        command: ["bundle", "exec", "rake", "db:migrate"]
        env:
{{- include "apps_envs" . | indent 10 }}
{{- include "database_envs" . | indent 10 }}
{{ tuple "rails" . | include "werf_container_env" | indent 10 }}
      restartPolicy: Never
```
{% endraw %}
{% endsnippetcut %}


Аннотации `"helm.sh/hook": post-install,post-upgrade` указывают условия запуска job а `"helm.sh/hook-weight": "2"` указывают на порядок выполнения (от меньшего к большему)

При запуске миграций мы используем тот же самый образ что и в деплойменте. Различие только в запускаемых командах.

