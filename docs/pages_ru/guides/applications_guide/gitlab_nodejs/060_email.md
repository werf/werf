---
title: Работа с электронной почтой
sidebar: applications_guide
permalink: documentation/guides/applications_guide/gitlab_nodejs/060_email.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/secret-values.yaml
- consumer.js
{% endfilesused %}

В этой главе мы настроим в нашем базовом приложении работу с почтой.

Для того чтобы использовать почту мы предлагаем лишь один вариант - использовать внешнее API. В нашем примере это [mailgun](https://www.mailgun.com/).

Для того, чтобы NodeJS приложение могло работать с mailgun необходимо установить и сконфигурировать зависимость и начать её использовать. Установим через `npm` зависимость:

```bash
npm install mailgun-js
```

И [сконфигурируем согласно документации](https://github.com/mailgun/mailgun-js#documentation) пакета

{% snippetcut name="consumer.js" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/060-email/consumer/consumer.js" %}
{% raw %}
const mailgun = require("mailgun-js");
...
const mg = mailgun({apiKey: process.env.MAILGUN_APIKEY, domain: process.env.MAILGUN_DOMAIN, host: "api.eu.mailgun.net"});
```
{% endraw %}
{% endsnippetcut %}

В коде приложения подключение к API и отправка сообщения может выглядеть так:

{% snippetcut name="consumer.js" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/060-email/consumer/consumer.js" %}
{% raw %}
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
{% endraw %}
{% endsnippetcut %}

Для работы с mailgun необходимо пробросить в ключи доступа в приложение. Для этого стоит использовать [механизм секретных переменных](https://ru.werf.io/documentation/reference/deploy_process/working_with_secrets.html). *Вопрос работы с секретными переменными рассматривался подробнее, [когда мы делали базовое приложение](020_basic.html#secret-values-yaml)*

{% snippetcut name="secret-values.yaml (расшифрованный)" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/060-email/.helm/secret-values.yaml" %}
{% raw %}
```yaml
app:
  mailgun_apikey:
    _default: 192edaae18f13aaf120a66a4fefd5c4d-7fsaaa4e-kk5d08a5
  mailgun_domain:
    _default: sandboxf1b90123966447a0514easd0ea421rba.mailgun.org
```
{% endraw %}
{% endsnippetcut %}

После того, как значения корректно прописаны и зашифрованы — мы можем пробросить соответствующие значения в Deployment.

{% snippetcut name="deployment.yaml" url="https://github.com/werf/demos/blob/master/applications-guide/gitlab-nodejs/examples/060-email/.helm/templates/deployment.yaml" %}
{% raw %}
```yaml
        - name: MAILGUN_APIKEY
          value: {{ pluck .Values.global.env .Values.app.mailgun_apikey | first | default .Values.app.mailgun_apikey._default }}
        - name: MAILGUN_DOMAIN
          value: {{ pluck .Values.global.env .Values.app.mailgun_domain | first | default .Values.app.mailgun_domain._default | quote }}
```
{% endraw %}
{% endsnippetcut %}

<div>
    <a href="070_redis.html" class="nav-btn">Далее: Подключаем redis</a>
</div>
