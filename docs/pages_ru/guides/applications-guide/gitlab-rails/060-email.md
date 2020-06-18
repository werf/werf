---
title: Работа с электронной почтой
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-rails/060-email.html
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

Внутри исходного кода подключение к API и отправка сообщения может выглядеть так:

{% snippetcut name="TODO название файла" url="#" %}
```ruby
require 'mailgun-ruby'

# First, instantiate the Mailgun Client with your API key
mg_client = Mailgun::Client.new ENV['MAILGUN_APIKEY']

# Define your message parameters
message_params =  { from: 'bob@sending_domain.com',
                    to:   'sally@example.com',
                    subject: 'The Ruby SDK is awesome!',
                    text:    'It is really easy to send a message!'
                  }

# Send your message through the client
mg_client.send_message 'sending_domain.com', message_params
```
{% endsnippetcut %}

Для работы с mailgun необходимо пробросить в ключи доступа в приложение. Для этого стоит использовать [механизм секретных переменных](#######TODO). *Вопрос работы с секретными переменными рассматривался подробнее, [когда мы делали базовое приложение](020-basic.html#secret-values-yaml)*

{% snippetcut name="secret-values.yaml (расшифрованный)" url="#" %}
```yaml
app:
  mailgun_apikey:
    _default: 192edaae18f13aaf120a66a4fefd5c4d-7fsaaa4e-kk5d08a5
```
{% endsnippetcut %}

После того, как значения корректно прописаны и зашифрованы — мы можем пробросить соответствующие значения в Deployment.

{% snippetcut name="deployment.yaml" url="#" %}
```yaml
        - name: MAILGUN_APIKEY
          value: {{ pluck .Values.global.env .Values.app.mailgun_apikey | first | default .Values.app.mailgun_apikey._default }}
```
{% endsnippetcut %}

TODO: надо дать отсылку на какой-то гайд, где описано, как конкретно использовать гем mailgun-ruby. Мало же просто его установить — надо ещё как-то юзать в коде.


<div>
    <a href="070-redis.html" class="nav-btn">Далее: Подключаем redis</a>
</div>
