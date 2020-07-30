---
title: Подключение базы данных
sidebar: applications_guide
permalink: documentation/guides/applications_guide/gitlab_nodejs/080_database.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/templates/postgres-pv.yaml
- .helm/templates/job.yaml
- .helm/templates/_envs.tpl
- .helm/requirements.yaml
- .helm/values.yaml
- .helm/secret-values.yaml
- .gitlab-ci.yml
- server.js
{% endfilesused %}

В этой главе мы настроим в нашем базовом приложении продвинутую работу с базой данных, включающую в себя вопросы выполнения миграций. В качестве базы данных возьмём PostgreSQL.

Мы не будем вносить изменения в сборку — будем использовать образы с DockerHub и конфигурировать их и инфраструктуру.

<a name="kubeconfig" />

## Сконфигурировать PostgreSQL в Kubernetes

Подключение базы данных само по себе не является сложной задачей, особенно если она находится вне Kubernetes. Особенность подключения заключается только лишь в указании логина, пароля, хоста, порта и названия самой базы данных внутри инстанса с БД в переменных окружения к вашему приложению.

Мы же рассмотрим вариант с базой внутри нашего кластера, удобство такого способа в том, что мы можем невероятно быстро развернуть полноценное окружение для нашего приложения, однако предстоит решить непростую задачу конфигурирования базы данных, в том числе — вопрос конфигурирования места хранения данных базы.

{% offtopic title="А что с местом хранения?" %}
Kubernetes автоматически разворачивает приложение и переносит его с одного сервера на другой. А значит "по умолчанию" приложение может сначала сохранить данные на одном сервере, а потом быть запущено на другом и оказаться без данных. Хранилища нужно конфигурировать и связывать с приложением и есть несколько способов это сделать.
{% endofftopic %}

Есть два способа подключить нашу БД: прописать helm-чарт самостоятельно или подключить внешний чарт. Мы рассмотрим второй вариант, подключим PostgreSQL как внешний subchart.

Более того мы рассмотрим вопрос того каким образом можно обеспечить хранение данных для такой непостоянной сущности как под.

Для этого нужно:
1. Указать Postgres как зависимый сабчарт в requirements.yaml;
2. Сконфигурировать в werf работу с зависимостями;
3. Убедиться что ваш кластер настроен на работу с персистентным хранилищем;
3. Сконфигурировать подключённый сабчарт;
5. Убедиться, что создаётся под с Postgres.

Пропишем helm-зависимости:

{% snippetcut name=".helm/requirements.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.helm/requirements.yaml" %}
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

{% snippetcut name=".gitlab-ci.yml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.gitlab-ci.yml" %}
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

Изучив [документацию](https://github.com/bitnami/charts/tree/master/bitnami/postgresql#parameters) к нашему сабчарту мы можем увидеть, что задать название основной базы данных, пользователя, хоста и пароля и даже версии postgres мы можем через следующие переменные:

{% snippetcut name=".helm/values.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.helm/values.yaml" %}
{% raw %}
```yaml
postgresql:
  enabled: true
  postgresqlDatabase: guided-database
  postgresqlUsername: guide-username
  postgresqlHost: postgres
  imageTag: "12"
```
{% endraw %}
{% endsnippetcut %}

Пароль от базы данных мы тоже конфигурируем, но хранить их нужно в секретных переменных. Для этого стоит использовать [механизм секретных переменных](#######TODO). *Вопрос работы с секретными переменными рассматривался подробнее, [когда мы делали базовое приложение](020_basic.html#secret-values-yaml)*.

{% snippetcut name=".helm/secret-values.yaml (зашифрованный)" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.helm/secret-values.yaml" %}
{% raw %}
```yaml
postgresql:
  postgresqlPassword: 1000b925471a9491456633bf605d7d3f74c3d5028f2b1e605b9cf39ba33962a4374c51f78637b20ce7f7cd27ccae2a3b5bcf
```
{% endraw %}
{% endsnippetcut %}

{% offtopic title="А где база будет хранить свои данные?" %}
Наверное один из самых главных вопросов при разворачивании базы данных - это вопрос о месте её хранения. Каким образом мы можем сохранять данные в Kubernetes если контейнер по определению является чем-то не постоянным.

В качестве хранилища данных в kubernetes может быть настроен специальный [storage class](https://kubernetes.io/docs/concepts/storage/storage-classes/) который позвляет динамически создавать хранилища с помощью внешних ресурсов. Мы не будем подробно расписывать способы настройки этой сущности, т.к. её конфигурация может отличаться в зависимости от версии вашего кластера и места развертывания( на своих ресурсах или в облаке), предлагаем вам прокосультироваться по этому вопросу с людьми поддерживающими ваш кластер.

Хорошим и простым примером `storage-class` является `local-storage`, который мы и решили использовать в качестве типа хранилища для нашей базы данных.

`local-storage` позволяет хранить данные непосредственно на диске самих нод нашего кластера.
{% endofftopic %}

Далее мы указываем настройки `persistence`, с помощью которых мы настроим хранилище для нашей БД.

{% snippetcut name=".helm/values.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.helm/values.yaml" %}
{% raw %}
```yaml
postgresql:
  ....
  persistence:
    enabled: true
    storageClass: "local-storage"
    accessModes:
    - ReadWriteOnce
    size: 8Gi
```
{% endraw %}
{% endsnippetcut %}
На данный момент возможно осталась непонятной только одна настройка:
`accessModes` - настройка которая определят тип доступа к нашему хранилищу. `ReadWriteOnce` - сообщает о том что к нашему хранилищу единовременно может обращаться только один под.


В нашем случае для корректного развертывания БД нам осталось указать только одну сущность - [PersitentVolume](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)

{% snippetcut name="postgres-pv.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.helm/templates/postgres-pv.yaml" %}
{% raw %}
```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  finalizers:
  - kubernetes.io/pv-protection
  name: posgresql-data
spec:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: 8Gi
  local:
    path: /mnt/guided-postgresql-stage/posgresql-data-0
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - article-kube-node-2
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  volumeMode: Filesystem
```
{% endraw %}
{% endsnippetcut %}

Мы не будем подробно останавливаться на каждом параметре этой сущности, т.к. сама её настройка достойна отдельной главы, для глубокого понимания вы можете обратиться к официальной документации Kubernetes.

Поясним лишь отдельные важные моменты настройки:
{% raw %}
```yaml
  local:
    path: /mnt/guided-postgresql-stage/posgresql-data-0
```
{% endraw %}
Таким образом мы указываем путь до директории где на ноде нашего кластера будут хранится данные postgres.
{% raw %}
```yaml
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - article-kube-node-2
```
{% endraw %}
Эта настройка позволяет нам выбрать на какой именно ноде будут хранится наши данные, т.к. если её не указать под с нашей базой может запуститься на другой ноде где, само собой, наших данных не будет.

{% offtopic title="Но каким образом под с нашей базой поймет что ему нужно использовать именно этот PersistenVolume?" %}

В данном случае сабчарт который мы указали автоматически создаст еще одну сущность под названием [PersitentVolumeClaim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims), за его создание отвечают эти [cтроки](https://github.com/bitnami/charts/blob/master/bitnami/postgresql/templates/statefulset.yaml#L442-L460)

Эта сущность является ни чем иным как прослойкой между хранилищем и подом с нашим приложением.

Мы можем подключать её в качестве `volume` прямиком в наш под:
{% raw %}
```yaml
  volumes:
    - name: data
      persistentVolumeClaim:
        claimName: postgres-data
```
{% endraw %}
И затем монитровать в нужное место в контейнере:
{% raw %}
```yaml
      volumeMounts:
        - mountPath: "/var/lib/postgres"
          name: data
```
{% endraw %}
Хороший пример того как это можно сделать со всеми подробностями вы можете найти [тут](https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/)

Однако вопрос того каким образом PersitentVolumeClaim подключается к PersitentVolume остается открытым.
Дело в том что в Kubernetes работает механизм [binding](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#binding), который позволяет подобрать нашему PVC любой PV удовлетворяющий его по указанным параметрам (размер, тип доступа и т.д.). Тем самым указав в PV и в values.yaml эти параметры одинаково мы гарантируем, что наше хранилище будет подключено к нашей базе.

Есть и более очевидный способ соединить наш PV, добавив в него настройку в которой мы укажем имя нашего PVC и имя неймспейса в котором он находится(имя неймспейса мы взяли из переменной которую генерит werf при деплое):
{% raw %}
```yaml
  claimRef:
    namespace: {{ .Values.global.namespace }}
    name: postgres-data
```
{% endraw %}
{% endofftopic %}

После описанных изменений деплой в любое из окружений должно привести к созданию PostgreSQL.

<a name="appconnect" />

## Подключение приложения к базе PostgreSQL

Для подключения NodeJS приложения к PostgreSQL необходимо установить пакет npm `pg` и сконфигурировать:

{% snippetcut name="server.js" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/backend/src/server/server.js" %}
{% raw %}
```js
const pgconnectionString =
  "postgresql://" +process.env.POSTGRESQL_LOGIN +":" + process.env.POSTGRESQL_PASSWORD + "@"+process.env.POSTGRESQL_HOST +":" + process.env.POSTGRESQL_PORT+"/"+ process.env.POSTGRESQL_DATABASE || "postgresql://127.0.0.1/postgres";
// Postgres connect
const pool = new pg.Pool({
  connectionString: pgconnectionString,
});
pool.on("error", (err, client) => {
  console.error("Unexpected error on idle client", err);
  process.exit(-1);
});
```
{% endraw %}
{% endsnippetcut %}

Для подключения к базе данных нам, очевидно, нужно знать: хост, порт, имя базы данных, логин, пароль. В коде приложения мы используем несколько переменных окружения: `POSTGRESQL_HOST`, `POSTGRESQL_PORT`, `POSTGRESQL_DATABASE`, `POSTGRESQL_LOGIN`, `POSTGRESQL_PASSWORD`

Настраиваем эти переменные окружения по аналогии с тем, как [настраивали Redis](070_redis.md), но вынесем все переменные в блок `database_envs` в отдельном файле `_envs.tpl`. Это позволит переиспользовать один и тот же блок в Pod-е с базой данных и в Job с миграциями.

{% offtopic title="Как работает вынос части шаблона в блок?" %}

{% snippetcut name=".helm/templates/_envs.tpl" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.helm/templates/_envs.tpl" %}
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

{% snippetcut name=".helm/templates/deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.helm/templates/deployment.yaml" %}
{% raw %}
```yaml
{{- include "database_envs" . | indent 8 }}
```
{% endraw %}
{% endsnippetcut %}

{% endofftopic %}


{% offtopic title="Какие значения прописываем в переменные окружения?" %}
Будем **конфигурировать хост** через `values.yaml`:

{% snippetcut name=".helm/templates/_envs.tpl" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.helm/templates/_envs.tpl" %}
{% raw %}
```yaml
- name: POSTGRESQL_HOST
  value: "{{ pluck .Values.global.env .Values.postgresql.host | first | default .Values.postgresql.host_default | quote }}"
```
{% endraw %}
{% endsnippetcut %}

**Конфигурируем логин и порт** через `values.yaml`, просто прописывая значения:

{% snippetcut name=".helm/templates/deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.helm/templates/deployment.yaml" %}
{% raw %}
```yaml
- name: POSTGRESQL_LOGIN
  value: "{{ pluck .Values.global.env .Values.postgresql.login | first | default .Values.postgresql.login_default | quote }}"
- name: POSTGRESQL_PORT
  value: "{{ pluck .Values.global.env .Values.postgresql.port | first | default .Values.postgresql.port_default | quote }}"
```
{% endraw %}
{% endsnippetcut %}

{% snippetcut name="values.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.helm/values.yaml" %}
{% raw %}
```yaml
postgresql:
   login:
      _default: chat
   port:
      _default: 5432
```
{% endraw %}
{% endsnippetcut %}

**Конфигурируем пароль** через `values.yaml`, просто прописывая значения:

{% snippetcut name=".helm/templates/deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.helm/templates/deployment.yaml" %}
{% raw %}
```yaml
- name: POSTGRESQL_PASSWORD
  value: "{{ pluck .Values.global.env .Values.postgresql.password | first | default .Values.postgresql.password_default | quote }}"
```
{% endraw %}
{% endsnippetcut %}

{% snippetcut name="secret-values.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.helm/secret-values.yaml" %}
{% raw %}
```yaml
postgresql:
  password:
    _default: 100067e35229a23c5070ad5407b7406a7d58d4e54ecfa7b58a1072bc6c34cd5d443e
```
{% endraw %}
{% endsnippetcut %}

{% endofftopic %}


{% offtopic title="Безопасно ли хранить пароли в переменных к нашему приложению?" %}

Мы выбрали самый быстрый и простой способ того как вы можете более менее безопасно хранить ваши данные в репозитории и передавать из в приложение. Но важно осозновать, что любой человек с доступом к приложению в кластере, в особенности к самой сущности пода сможет получить любой пароль просто прописав команду `env` внутри запущенного контейнера.
Можно избежать этого не выдавая доступы на исполнение команд внутри контейнеров кому либо, либо собирать свои собственные образа с нуля, убирая из них небезопасные команды, а сами переменные класть в [сущность Secret](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-environment-variables) и запрещать доступ к ним кому либо кроме доверенных лиц.

Но на данный момент мы настоятельно не рекомендуем использовать сущности самого Kubernetes для хранения любой секретной информации. Вместо этого попробуйте использовать инструменты созданные специально для таких случаев, например [Vault](https://www.vaultproject.io/). И затем интегрировав клиент Vault прямо в ваше приложение, получать любые настройки непосредственно при запуске.

{% endofftopic %}
<a name="migrations" />

## Выполнение миграций

Работа реальных приложений почти немыслима без выполнения миграций. С точки зрения Kubernetes миграции выполняются созданием объекта Job, который разово запускает под с необходимыми контейнерами. Запуск миграций мы пропишем после каждого деплоя приложения.

Для выполнения миграций в БД мы используем пакет `node-pg-migrate`, на его примере и будем рассматривать выполнение миграций.

{% offtopic title="Как настраиваем node-pg-migrate?" %}

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

Далее нам необходимо добавить запуск миграций непосредственно в Kubernetes.
Запуск миграций мы произведем с помощью создания сущности Job в kubernetes. Это единоразовый запуск пода с необходимыми нам контейнерами, который подразумевает, то что он он выполнит какую то конечную функцию и завершиться. И может быть запущен только при последующих деплоях.

{% offtopic title="Как конфигурируем сам Job?" %}

Подробнее о конфигурировании объекта Job можно почитать [в документации Kubernetes](https://kubernetes.io/docs/concepts/workloads/controllers/job/).

Также мы воспользуемся аннотациями Helm [`helm.sh/hook` и `helm.sh/weight`](https://helm.sh/docs/topics/charts_hooks/), чтобы Job выполнялся после того, как применится новая конфигурация.

{% snippetcut name=".helm/templates/job.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.helm/templates/job.yaml" %}
{% raw %}
```yaml
    "helm.sh/hook": post-install,post-upgrade
    "helm.sh/weight": "1"
```
{% endraw %}
{% endsnippetcut %}

Вопросы по настройке нашего Job могут возникнуть уже на том этапе когда мы видим блок `annotations:`
Он содержит в себе настройки для `helm`, которые определяют когда именно нужно запускать наш Job, подробнее про них можно узнать [тут](https://v2.helm.sh/docs/charts_hooks/)
В данном случае первой аннотацией мы указываем, что Джоб нужно запускать только после того как все объекты нашего чарта будут загружены и запущены в Kubernetes.

А вторая аннотация отвечает за порядок запуска этого Job, например если мы имеем в нашем чарте несколько Job разного назначения и мы не хотим запускать их единовременно, а только по порядку, мы можем указать для них веса.

{% endofftopic %}

Так как состояние кластера постоянно меняется — мы не можем быть уверены, что на момент запуска миграций база работает и доступна. Поэтому в Job мы добавляем `initContainer`, который не даёт запуститься скрипту миграции, пока не станет доступна база данных.

{% snippetcut name=".helm/templates/job.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.helm/templates/job.yaml" %}
{% raw %}
```yaml
      initContainers:
      - name: wait-postgres
        image: postgres:12
        command:
          - "sh"
          - "-c"
          - "until pg_isready -h {{ pluck .Values.global.env .Values.postgresql.host | first | default .Values.postgresql.host._default }} -U {{ .Values.postgresql.postgresqlUsername }}; do sleep 2; done;"
```
{% endraw %}
{% endsnippetcut %}

И, непосредственно запускаем миграции. При запуске миграций мы используем тот же самый образ что и в Deployment приложения.

{% snippetcut name=".helm/templates/job.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/080-database/.helm/templates/job.yaml" %}
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

И на этом все, далее мы производим те же действия что мы описывали ранее. Готовые пример вы можете найти тут - [fixtures.yaml](/werf-articles/gitlab_nodejs-files/examples/fixtures.yaml)

<div>
    <a href="090_unittesting.html" class="nav-btn">Далее: Юнит-тесты и Линтеры</a>
</div>
