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

Работа с электронной почтой производится с помощью внешнего api например mailgun.
В django есть отдельный пакет для работы с ним.
Устанавливаем его: `pip install django-mailgun`

Далее, в настройках необходимо указать параметры подключения к MAILGUN, и после им можно пользоваться:
```yaml
EMAIL_BACKEND = 'django_mailgun.MailgunBackend'
MAILGUN_ACCESS_KEY = 'ACCESS-KEY'
MAILGUN_SERVER_NAME = 'SERVER-NAME'
```
Более подробно все описано в оффициальной документации пакета - https://pypi.org/project/django-mailgun/



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
        - name: MAILGUN_SERVER_NAME
          value: {{ pluck .Values.global.env .Values.app.mailgun_domain | first | default .Values.app.mailgun_domain._default | quote }}
```


