---
title: Работа с электронной почтой
sidebar: applications-guide
permalink: documentation/guides/applications-guide/template/060-email.html
layout: guide
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/secret-values.yaml
- TODO название файла
{% endfilesused %}

На наш взгляд самым правильным способом отправки email-сообщений будет внешнее api - провайдер для почты. Например sendgrid, mailgun, amazon ses и подобные.

Рассмотрим на примере sendgrid. spring умеет с ним работать, даже есть [автоконфигуратор](https://docs.spring.io/spring-boot/docs/current/api/org/springframework/boot/autoconfigure/sendgrid/SendGridAutoConfiguration.html). Для этого нужно использовать [библиотеку sendgrid для java](https://github.com/sendgrid/sendgrid-java)
Включим библиотеку в pom.xml, согласно документации:
```xml
...
dependencies {
  ...
  implementation 'com.sendgrid:sendgrid-java:4.5.0'
}

repositories {
  mavenCentral()
}
```

Доступы к sendgrid, как и в случае с s3 пропишем в .helm/values.yaml, пробросим их в виде переменных окружения в наше приложение через deployment, а в коде (в application.properties) сопоставим [java-переменные](https://docs.spring.io/spring-boot/docs/current/reference/html/appendix-application-properties.html) используемые в коде и переменные окружения:

```
# SENDGRID (SendGridAutoConfiguration)
spring.sendgrid.api-key= ${SGAPIKEY}
spring.sendgrid.username= ${SGUSERNAME}
spring.sendgrid.password= ${SGPASSWORD}
spring.sendgrid.proxy.host= ${SGPROXYHOST} #optional
```

Теперь можем использовать эти данные в приложении для отправки почты.

