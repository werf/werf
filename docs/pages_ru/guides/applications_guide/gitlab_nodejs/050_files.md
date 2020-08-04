---
title: Работа с файлами
sidebar: applications_guide
guide_code: gitlab_nodejs
permalink: documentation/guides/applications_guide/gitlab_nodejs/050_files.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/secret-values.yaml
- .helm/values.yaml
- src/js/index.js
{% endfilesused %}

В этой главе мы настроим в нашем базовом приложении работу с пользовательскими файлами. Для этого нам нужно персистентное хранилище.

В идеале — нужно добиться, чтобы приложение было stateless, а данные хранились в S3-совместимом хранилище, например minio или aws s3. Это обеспечивает простое масштабирование, работу в HA режиме и высокую доступность.

{% offtopic title="А есть какие-то способы кроме S3?" %}
Первый и более общий способ — это использовать как [volume](https://kubernetes.io/docs/concepts/storage/volumes/) хранилище [NFS](https://kubernetes.io/docs/concepts/storage/volumes/#nfs), [CephFS](https://kubernetes.io/docs/concepts/storage/volumes/#cephfs) или [hostPath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath).

Мы не рекомендуем этот способ, потому что при возникновении неполадок с такими типами volume’ов мы будем влиять на работоспособность контейнера и всего демона docker в целом, тем самым могут пострадать приложения, не имеющие никакого отношения к вашему.

Более надёжный путь — пользоваться S3. Таким образом мы используем отдельный сервис, который имеет возможность масштабироваться, работать в HA режиме, и иметь высокую доступность. Можно воспользоваться cloud решением, таким, как AWS S3, Google Cloud Storage, Microsoft Blobs Storage и т.д.

Если мы будем сохранять файлы в какой-либо директории у приложения запущенного в Kubernetes, то после перезапуска контейнера все изменения пропадут.
{% endofftopic %}

Данная настройка производится полностью в рамках приложения, рассмотрим подключение к S3 на примере Minio.

Для этого — подключим пакет minio в npm:

```bash
$ npm install minio --save
```

И настраиваем работу с s3 minio в приложении:

{% snippetcut name="src/server/server.js" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/050-files/src/server/server.js" %}
```js
const Minio = require("minio");
const S3_ENDPOINT = process.env.S3_ENDPOINT || "127.0.0.1";
const S3_PORT = Number(process.env.S3_PORT) || 9000;
const TMP_S3_SSL = process.env.S3_SSL || "true";
const S3_SSL = TMP_S3_SSL.toLowerCase() == "true";
const S3_ACCESS_KEY = process.env.S3_ACCESS_KEY || "SECRET";
const S3_SECRET_KEY = process.env.S3_SECRET_KEY || "SECRET";
const S3_BUCKET = process.env.S3_BUCKET || "avatars";
const CDN_PREFIX = process.env.CDN_PREFIX || "http://127.0.0.1:9000";

// S3 client
var s3Client = new Minio.Client({
  endPoint: S3_ENDPOINT,
  port: S3_PORT,
  useSSL: S3_SSL,
  accessKey: S3_ACCESS_KEY,
  secretKey: S3_SECRET_KEY,
});
```
{% endsnippetcut %}

Для работы с S3 необходимо пробросить в ключи доступа в приложение. Для этого стоит использовать [механизм секретных переменных](https://ru.werf.io/documentation/reference/deploy_process/working_with_secrets.html). *Вопрос работы с секретными переменными рассматривался подробнее, [когда мы делали базовое приложение](020_basic.html#secret-values-yaml)*

{% snippetcut name="secret-values.yaml (расшифрованный)" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/050-files/.helm/secret-values.yaml" %}
{% raw %}
```yaml
app:
  s3:
    access_key:
      _default: bNGXXCF1GF
    secret_key:
      _default: zpThy4kGeqMNSuF2gyw48cOKJMvZqtrTswAQ
```
{% endraw %}
{% endsnippetcut %}

А не секретные значения — храним в `values.yaml`

{% snippetcut name="values.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/050-files/.helm/values.yaml" %}
{% raw %}
```yaml
app:
  s3:
    host:
      _default: minio
    port:
      _default: 9000
    bucket:
      _default: 'avatars'
    ssl:
      _default: 'false'
```
{% endraw %}
{% endsnippetcut %}

После того, как значения корректно прописаны и зашифрованы — мы можем пробросить соответствующие значения в Deployment.

{% snippetcut name="deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/050-files/.helm/templates/deployment.yaml" %}
{% raw %}
```yaml
        - name: CDN_PREFIX
          value: {{ printf "%s%s" (pluck .Values.global.env .Values.app.cdn_prefix | first | default .Values.app.cdn_prefix._default) (pluck .Values.global.env .Values.app.s3.bucket | first | default .Values.app.s3.bucket._default) | quote }}
        - name: S3_SSL
          value: {{ pluck .Values.global.env .Values.app.s3.ssl | first | default .Values.app.s3.ssl._default | quote }}
        - name: S3_ENDPOINT
          value: {{ pluck .Values.global.env .Values.app.s3.host | first | default .Values.app.s3.host._default }}
        - name: S3_PORT
          value: {{ pluck .Values.global.env .Values.app.s3.port | first | default .Values.app.s3.port._default | quote }}
        - name: S3_ACCESS_KEY
          value: {{ pluck .Values.global.env .Values.app.s3.access_key | first | default .Values.app.s3.access_key._default }}
        - name: S3_SECRET_KEY
          value: {{ pluck .Values.global.env .Values.app.s3.secret_key | first | default .Values.app.s3.secret_key._default }}
        - name: S3_BUCKET
          value: {{ pluck .Values.global.env .Values.app.s3.bucket | first | default .Values.app.s3.bucket._default }}
```
{% endraw %}
{% endsnippetcut %}

<div>
    <a href="060_email.html" class="nav-btn">Далее: Работа с электронной почтой</a>
</div>
