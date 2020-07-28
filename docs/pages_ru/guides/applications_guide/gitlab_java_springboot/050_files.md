---
title: Работа с файлами
sidebar: applications_guide
permalink: documentation/guides/applications_guide/gitlab_java_springboot/050_files.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/secret-values.yaml
- .helm/values.yaml
- src/main/resources/application.properties
- pom.xml
{% endfilesused %}

В этой главе мы настроим в нашем базовом приложении работу с пользовательскими файлами. Для этого нам нужно персистентное хранилище.

В идеале — нужно добиться, чтобы приложение было stateless, а данные хранились в S3-совместимом хранилище, например minio или aws s3. Это обеспечивает простое масштабирование, работу в HA режиме и высокую доступность.

{% offtopic title="А есть какие-то способы кроме S3?" %}
Первый и более общий способ — это использовать как [volume](https://kubernetes.io/docs/concepts/storage/volumes/) хранилище [NFS](https://kubernetes.io/docs/concepts/storage/volumes/#nfs), [CephFS](https://kubernetes.io/docs/concepts/storage/volumes/#cephfs) или [hostPath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath).

Мы не рекомендуем этот способ, потому что при возникновении неполадок с такими типами volume’ов мы будем влиять на работоспособность контейнера и всего демона docker в целом, тем самым могут пострадать приложения, не имеющие никакого отношения к вашему.

Более надёжный путь — пользоваться S3. Таким образом мы используем отдельный сервис, который имеет возможность масштабироваться, работать в HA режиме, и иметь высокую доступность. Можно воспользоваться cloud решением, таким, как AWS S3, Google Cloud Storage, Microsoft Blobs Storage и т.д.

Если мы будем сохранять файлы в какой-либо директории у приложения запущенного в Kubernetes, то после перезапуска контейнера все изменения пропадут.
{% endofftopic %}

Данная настройка производится полностью в рамках приложения. Нужно подключить `aws-java-sdk`, сконфигурировать его и начать использовать.

Пропишем использование `aws-java-sdk` как зависимость:

{% snippetcut name="pom.xml" url="#" %}
```xml
<dependency>
   <groupId>com.amazonaws</groupId>
   <artifactId>aws-java-sdk</artifactId>
   <version>1.11.133</version>
</dependency>
```
{% endsnippetcut %}

И сконфигурируем:

{% snippetcut name="src/main/resources/application.properties" url="#" %}
{% raw %}
```yaml
amazonProperties:
  endpointUrl: ${S3ENDPOINT}
  accessKey: ${S3KEY}
  secretKey: ${S3SECRET}
  bucketName: ${S3BUCKET}
```
{% endraw %}
{% endsnippetcut %}

Для работы с S3 необходимо пробросить в ключи доступа в приложение. Для этого стоит использовать [механизм секретных переменных](https://ru.werf.io/documentation/reference/deploy_process/working_with_secrets.html). *Вопрос работы с секретными переменными рассматривался подробнее, [когда мы делали базовое приложение](020_basic.html#secret-values-yaml)*

{% snippetcut name="secret-values.yaml (расшифрованный)" url="#" %}
{% raw %}
```yaml
app:
  s3:
    key:
      _default: mys3keyidstage
      production: mys3keyidprod
    secret:
      _default: mys3keysecretstage
      production: mys3keysecretprod
```
{% endraw %}
{% endsnippetcut %}

А не секретные значения — храним в `values.yaml`

{% snippetcut name="values.yaml" url="#" %}
{% raw %}
```yaml
app:
  s3:
    epurl:
      _default: https://s3.us-east-2.amazonaws.com
    bucket:
      _default: mydefaultbucket
      production: myproductionbucket
```
{% endraw %}
{% endsnippetcut %}

После того, как значения корректно прописаны и зашифрованы — мы можем пробросить соответствующие значения в Deployment.

{% snippetcut name="deployment.yaml" url="#" %}
{% raw %}
```yaml
       env:
{{ tuple "basicapp" . | include "werf_container_env" | indent 8 }}
        - name: S3ENDPOINT
          value: {{ pluck .Values.global.env .Values.app.s3.epurl | first | default .Values.app.s3.epurl._default | quote }}
        - name: S3KEY
          value: {{ pluck .Values.global.env .Values.app.s3.key | first | default .Values.app.s3.key._default | quote }}
        - name: S3SECRET
          value: {{ pluck .Values.global.env .Values.app.s3.secret | first | default .Values.app.s3.secret._default | quote }}
        - name: S3BUCKET
          value: {{ pluck .Values.global.env .Values.app.s3.bucket | first | default .Values.app.s3.bucket._default | quote }}
```
{% endraw %}
{% endsnippetcut %}

Об особенностях использования aws-java-sdk можно подробно почитать в [документации](https://cloud.spring.io/spring-cloud-aws/spring-cloud-aws.html)

<div>
    <a href="060_email.html" class="nav-btn">Далее: Работа с электронной почтой</a>
</div>
