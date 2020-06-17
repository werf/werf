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



Для того чтобы использовать почту мы предлагаем лишь один вариант - использовать внешнее API. В нашем примере это mailgun.
Внутри исходного кода подключение к API и отправка сообщения может выглядеть так:
```js
function sendMessage(message) {
  try {
    const mg = mailgun({apiKey: process.env.MAILGUN_APIKEY, domain: process.env.MAILGUN_DOMAIN});
    const email = JSON.parse(message.content.toString());
    email.from = "Mailgun Sandbox <postmaster@sandbox"+process.env.MAILGUN_FROM+">",
    email.subject = "Welcome to Chat!"
    mg.messages().send(email, function (error, body) {
      console.log(body);
    });
  } catch (error) {
    console.error(error)
  }
}

```
Главное заметить, что мы также как и в остальных случаях выносим основную конфигурацию в переменные. А далее мы по тому же принципу добавляем их параметризированно в наш secret-values.yaml не забыв использовать [шифрование](####секретные-переменные).
```yaml
  mailgun_apikey:
    _default: 192edaae18f13aaf120a66a4fefd5c4d-7fsaaa4e-kk5d08a5
  mailgun_domain:
    _default: sandboxf1b90123966447a0514easd0ea421rba.mailgun.org
```
И теперь мы можем использовать их внутри манифеста.
```yaml
        - name: MAILGUN_APIKEY
          value: {{ pluck .Values.global.env .Values.app.mailgun_apikey | first | default .Values.app.mailgun_apikey._default }}
        - name: MAILGUN_DOMAIN
          value: {{ pluck .Values.global.env .Values.app.mailgun_domain | first | default .Values.app.mailgun_domain._default | quote }}
```

