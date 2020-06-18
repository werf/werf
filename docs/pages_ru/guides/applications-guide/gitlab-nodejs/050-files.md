---
title: Работа с файлами
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-nodejs/050-files.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/secret-values.yaml
- config/initializers/aws.rb
- Gemfile
{% endfilesused %}

В разработке может встретиться возможность когда требуется сохранять загружаемые пользователями файлы. Встает резонный вопрос о том каким образом их нужно хранить, и как после этого получать.

Первый и более общий способ. Это использовать как volume в подах [NFS](https://kubernetes.io/docs/concepts/storage/volumes/#nfs), [CephFS](https://kubernetes.io/docs/concepts/storage/volumes/#cephfs) или [hostPath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath), который будет направлен на директорию на ноде, куда будет подключено одно из сетевых хранилищ.

Мы не рекомендуем этот способ, потому что при возникновении неполадок с такими типами volume’ов мы будем влиять на работоспособность всего докера, контейнера и демона docker в целом, тем самым могут пострадать приложения которые даже не имеют отношения к вашему приложению.

Мы рекомендуем пользоваться S3. Такой способ выглядит намного надежнее засчет того что мы используем отдельный сервис, который имеет свойства масштабироваться, работать в HA режиме, и будет иметь высокую доступность.

Есть cloud решения S3, такие как AWS S3, Google Cloud Storage, Microsoft Blobs Storage и т.д. которые будут самым надежным решением из всех что мы можем использовать.

Но для того чтобы просто посмотреть на то как работать с S3 или построить собственное решение, хватит Minio.


## Подключаем наше приложение к S3 Minio


Сначала с помощью npm устанавливает пакет, который так и называется Minio.


```bash
$ npm install minio --save
```


После в исходники в src/js/index.js мы добавляем следующие строки: 



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


И этого вполне хватит для того чтобы вы могли использовать minio S3 в вашем приложении.

Полный пример использования можно посмотреть в тут.

Важно заметить, что мы не указываем жестко параметры подключения прямо в коде, а производим попытку получения их из переменных. Точно так же как и с генерацией статики тут нельзя допускать фиксированных значений. \
 \
Остается только настроить наше приложение со стороны Helm.

Добавляем значения в values.yaml


```yaml
app:
  s3:
    host:
      _default: chat-test-minio
    port:
      _default: 9000
    bucket:
      _default: 'avatars'
    ssl:
      _default: 'false'
```
И в secret-values.yaml

```yaml
app:
  s3:
    access_key:
      _default: bNGXXCF1GF
    secret_key:
      _default: zpThy4kGeqMNSuF2gyw48cOKJMvZqtrTswAQ
```
Далее мы добавляем переменные непосредственно в Deployment с нашим приложением: \



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
Тот способ которым мы передаем переменные в под, с помощью GO шаблонов означает:
  1. `{{ pluck .Values.global.env .Values.app.s3.access_key | first` пробуем взять из поля _app.s3.access_key_ значение из поля которое равно environment которое werf берет из текущей стадии в .gitlab-ci.yml.

  2. `default .Values.app.s3.access_key._default }} `и если такого нет, то мы берём значение из поля _default.

И всё, этого достаточно!

<div>
    <a href="060-email.html" class="nav-btn">Далее: Работа с электронной почтой</a>
</div>
