---
title: Работа с файлами
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-java-springboot/050-files.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/secret-values.yaml
- config/initializers/aws.rb
- Gemfile
{% endfilesused %}

В этой главе мы настроим в нашем базовом приложении работу с пользовательскими файлами. Для этого нам нужно персистентное хранилище.

В идеале — нужно добиться, чтобы приложение было stateless, а данные хранились в S3-совместимом хранилище, например minio или aws s3. Это обеспечивает простое масштабирование, работу в HA режиме и высокую доступность.

{% offtopic title="А есть какие-то способы кроме S3?" %}
TODO: пишем про 
Первый и более общий способ. Это использовать как volume в подах [NFS](https://kubernetes.io/docs/concepts/storage/volumes/#nfs), [CephFS](https://kubernetes.io/docs/concepts/storage/volumes/#cephfs) или [hostPath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath), который будет направлен на директорию на ноде, куда будет подключено одно из сетевых хранилищ.

Мы не рекомендуем этот способ, потому что при возникновении неполадок с такими типами volume’ов мы будем влиять на работоспособность всего докера, контейнера и демона docker в целом, тем самым могут пострадать приложения которые даже не имеют отношения к вашему приложению.

Мы рекомендуем пользоваться S3. Такой способ выглядит намного надежнее засчет того что мы используем отдельный сервис, который имеет свойства масштабироваться, работать в HA режиме, и будет иметь высокую доступность.

Есть cloud решения S3, такие как AWS S3, Google Cloud Storage, Microsoft Blobs Storage и т.д. которые будут самым надежным решением из всех что мы можем использовать.

Если мы будем сохранять файлы какой - либо директории у приложения запущенного в kubernetes - то после перезапуска контейнера все изменения пропадут.
{% endofftopic %}

Данная настройка производится полностью в рамках приложения. Нужно подключить `aws-java-sdk`, сконфигурировать его и начать использовать.

Пропишем использование `aws-java-sdk` как зависимость:

{% snippetcut name="pom.xml" url="gitlab-java-springboot-files/02-demo-with-assets/pom.xml:27-31" %}
```xml
<dependency>
   <groupId>com.amazonaws</groupId>
   <artifactId>aws-java-sdk</artifactId>
   <version>1.11.133</version>
</dependency>
```
{% endsnippetcut %}

И сконфигурируем:

{% snippetcut name="src/main/resources/application.properties" url="gitlab-java-springboot-files/02-demo-with-assets/src/main/resources/application.properties" %}
```yaml
amazonProperties:
  endpointUrl: ${S3ENDPOINT}
  accessKey: ${S3KEY}
  secretKey: ${S3SECRET}
  bucketName: ${S3BUCKET}
```
{% endsnippetcut %}

Для работы с S3 необходимо пробросить в ключи доступа в приложение. Для этого стоит использовать [механизм секретных переменных](#######TODO). *Вопрос работы с секретными переменными рассматривался подробнее, [когда мы делали базовое приложение](020-basic.html#secret-values-yaml)*

{% snippetcut name="secret-values.yaml (расшифрованный)" url="gitlab-java-springboot-files/02-demo-with-assets/.helm/secret-values.yaml" %}
```yaml
app:
  s3:
    key:
      _default: mys3keyidstage
      production: mys3keyidprod
    secret:
      _default: mys3keysecretstage
      production: mys3keysecretprod
{% endsnippetcut %}

А не секретные значения — храним в `values.yaml`

{% snippetcut name="values.yaml" url="gitlab-java-springboot-files/02-demo-with-assets/.helm/values.yaml:8-13" %}
app:
  s3:
    epurl:
      _default: https://s3.us-east-2.amazonaws.com
    bucket:
      _default: mydefaultbucket
      production: myproductionbucket
```
{% endsnippetcut %}

После того, как значения корректно прописаны и зашифрованы — мы можем пробросить соответствующие значения в Deployment.

{% snippetcut name="deployment.yaml" url="gitlab-java-springboot-files/02-demo-with-assets/.helm/templates/10-deployment.yaml:53-60" %}
```yaml
       env:
{{ tuple "hello" . | include "werf_container_env" | indent 8 }}
        - name: S3ENDPOINT
          value: {{ pluck .Values.global.env .Values.app.s3.epurl | first | default .Values.app.s3.epurl._default | quote }}
        - name: S3KEY
          value: {{ pluck .Values.global.env .Values.app.s3.key | first | default .Values.app.s3.key._default | quote }}
        - name: S3SECRET
          value: {{ pluck .Values.global.env .Values.app.s3.secret | first | default .Values.app.s3.secret._default | quote }}
        - name: S3BUCKET
          value: {{ pluck .Values.global.env .Values.app.s3.bucket | first | default .Values.app.s3.bucket._default | quote }}
```
{% endsnippetcut %}

TODO: надо дать отсылку на какой-то гайд, где описано, как конкретно использовать гем aws-sdk. Мало же просто его установить — надо ещё как-то юзать в коде.

<div>
    <a href="060-email.html" class="nav-btn">Далее: Работа с электронной почтой</a>
</div>
