---
title: Подключение базы данных
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-nodejs/080-database.html
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

{% snippetcut name=".helm/requirements.yaml" url="template-files/examples/example_3/.helm/requirements.yaml" %}
```yaml
dependencies:
- name: postgresql
  version: 8.0.0
  repository: https://kubernetes-charts.storage.googleapis.com/
  condition: postgresql.enabled
```
{% endsnippetcut %}

Для того чтобы werf при деплое загрузил необходимые нам сабчарты - нужно прописать в `.gitlab-ci.yml` работу с зависимостями

{% snippetcut name=".gitlab-ci.yml" url="template-files/examples/example_3/.gitlab-ci.yml#L24" %}
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

{% snippetcut name=".helm/values.yaml" url="template-files/examples/example_3/.helm/values.yaml#L4" %}
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

{% snippetcut name=".helm/secret-values.yaml (зашифрованный)" url="template-files/examples/example_3/.helm/secret-values.yaml#L3" %}
```yaml
postgresql:
  postgresqlPassword: 1000b925471a9491456633bf605d7d3f74c3d5028f2b1e605b9cf39ba33962a4374c51f78637b20ce7f7cd27ccae2a3b5bcf
```
{% endsnippetcut %}

После описанных изменений деплой в любое из окружений должно привести к созданию PostgreSQL.

<a name="appconnect" />

## Подключение приложения к базе PostgreSQL

Для подключения NodeJS приложения к PostgreSQL необходимо установить пакет npm `pg` и сконфигурировать:

TODO: выправить использование переменных в этом коде

{% snippetcut name="____________" url="____________" %}
```js
const pgconnectionString =
  process.env.DATABASE_URL || "postgresql://127.0.0.1/postgres";

// Postgres connect
const pool = new pg.Pool({
  connectionString: pgconnectionString,
});
pool.on("error", (err, client) => {
  console.error("Unexpected error on idle client", err);
  process.exit(-1);
});
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

{% snippetcut name=".helm/templates/_envs.tpl" url="____________" %}
{% raw %}
```yaml
- name: POSTGRESQL_HOST
  value: "{{ pluck .Values.global.env .Values.postgre.host | first | default .Values.postgre.host_default | quote }}"
```
{% endraw %}
{% endsnippetcut %}

**Конфигурируем логин и порт** через `values.yaml`, просто прописывая значения:

{% snippetcut name=".helm/templates/deployment.yaml" url="____________" %}
{% raw %}
```yaml
- name: POSTGRESQL_LOGIN
  value: "{{ pluck .Values.global.env .Values.postgre.login | first | default .Values.postgre.login_default | quote }}"
- name: POSTGRESQL_PORT
  value: "{{ pluck .Values.global.env .Values.postgre.port | first | default .Values.postgre.port_default | quote }}"
```
{% endraw %}
{% endsnippetcut %}

{% snippetcut name="values.yaml" url="____________" %}
```yaml
postgre:
   login:
      _default: ____________
   port:
      _default: ____________
```
{% endsnippetcut %}

TODO: Конфигурируем пароль ХУЙ ЗНАЕТ КАК ВООБЩЕ

{% snippetcut name=".helm/templates/deployment.yaml" url="____________" %}
{% raw %}
```yaml
- name: POSTGRESQL_PASSWORD
  value: "{{ pluck .Values.global.env .Values.postgre.password | first | default .Values.postgre.password_default | quote }}"
```
{% endraw %}
{% endsnippetcut %}

{% snippetcut name="secret-values.yaml" url="____________" %}
```yaml
postgre:
  password:
    _default: 100067e35229a23c5070ad5407b7406a7d58d4e54ecfa7b58a1072bc6c34cd5d443e
```
{% endsnippetcut %}

{% endofftopic %}

<a name="migrations" />

## Выполнение миграций

Работа реальных приложений почти немыслима без выполнения миграций. С точки зрения Kubernetes миграции выполняются созданием объекта Job, который разово запускает под с необходимыми контейнерами. Запуск миграций мы пропишем после каждого деплоя приложения.

Для выполнения миграций в БД мы используем пакет `node-pg-migrate`, на его примере и будем рассматривать выполнение миграций.

{% offtopic title="Как настраиваем node-pg-migrate?" %}
TODO: причесать этот текст

Запуск миграции мы помещаем в package.json, чтобы его можно было  вызывать с помощью скрипта в npm: 
```json
   "migrate": "node-pg-migrate"
```
Сама конфигурация миграций у нас находится в отдельной директории `migrations`, которую мы создали на уровне исходного кода приложения.
```
node
├── migrations
│   ├── 1588019669425_001-users.js
│   └── 1588172704904_add-avatar-status.js
├── src
├── package.json
...
```
{% endofftopic %}

{% offtopic title="Как конфигурируем сам Job?" %}

TODO: разве "после деплоя, но до доступности"????

TODO: надо описать тут концептуальную часть про запуск джоба, вот это про хук и weight и где почитать.

{% snippetcut name=".helm/templates/job.yaml" url="template-files/examples/example_3/.helm/templates/job.yaml" %}
{% raw %}
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Chart.Name }}-migrations
  annotations:
    "helm.sh/hook": post-install,post-upgrade
    "helm.sh/weight": "2"
```
{% endraw %}
{% endsnippetcut %}

{% endofftopic %}

Так как состояние кластера постоянно меняется — мы не можем быть уверены, что на момент запуска миграций база работает и доступна. Поэтому в Job мы добавляем `initContainer`, который не даёт запуститься скрипту миграции, пока не станет доступна база данных.

{% snippetcut name=".helm/templates/job.yaml" url="template-files/examples/example_3/.helm/templates/job.yaml" %}
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

{% snippetcut name=".helm/templates/job.yaml" url="template-files/examples/example_3/.helm/templates/job.yaml" %}
{% raw %}
```yaml
      - name: migration
        command: ["npm", "migrate"]
{{ tuple "basicapp" . | include "werf_container_image" | indent 8 }}
        env:
{{- include "database_envs" . | indent 10 }}
{{ tuple "basicapp" . | include "werf_container_env" | indent 10 }}
```
{% endraw %}
{% endsnippetcut %}

<a name="database-fixtures" />

## Накатка фикстур

При первом деплое вашего приложения, вам может понадобится раскатить фикстуры. В нашем случае это дефолтный пользователь нашего чата.
Мы не будем расписывать подробного этот шаг, потому что он практически не отличается от предыдущего. 

Мы добавляем наши фикстуры в отдельную директорию `fixtures` также как это было с миграциями.
А их запуск добавляем в `package.json`

```json
    "fixtures": "node fixtures/01-default-user.js"
```

И на этом все, далее мы производим те же действия что мы описывали ранее. Готовые пример вы можете найти тут - [fixtures.yaml](/werf-articles/gitlab-nodejs-files/examples/fixtures.yaml)

<div>
    <a href="090-unittesting.html" class="nav-btn">Далее: Юнит-тесты и Линтеры</a>
</div>
