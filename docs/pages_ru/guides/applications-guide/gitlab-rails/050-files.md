---
title: Гайд по использованию Ruby On Rails + GitLab + Werf
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-rails/050-files.html
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

# Работа с файлами

Если нам необходимо сохранять какие то пользовательские данные - нам нужно  персистентное хранилище. Лучше всего для stateless приложений в таком случае использовать S3 совместимое хранилище (например minio или aws s3)

Данная настройка производится полностью в рамках приложения а нам остается только передать необходимые переменные окружения при запуске приложения.

Для работы например с s3 aws - подключим необходимые зависимости и опишем переменные окружения

Добавим в Gemfile
```
gem 'aws-sdk', '~> 2'
```

Добавим инициализатор в файл `config/initializers/aws.rb`

```ruby
Aws.config.update({
  region: 'us-east-1',
  credentials: Aws::Credentials.new(ENV['AWS_ACCESS_KEY_ID'], ENV['AWS_SECRET_ACCESS_KEY']),
})

S3_BUCKET = Aws::S3::Resource.new.bucket(ENV['S3_BUCKET'])
```

Главное заметить, что мы также как и в остальных случаях выносим основную конфигурацию в переменные. А далее мы по тому же принципу добавляем их параметризированно в наш secret-values.yaml не забыв использовать [шифрование](####секретные-переменные).

```yaml
app:
  aws_access_key_id:
    _default: EXAMPLEKVFOOOWWPYA
  aws_secret_access_key:
    _default: exampleBARZHS3sRew8xw5hiGLfroD/b21p2l
  s3_bucket:
    _default: my-s3-development
```

И теперь мы можем использовать их внутри манифеста.
```yaml
        - name: S3_BUCKET
          value: {{ pluck .Values.global.env .Values.app.s3_bucket | first | default .Values.app.s3_bucket._default }}
        - name: AWS_ACCESS_KEY_ID
          value: {{ pluck .Values.global.env .Values.app.aws_access_key_id | first | default .Values.app.aws_access_key_id._default }}
        - name: AWS_SECRET_ACCESS_KEY
          value: {{ pluck .Values.global.env .Values.app.aws_secret_access_key | first | default .Values.app.aws_secret_access_key._default }}
```

Если мы будем сохранять файлы какой - либо директории у приложения запущенного в kubernetes - то после перезапуска контейнера все изменения пропадут.
