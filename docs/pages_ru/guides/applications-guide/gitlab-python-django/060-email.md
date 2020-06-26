---
title: Работа с электронной почтой
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-python-django/060-email.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/secret-values.yaml
- TODO название файла
{% endfilesused %}

В этой главе мы настроим в нашем базовом приложении работу с почтой.

Для того чтобы использовать почту мы предлагаем лишь один вариант - использовать внешнее API. В нашем примере это [mailgun](https://www.mailgun.com/).

{% offtopic title="А почему бы просто не установить sendmail?" %}
TODO: ответить на этот непростой вопрос
{% endofftopic %}

Для того, чтобы Python приложение могло работать с mailgun необходимо установить и сконфигурировать зависимость и начать её использовать. Установим через `pip` зависимость:

```bash
pip install django-mailgun
```

И [сконфигурируем согласно документации](https://pypi.org/project/django-mailgun/) пакета

{% snippetcut name="settings.py" url="#" %}
{% raw %}
```python
EMAIL_BACKEND = 'django_mailgun.MailgunBackend'
MAILGUN_ACCESS_KEY = 'ACCESS-KEY'
MAILGUN_SERVER_NAME = 'SERVER-NAME'
```
{% endraw %}
{% endsnippetcut %}

TODO: ^^^^^ выглядит как какая-то чушь. Нужно вписать реальные значения в конфигурирование.

В коде приложения подключение к API и отправка сообщения может выглядеть так:

{% snippetcut name="____________" url="#" %}
{% raw %}
```____________
____________
____________
____________
```
{% endraw %}
{% endsnippetcut %}

Для работы с mailgun необходимо пробросить в ключи доступа в приложение. Для этого стоит использовать [механизм секретных переменных](#######TODO). *Вопрос работы с секретными переменными рассматривался подробнее, [когда мы делали базовое приложение](020-basic.html#secret-values-yaml)*

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

TODO: надо дать отсылку на какой-то гайд, где описано, как конкретно использовать ____________. Мало же просто его установить — надо ещё как-то юзать в коде.


<div>
    <a href="070-redis.html" class="nav-btn">Далее: Подключаем redis</a>
</div>
