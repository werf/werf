---
title: Подключение базы данных
sidebar: applications-guide
permalink: documentation/guides/applications-guide/template/080-database.html
layout: guide
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



Для текущего примера в приложении должны быть установлены необходимые зависимости. В качестве примера - мы возьмем приложение для работы которого необходима база данных.


## Как подключить БД


Подключим postgresql helm сабчартом, для этого внесем изменения в файл `.helm/requirements.yaml`


```
dependencies:
- name: postgresql
  version: 8.0.0
  repository: https://kubernetes-charts.storage.googleapis.com/
  condition: postgresql.enabled
```


Для того чтобы werf при деплое загрузил необходимые нам сабчарты - нужно добавить команды в .gitlab-ci


```
.base_deploy:
  stage: deploy
  script:
    - werf helm repo init
    - werf helm dependency update
    - werf deploy
```


Опишем параметры для postgresql в файле `.helm/values.yaml`


```
postgresql:
  enabled: true
  postgresqlDatabase: hello_world
  postgresqlUsername: hello_world_user
  postgresqlHost: postgres
  imageTag: "12"
  persistence:
    enabled: true
```


## Подключение Django приложения к базе postgresql

Для того, чтобы django научился взаимодействовать с postgresql, нужно поставить дополнительный пакет `pip install psycopg2-binary`.

Настройки подключения нашего приложения к базе данных мы будем передавать через переменные окружения. Такой подход позволит нам использовать один и тот же образ в разных окружениях, что должно исключить запуск непроверенного кода в production окружении.

Внесем изменения в файл настроек `settings.py`.


```
DATABASES = {
    'default': {
        'ENGINE': os.environ.get('SQL_ENGINE', 'django.db.backends.sqlite3'),
        'NAME': os.environ.get('SQL_DATABASE', os.path.join(BASE_DIR, 'db.sqlite3')),
        'USER': os.environ.get('SQL_USER', 'user'),
        'PASSWORD': os.environ.get('SQL_PASSWORD', 'password'),
        'HOST': os.environ.get('SQL_HOST', 'localhost'),
        'PORT': os.environ.get('SQL_PORT', '5432'),
    }
}
```


Параметры подключения приложения к базе данным мы опишем в файле `.helm/templates/_envs.tpl`


```
{{- define "database_envs" }}
- name: SQL_HOST
  value: "postgres://{{ .Values.postgresql.postgresqlUsername }}:{{ .Values.postgresql.postgresqlPassword }}@{{ .Chart.Name }}-{{ .Values.global.env }}-postgresql:5432"
- name: SQL_DATABASE
  value: {{ .Values.postgresql.postgresqlDatabase }}
- name: SQL_USER
  value: {{ .Values.postgresql.postgresqlUser }}
- name: SQL_PASSWORD
  value: {{ .Values.postgresql.postgresqlPassword }}
- name: SQL_PORT
  value: {{ .Values.postgresql.postgresqlPort }}
- name: SQL_ENGINE
  value: "django.db.backends.postgresql_psycopg2"
{{- end }}
```


Такой подход позволит нам переиспользовать данное определение переменных окружения для нескольких контейнеров. Имя для сервиса postgresql генерируется из названия нашего приложения, имени окружения и добавлением postgresql

Остальные значения подставляются из файлов values.yaml и secret-values.yaml


Пароль от базы данных добавим в `secret-values.yaml`

## Выполнение миграций


Запуск миграций производится созданием приметива Job в kubernetes. Это единоразовый запуск пода с необходимыми нам контейнерами.

Добавим запуск миграций после каждого деплоя приложения.


```yaml
---
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Chart.Name }}-migrations
  annotations:
    "helm.sh/hook": post-install,post-upgrade
    "helm.sh/weight": "2"
spec:
  activeDeadlineSeconds: 600
  template:
    metadata:
      name: {{ .Chart.Name }}-migrate
    spec:
      restartPolicy: Never
      initContainers:
      - name: wait-postgres
        command: ['/bin/sh', '-c', 'while ! getent ahostsv4 {{ pluck .Values.global.env .Values.db.host | first | default .Values.db.host._default }}; do sleep 1; done']
        image: alpine:3.6
        env:
      containers:
      - name: migration
        args: ["python manage.py migrate --noinput"]
        command:
        - /bin/bash
        - -c
        - --
        workingDir: "/usr/src/app"
{{ tuple "django" . | include "werf_container_image" | indent 8 }}
        env:
{{- include "env_app" . | indent 8 }}
{{ tuple "django" . | include "werf_container_env" | indent 8 }}
```


Аннотации `"helm.sh/hook": post-install,post-upgrade` указывают условия запуска job а `"helm.sh/hook-weight": "2"` указывают на порядок выполнения (от меньшего к большему)

При запуске миграций мы используем тот же самый образ что и в деплойменте. Различие только в запускаемых командах.


