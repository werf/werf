---
title: Работа с файлами
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

В этой главе мы настроим в нашем базовом приложении работу с пользовательскими файлами. Для этого нам нужно персистентное хранилище.

В идеале — добиться, чтобы приложение было stateless, а данные хранились в S3-совместимом хранилище, например minio или aws s3. Это обеспечивает простое масштабирование, работу в HA режиме и высокую доступность.

{% offtopic title="А есть какие-то способы кроме S3?" %}
TODO: пишем про 
Первый и более общий способ. Это использовать как volume в подах [NFS](https://kubernetes.io/docs/concepts/storage/volumes/#nfs), [CephFS](https://kubernetes.io/docs/concepts/storage/volumes/#cephfs) или [hostPath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath), который будет направлен на директорию на ноде, куда будет подключено одно из сетевых хранилищ.

Мы не рекомендуем этот способ, потому что при возникновении неполадок с такими типами volume’ов мы будем влиять на работоспособность всего докера, контейнера и демона docker в целом, тем самым могут пострадать приложения которые даже не имеют отношения к вашему приложению.

Если мы будем сохранять файлы какой - либо директории у приложения запущенного в kubernetes - то после перезапуска контейнера все изменения пропадут.
{% endofftopic %}

Данная настройка производится полностью в рамках приложения, а нам остается только передать необходимые переменные окружения при запуске приложения.

Добавим в Gemfile зависимости для работы с s3 aws:

{% snippetcut name="Gemfile" url="#" %}
```
gem 'aws-sdk', '~> 2'
```
{% endsnippetcut %}

Добавим инициализатор TODO: написать логику, почему это так. Что это за файл?

{% snippetcut name="config/initializers/aws.rb" url="#" %}
```ruby
Aws.config.update({
  region: 'us-east-1',
  credentials: Aws::Credentials.new(ENV['AWS_ACCESS_KEY_ID'], ENV['AWS_SECRET_ACCESS_KEY']),
})

S3_BUCKET = Aws::S3::Resource.new.bucket(ENV['S3_BUCKET'])
```
{% endsnippetcut %}

Для работы с S3 необходимо пробросить в ключи доступа в приложение. Для этого стоит использовать [механизм секретных переменных](#######TODO). *Вопрос работы с секретными переменными рассматривался подробнее, [когда мы делали базовое приложение](020-basic.html#secret-values-yaml)*

{% snippetcut name="secret-values.yaml (расшифрованный)" url="#" %}
```yaml
app:
  aws_access_key_id:
    _default: EXAMPLEKVFOOOWWPYA
  aws_secret_access_key:
    _default: exampleBARZHS3sRew8xw5hiGLfroD/b21p2l
  s3_bucket:
    _default: my-s3-development
```
{% endsnippetcut %}

После того, как значения корректно прописаны и зашифрованы — мы можем пробросить соответствующие значения в Deployment.

{% snippetcut name="deployment.yaml" url="#" %}
```yaml
        - name: S3_BUCKET
          value: {{ pluck .Values.global.env .Values.app.s3_bucket | first | default .Values.app.s3_bucket._default }}
        - name: AWS_ACCESS_KEY_ID
          value: {{ pluck .Values.global.env .Values.app.aws_access_key_id | first | default .Values.app.aws_access_key_id._default }}
        - name: AWS_SECRET_ACCESS_KEY
          value: {{ pluck .Values.global.env .Values.app.aws_secret_access_key | first | default .Values.app.aws_secret_access_key._default }}
```
{% endsnippetcut %}

TODO: надо дать отсылку на какой-то гайд, где описано, как конкретно использовать гем aws-sdk. Мало же просто его установить — надо ещё как-то юзать в коде.
