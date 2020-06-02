---
title: Гайд по использованию Ruby On Rails + GitLab + Werf
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-rails/060-email.html
author: alexey.chazov <alexey.chazov@flant.com>
layout: guide
toc: false
author_team: "bravo"
author_name: "alexey.chazov"
ci: "gitlab"
language: "ruby"
framework: "rails"
is_compiled: 0
package_managers_possible:
 - bundler
package_managers_chosen: "bundler"
unit_tests_possible:
 - Rspec
unit_tests_chosen: "Rspec"
assets_generator_possible:
 - webpack
 - gulp
assets_generator_chosen: "webpack"
---

# Работа с электронной почтой

Для того чтобы использовать почту мы предлагаем лишь один вариант - использовать внешнее API. В нашем примере это mailgun.
Внутри исходного кода подключение к API и отправка сообщения может выглядеть так:

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

Главное заметить, что мы также как и в остальных случаях выносим основную конфигурацию в переменные. А далее мы по тому же принципу добавляем их параметризированно в наш secret-values.yaml не забыв использовать [шифрование](####секретные-переменные).

```yaml
app:
  mailgun_apikey:
    _default: 192edaae18f13aaf120a66a4fefd5c4d-7fsaaa4e-kk5d08a5
```

И теперь мы можем использовать их внутри манифеста.
```yaml
        - name: MAILGUN_APIKEY
          value: {{ pluck .Values.global.env .Values.app.mailgun_apikey | first | default .Values.app.mailgun_apikey._default }}
```
