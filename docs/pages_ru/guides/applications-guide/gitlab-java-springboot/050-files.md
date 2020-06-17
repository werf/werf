---
title: Работа с файлами
sidebar: applications-guide
permalink: documentation/guides/applications-guide/template/050-files.html
layout: guide
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/secret-values.yaml
- config/initializers/aws.rb
- Gemfile
{% endfilesused %}



Секция про ассеты подводит нас к другому вопросу - как мы можем сохранять файлы в условиях kubernetes? Нужно учитывать то, что наше приложение в любой момент времени может быть перезапущено kubernetes (и это нормально). Да, kubernetes умеет работать с сетевыми файловыми системами (EFS, NFS, к примеру), которые позволяют подключать общую директорию ко многим подам одновременно. Но правильный путь - использовать s3 для хранения файлов. 
Тем более, что в нашем случае его подключить достаточно просто.
Нужно прописать в pom.xml использование aws-java-sdk.

```xml
<dependency>
   <groupId>com.amazonaws</groupId>
   <artifactId>aws-java-sdk</artifactId>
   <version>1.11.133</version>
</dependency>
```

[pom.xml](gitlab-java-springboot-files/02-demo-with-assets/pom.xml:27-31)

Затем добавить в файл properties приложения - в нашем случае `src/main/resources/application.properties` - сопоставление перменных java с переменными окружения:

```yaml
amazonProperties:
  endpointUrl: ${S3ENDPOINT}
  accessKey: ${S3KEY}
  secretKey: ${S3SECRET}
  bucketName: ${S3BUCKET}
```

[application.properties](gitlab-java-springboot-files/02-demo-with-assets/src/main/resources/application.properties)

Сами доступы нужно записать в values.yaml и secret-values.yaml (используя `werf helm secret values edit .helm/secret-values.yaml` как было описано в главе Секретные переменные). Например:

```yaml
app:
  s3:
    epurl:
      _default: https://s3.us-east-2.amazonaws.com
    bucket:
      _default: mydefaultbucket
      production: myproductionbucket
```

[values.yaml](gitlab-java-springboot-files/02-demo-with-assets/.helm/values.yaml:8-13)

и в secret-values:

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

При звкрытии редактора значения [зашифруются](gitlab-java-springboot-files/02-demo-with-assets/.helm/secret-values.yaml).

Чтобы пробросить эти переменные в контейнер нужно в разделе env deployment их отдельно объявить. Наприме:

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

[deployment.yaml](gitlab-java-springboot-files/02-demo-with-assets/.helm/templates/10-deployment.yaml:53-60)

И мы, в зависимости от используемого окружения, можем подставлять нужные нам значения.

