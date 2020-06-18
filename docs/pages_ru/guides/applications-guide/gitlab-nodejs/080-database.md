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


Для текущего примера в приложении должны быть установлены необходимые зависимости. В данном случае мы рассмотрим подключение нашего приложения на базе пакета npm `pg`.

<a name="database-generic" />

## Общий подход

TODO: ....

<a name="database-app-connection" />

## Подключение приложения к базе данных

Внутри кода приложения подключение нашей базы будет выглядеть так:

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


В данном случае мы также используем сабчарт для деплоя базы из того же репозитория. Этого должно хватить для нашего небольшого приложения. В случае большой высоконагруженной инфраструктуры деплой базы непосредственно в кубернетес не рекомендуется. 

Добавляем информацию о подключении в values.yaml


```yaml
app:
 postgresql:
    host:
      stage: chat-test-postgresql
      production: postgres
    user: chat
    db: chat
```


И в secret-values.yaml


```yaml
app:
  postgresql:
    password:
      stage: 1000acb579eaee19bec317079a014346d6aab66bbf84e4a96b395d4a5e669bc32dd1
      production: 1000acb5f9eaee19basdc3127sfa79a014346qwr12b66bbf84e4a96b395d4a5e631255ad1
```


Далее привносим подключени внутрь манифеста нашего приложения переменную подключения:



```yaml
       - name: DATABASE_URL
         value: "postgresql://{{ .Values.app.postgresql.user }}:{{ pluck .Values.global.env .Values.app.postgresql.password | first }}@{{ pluck .Values.global.env .Values.app.postgresql.host | first }}:5432/{{ .Values.app.postgresql.db }}"
```

<a name="database-migrations" />

## Выполнение миграций

Для выполнения миграций в БД мы используем пакет `node-pg-migrate`, на его примере и будем рассматривать выполнение миграций.

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

Далее нам необходимо добавить запуск миграций непосредственно в Kubernetes.
Запуск миграций производится созданием сущности Job в kubernetes. Это единоразовый запуск пода с необходимыми нам контейнерами.

Добавим запуск миграций после каждого деплоя приложения. Потому создадим в нашей директории с манифестами еще один файл - [migrations.yaml](/werf-articles/gitlab-nodejs-files/examples/migrations.yaml)

<details><summary>migrations.yaml</summary>
<p>

```yaml
---
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Chart.Name }}-migrate-db
  annotations:
    "helm.sh/hook": post-install,post-upgrade
    "helm.sh/hook-weight": "2"
spec:
  backoffLimit: 0
  template:
    metadata:
      name: {{ .Chart.Name }}-migrate-db
    spec:
      initContainers:
      - name: wait-postgres
        image: postgres:12
        command:
          - "sh"
          - "-c"
          - "until pg_isready -h {{ pluck .Values.global.env .Values.app.postgresql.host | first }} -U {{ .Values.app.postgresql.user }}; do sleep 2; done;"
      containers:
      - name: node
{{ tuple "node" . | include "werf_container_image" | indent 8 }}
        command: ["npm", "migrate"]
        env:
        - name: DATABASE_URL
          value: "postgresql://{{ .Values.app.postgresql.user }}:{{ pluck .Values.global.env .Values.app.postgresql.password | first }}@{{ pluck .Values.global.env .Values.app.postgresql.host | first }}:5432/{{ .Values.app.postgresql.db }}"
{{ tuple "node" . | include "werf_container_env" | indent 10 }}
      restartPolicy: Never
```

</p>
</details>

Аннотации `"helm.sh/hook": post-install,post-upgrade` указывают условия запуска job а `"helm.sh/hook-weight": "2"` указывают на порядок выполнения (от меньшего к большему)
`backoffLimit: 0` - запрещает перезапускать наш под с Job, тем самым гарантируя что он выполнится всего один раз при деплойменте.

`initContainers:` - блок описывающий контейнеры которые будут отрабатывать единоразово перед запуском основных контейнеров. Это самый быстрый и удобный способ для подготовки наших приложений к запуску в реальном времени. В данном случае мы взяли официальный образ с postgres, для того чтобы с помощью его инструментов отследить запустилась ли наша база.
```yaml
        command:
          - "sh"
          - "-c"
          - "until pg_isready -h {{ pluck .Values.global.env .Values.app.postgresql.host | first }} -U {{ .Values.app.postgresql.user }}; do sleep 2; done;"
```
Мы используем цикл, который каждые 2 секунды будет проверять нашу базу на доступность. После того как команда `pg_isready` сможет выполнится успешно, инит контейнер завершит свою работу, а основной запустится и без проблем сразу сможет подключится к базе. 

При запуске миграций мы используем тот же самый образ что и в деплойменте. Различие только в запускаемых командах.

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
