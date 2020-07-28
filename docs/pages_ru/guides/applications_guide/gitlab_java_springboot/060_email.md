---
title: Работа с электронной почтой
sidebar: applications_guide
permalink: documentation/guides/applications_guide/gitlab_java_springboot/060_email.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/secret-values.yaml
- .helm/values.yaml
- src/main/java/com/example/demo/SendGridClient.java
- pom.xml
{% endfilesused %}

В этой главе мы настроим в нашем базовом приложении работу с почтой.

Для того чтобы использовать почту мы предлагаем лишь один вариант - использовать внешнее API. В нашем примере это [sendgrid](https://sendgrid.com/).

Для того, чтобы Java приложение могло работать с sendgrid необходимо установить и сконфигурировать зависимость `sendgrid` и начать её использовать. Пропишем зависимости в `pom.xml`, чтобы они устаналивались:

{% snippetcut name="pom.xml" url="#" %}
{% raw %}
```xml
dependencies {
  ...
  implementation 'com.sendgrid:sendgrid-java:4.5.0'
}

repositories {
  mavenCentral()
}
```
{% endraw %}
{% endsnippetcut %}

В коде приложения подключение к API и отправка сообщения может выглядеть так:

{% snippetcut name="SendGridClient.java" url="#" %}
{% raw %}
```java
@Service
public class SendGridEmailService implements EmailService {
    private SendGrid sendGridClient;
    @Autowired
    public SendGridEmailService(SendGrid sendGridClient) {
        this.sendGridClient = sendGridClient;
    }
    @Override
    public void sendText(String from, String to, String subject, String body) {
        Response response = sendEmail(from, to, subject, new Content("text/plain", body));
        System.out.println("Status Code: " + response.getStatusCode() + ", Body: " + response.getBody() + ", Headers: "
                + response.getHeaders());
    }
    @Override
    public void sendHTML(String from, String to, String subject, String body) {
        Response response = sendEmail(from, to, subject, new Content("text/html", body));
        System.out.println("Status Code: " + response.getStatusCode() + ", Body: " + response.getBody() + ", Headers: "
                + response.getHeaders());
    }
    private Response sendEmail(String from, String to, String subject, Content content) {
        Mail mail = new Mail(new Email(from), subject, new Email(to), content);
        mail.setReplyTo(new Email("abc@gmail.com"));
        Request request = new Request();
        Response response = null;
        try {
            request.setMethod(Method.POST);
            request.setEndpoint("mail/send");
            request.setBody(mail.build());
            this.sendGridClient.api(request);
        } catch (IOException ex) {
            System.out.println(ex.getMessage());
        }
        return response;
    }
}
```
{% endraw %}
{% endsnippetcut %}

Для работы с `sendgrid` необходимо пробросить в ключи доступа в приложение. Для этого стоит использовать [механизм секретных переменных](https://ru.werf.io/documentation/reference/deploy_process/working_with_secrets.html). *Вопрос работы с секретными переменными рассматривался подробнее, [когда мы делали базовое приложение](020_basic.html#secret-values-yaml)*

{% snippetcut name="secret-values.yaml (расшифрованный)" url="#" %}
{% raw %}
```yaml
app:
  sendgrid:
    apikey: 
      _default: sendgridapikey
    password:
      _default: sendgridpassword
```
{% endraw %}
{% endsnippetcut %}

А не секретные значения — храним в `values.yaml`

{% snippetcut name="values.yaml" url="#" %}
{% raw %}
```yaml
app:
  sendgrid:
    username:
      _default: sendgridusername
```
{% endraw %}
{% endsnippetcut %}

После того, как значения корректно прописаны и зашифрованы — мы можем пробросить соответствующие значения в Deployment.

{% snippetcut name="deployment.yaml" url="#" %}
{% raw %}
```yaml
        - name: SGAPIKEY
          value: {{ pluck .Values.global.env .Values.app.sendgrid.apikey | first | default .Values.app.sendgrid.apikey._default | quote }} 
        - name: SGUSERNAME
          value: {{ pluck .Values.global.env .Values.app.sendgrid.username | first | default .Values.app.sendgrid.username._default | quote }}
        - name: SGPASSWORD
          value: {{ pluck .Values.global.env .Values.app.sendgrid.password | first | default .Values.app.sendgrid.password._default | quote }}
```
{% endraw %}
{% endsnippetcut %}

В интернете можно найти много разных примеров настройки почты через `sendgrid` используя spring. Имплементация может отличаться, но нам важно понять, что нужно параметризировать `application.properties`, чтобы java узнавала о значения из переменных окружения уже во время выполнения в кластере, а не на этапе сборки.

<div>
    <a href="070_redis.html" class="nav-btn">Далее: Подключаем redis</a>
</div>
