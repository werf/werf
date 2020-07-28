---
title: Работа с электронной почтой
sidebar: applications_guide
permalink: documentation/guides/applications_guide/gitlab_python_django/060_email.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/secret-values.yaml
- settings.py
- mail_example.py
{% endfilesused %}

В этой главе мы настроим в нашем базовом приложении работу с почтой.

Для того чтобы использовать почту мы предлагаем лишь один вариант - использовать внешнее API. В нашем примере это [mailgun](https://www.mailgun.com/).

Для того, чтобы Django приложение могло работать с mailgun необходимо установить и сконфигурировать зависимость и начать её использовать. Установим через `pip` зависимость:

```bash
pip install django-anymail
```

И [сконфигурируем согласно документации](https://github.com/anymail/django-anymail) пакета

{% snippetcut name="settings.py" url="#" %}
{% raw %}
```python
INSTALLED_APPS = [
    # ...
    "anymail",
    # ...
]

ANYMAIL = {
    "MAILGUN_API_KEY": "192edaae18f13aaf120a66a4fefd5c4d-7fsaaa4e-kk5d08a5",
    "MAILGUN_SENDER_DOMAIN": 'https://api.mailgun.net/v3/domain.io',
}
EMAIL_BACKEND = "anymail.backends.mailgun.EmailBackend"
DEFAULT_FROM_EMAIL = "you@domain.io"
SERVER_EMAIL = "your-server@domain.io"
```
{% endraw %}
{% endsnippetcut %}

В коде приложения подключение к API и отправка сообщения может выглядеть так:

{% snippetcut name="mail_example.py" url="#" %}
{% raw %}
```python
from django.core.mail import send_mail

send_mail("It works!", "This will get sent through Mailgun",
          "Anymail Sender <from@domain.io", ["to@domain.io"])
```
{% endraw %}
{% endsnippetcut %}

Для работы с mailgun необходимо пробросить в ключи доступа в приложение. Для этого стоит использовать [механизм секретных переменных](https://ru.werf.io/documentation/reference/deploy_process/working_with_secrets.html). *Вопрос работы с секретными переменными рассматривался подробнее, [когда мы делали базовое приложение](020_basic.html#secret-values-yaml)*

{% snippetcut name="secret-values.yaml (расшифрованный)" url="#" %}
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

{% snippetcut name="deployment.yaml" url="#" %}
{% raw %}
```yaml
        - name: MAILGUN_APIKEY
          value: {{ pluck .Values.global.env .Values.app.mailgun_apikey | first | default .Values.app.mailgun_apikey._default }}
        - name: MAILGUN_SERVER_NAME
          value: {{ pluck .Values.global.env .Values.app.mailgun_domain | first | default .Values.app.mailgun_domain._default | quote }}
```
{% endraw %}
{% endsnippetcut %}

Как использовать данный пакет для отправки почты более подробно можно почитать в [статье](https://medium.com/@9cv9official/sending-html-email-in-django-with-anymail-7163dc332113).

<div>
    <a href="070_redis.html" class="nav-btn">Далее: Подключаем redis</a>
</div>
