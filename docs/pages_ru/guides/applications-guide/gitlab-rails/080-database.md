---
title: Гайд по использованию Ruby On Rails + GitLab + Werf
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

# Подключаем базу данных

Для текущего примера в приложении должны быть установлены необходимые зависимости. В качестве примера - мы возьмем приложение для работы которого необходима база данных.


## Как подключить БД

Подключим postgresql helm сабчартом, для этого внесем изменения в файл `.helm/requirements.yaml`

[.helm/requirements.yaml](gitlab-rails-files/examples/example_3/.helm/requirements.yaml)
```yaml
dependencies:
- name: postgresql
  version: 8.0.0
  repository: https://kubernetes-charts.storage.googleapis.com/
  condition: postgresql.enabled
```


Для того чтобы werf при деплое загрузил необходимые нам сабчарты - нужно добавить команды в .gitlab-ci

[.gitlab-ci.yml](gitlab-rails-files/examples/example_3/.gitlab-ci.yml#L24)
```yaml
.base_deploy:
  stage: deploy
  script:
    - werf helm repo init
    - werf helm dependency update
    - werf deploy
```


Опишем параметры для postgresql в файле `.helm/values.yaml`

[.helm/values.yaml](gitlab-rails-files/examples/example_3/.helm/values.yaml#L4)
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


Пароль от базы данных добавим в `secret-values.yaml`

[.helm/secret-values.yaml](gitlab-rails-files/examples/example_3/.helm/secret-values.yaml#L3)
```yaml
postgresql:
  postgresqlPassword: 1000b925471a9491456633bf605d7d3f74c3d5028f2b1e605b9cf39ba33962a4374c51f78637b20ce7f7cd27ccae2a3b5bcf
```

## Подключение Rails приложения к базе postgresql

Настройки подключения нашего приложения к базе данных мы будем передавать через переменные окружения. Такой подход позволит нам использовать один и тот же образ в разных окружениях, что должно исключить запуск непроверенного кода в production окружении.

Внесем изменения в файл настроек подключения к базе данных

[config/database.yml](gitlab-rails-files/examples/example_3/config/database.yml#L17)
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
test:
  <<: *default
production:
  <<: *default
```


Параметры подключения приложения к базе данным мы опишем в файле `.helm/templates/_envs.tpl`

[.helm/templates/_envs.tpl](gitlab-rails-files/examples/example_3/.helm/templates/_envs.tpl#L10)
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

Такой подход позволит нам переиспользовать данное определение переменных окружения для нескольких контейнеров. Имя для сервиса postgresql генерируется из названия нашего приложения, имени окружения и добавлением postgresql

[.helm/templates/deployment.yaml](gitlab-rails-files/examples/example_3/.helm/templates/deployment.yaml#L24)
```yaml
{{- include "database_envs" . | indent 8 }}
```

Остальные значения подставляются из файлов `values.yaml` и `secret-values.yaml`


## Выполнение миграций

Запуск миграций производится созданием приметива Job в kubernetes. Это единоразовый запуск пода с необходимыми нам контейнерами.

Добавим запуск миграций после каждого деплоя приложения.

[.helm/templates/job](gitlab-rails-files/examples/example_3/.helm/templates/job.yaml)
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


Аннотации `"helm.sh/hook": post-install,post-upgrade` указывают условия запуска job а `"helm.sh/hook-weight": "2"` указывают на порядок выполнения (от меньшего к большему)

При запуске миграций мы используем тот же самый образ что и в деплойменте. Различие только в запускаемых командах.

## Накатка фикстур при первом выкате


