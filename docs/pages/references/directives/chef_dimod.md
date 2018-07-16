---
title: Выделение модулей (chef only)
sidebar: reference
permalink: chef_dimod.html
folder: advanced_build
---

Dapp is able to use same snippets of code in multiple repos.

Поддержка сборщиками

- Shell - нет
- Chef - да
- Ansible - нет

## Проблематика

В процессе описания правил сборки образов приложений неизбежно образуются общие части, которые необходимо повторять в нескольких приложениях. В chef существует механизм cookbook'ов для решения этой проблемы. Для простоты интеграции с правилами сборки, описываемыми в Dappfile, и для учета особенностей влияния правил сборки на кэширование собираемых образов в dapp сделаны модули chef сборщика на основе chef cookbook'ов, называемые dimod.

## Решение

Dimod — это отдельный chef cookbook, оформленный специальным образом. В названии cookbook'а dimod'а обязателен префикс 'dimod-'. Файловая структура dimod cookbook'а похожа на структуру директории .dapp\_chef, но имеет свои особенности:

* metadata.rb — обязательно присутствует, описывает имя cookbook'а, версию и его зависимости от других cookbook'ов.
* recipes
  * \<стадия\>.rb
* files
  * \<стадия\>
    * \<имя файла\>
  * common
    * \<имя файла\>
* templates
  * \<стадия\>
    * \<имя файла шаблона\>
  * common
    * \<имя файла шаблона\>
* attributes
  * \<стадия\>.rb
  * common.rb

В директории common лежат файлы и шаблоны, доступные на любой стадии сборки образа. Файлы и шаблоны из директории с именем стадии доступны только указанной стадии.

Технически сборщик устроен так, что при сборки стадии файлы из директории common и директории текущей стадии совмещаются в единую директорию `{files|templates}/default`. Поэтому указывать в рецептах параметр source для ресурсов cookbook\_file и template для простейших случаев не обязательно. В директории common и в директориях стадий не могут быть файлы с одинаковыми именами — это приведет к ошибке сборки.

Атрибуты можно определить для конкретной стадии в файле `attributes/<стадия>.rb` или для всех стадий в `attributes/common.rb`.

***Примечание***. Если в dimod не объявлено ни одного рецепта, но объявлены атрибуты — эти атрибуты будут доступны при сборке приложения. Технически в случае отсутствия рецептов в cookbook'е добавляется автоматический фиктивный рецепт, за счет которого в chef становятся доступны атрибуты cookbook'а.

### Установка ruby через dimod-example-ruby

В [сборке с помощью chef](chef_for_advanced_build.html) был приведен [пример](https://github.com/flant/dapp-example-chef-for-advanced-build-1.git), в котором происходит установка rvm и ruby в образ. Вынесем установку ruby в dimod, т.к. этот кусок правил сборки образа нужен для каждого приложения, использующего ruby.

Рассмотрим подготовленный репозиторий с кодом dimod-example-ruby:

```shell
git clone https://github.com/flant/dimod-example-ruby
cd dimod-example-ruby
```

Dimod может иметь параметры, которые пользователь задает при подключении этого dimod'а в Dappfile. Технически для этого используются атрибуты. В dimod-example-ruby параметр с устанавливаемой версией ruby хранится в атрибуте `dimod-example-ruby.ruby_version`. В файле `attributes/before_install.rb` определяется значение по умолчанию:

```ruby
default['dimod-example-ruby']['ruby_version'] = '2.2.4'
```

Хорошая практика написания dimod'ов предполагает использование файлов атрибутов в dimod'ах для декларирования всех возможных атрибутов для настройки этого dimod. Если значения по умолчанию не нужны — можно создать файл атрибутов с закомментированными примерами назначениями возможных атрибутов.

В первом примере dapp-example-chef-for-advanced-build-1 установка rvm и ruby происходила на стадии before\_install, поэтому для dimod заведен рецепт `recipes/before_install.rb`:

```ruby
include_recipe 'apt'

node.default['rvm']['gpg'] = {}
node.default['rvm']['install_rubies'] = true
node.default['rvm']['rubies'] = [node['dimod-example-ruby']['ruby_version']]
node.default['rvm']['default_ruby'] = node['dimod-example-ruby']['ruby_version']
node.default['rvm']['global_gems'] = [{name: 'bundler'}]
include_recipe 'rvm::system'
```

Использование рецепта apt в данном случае переехало в dimod, чтобы не усложнять конфигурацию. Данный рецепт не помешает использованию dimod-example-ruby для centos (особенность реализации cookbook'а apt).

### Использование dimod-example-ruby в проекте

Переделаем пример dapp-example-chef-for-advanced-build-1 на использование dimod-example-ruby:

```shell
git clone https://github.com/flant/dapp-example-chef-for-advanced-build-1.git
cd dapp-example-chef-for-advanced-build-1
```

Dappfile с подключением dimod-example-ruby:

```ruby
dimg do
  docker.from 'ubuntu:16.04'

  git.add('/').to('/app')
  docker.workdir '/app'

  docker.cmd ['/bin/bash', '-lec', 'bundle exec ruby app.rb']
  docker.expose 4567

  chef do
    attributes['dimod-example-ruby']['ruby_version'] = '2.3.1'
    dimod 'dimod-example-ruby', github: 'flant/dimod-example-ruby'

    recipe 'bundle_gems'
    recipe 'app_config'
  end
end
```

В dimod-example-ruby по умолчанию устанавливается версия ruby 2.2.4, предположим для нашего приложения требуется версия 2.3.1. Через директиву `chef.attributes` установлен соответствующий атрибут. [Подробнее об определении атрибутов](chef_for_advanced_build.html#атрибуты).

Рецепты, добавляемые через директиву dimod будут запускаться перед рецептами приложения в том порядке, в котором подключаются dimod'ы. В итоге формируется последовательность рецептов в runlist: сначала все dimod'ы, затем все рецепты основного приложения. На данный момент нет поддержки запуска рецептов приложения перед рецептами из dimod'ов — но такая возможность рассматривается: [https://github.com/flant/dapp/issues/177](https://github.com/flant/dapp/issues/177).

Сборка приложения, запуск контейнера и проверка работоспособности (старый контейнер dapp-example-chef-for-advanced-build-1 требуется удалить, чтобы освободить tcp порт):

```shell
$ dapp dimg build
Before install ...                                                          [OK] 677.09 sec
Git artifacts: create archive ...                                           [OK] 2.57 sec
Install group
  Git artifacts: apply patches (before install) ...                         [OK] 2.17 sec
  Install ...                                                               [OK] 29.0 sec
  Git artifacts: apply patches (after install) ...                          [OK] 1.96 sec
Setup group
  Git artifacts: apply patches (before setup) ...                           [OK] 1.93 sec
  Setup ...                                                                 [OK] 12.21 sec
  Git artifacts: apply patches (after setup) ...                            [OK] 1.97 sec
Docker instructions ...                                                     [OK] 2.03 sec
$ docker rm -f dapp-example-chef-for-advanced-build-1
$ dapp dimg run --detach -p 4567:4567 --name dapp-example-chef-for-advanced-build-2
622470f874b0497c15ea0a796267ee01203db37dc828e75c6330b7deaadf5316
$ curl localhost:4567/message
Hello from setup/app_config.rb recipe
```

Проверим что в образе используется версия ruby 2.3.1:

```shell
$ docker exec -ti dapp-example-chef-for-advanced-build-2 bash -l
root@622470f874b0:/app# ruby --version
ruby 2.3.1p112 (2016-04-26 revision 54768) [x86_64-linux]
```

### Добавление локального web-фронтенда в приложение
а
Допустим, появилась частая задача добавлять в web-приложение локальный nginx-фронтенд. Создадим новый dimod-example-local-nginx с простыми настройками, который позволит быстро подключать фронтенд для подобных приложений.

Задачи dimod — установка пакета nginx и генерация его конфига.

Задаем в атрибутах устанавливаемый пакет nginx. Пакеты как правило устанавливаются на стадии before\_install, чтобы закэшировать их надолго. Поэтому данные настройки будут нужны в before\_install атрибутах в файле `attributes/before_install.rb`:

```ruby
default['dimod-example-local-nginx']['nginx_package'] = 'nginx'
# default['dimod-example-local-nginx']['nginx_package_version'] = '1.10.0-0ubuntu0.16.04.4'
```

Рецепт осуществляющий установку пакета в файле `recipes/before_install.rb`:

```ruby
package node['dimod-example-local-nginx']['nginx_package'] do
  package_version = node['dimod-example-local-nginx']['nginx_package_version']
  version(package_version) if package_version
end
```

Порт куда nginx будет проксировать задается атрибутом `dimod-example-local-nginx.proxy\_to\_port`. Генерацию конфига для nginx стоит делать одним из последних шагов сборки образа, т.к. это быстрый процесс. Например, на стадии setup — последней пользовательской стадии из предоставляемых dapp стадий.

Объявляем поддерживаемые атрибуты в файле `attributes/setup.rb`:

```ruby
default['dimod-example-local-nginx']['proxy_to_port'] = 8080
```

Шаблон конфига для nginx из `templates/setup/nginx.conf.erb`:

```
server {
    listen 80 default_server;
    listen [::]:80 default_server;

    location / {
        gzip_types *;

        proxy_redirect    off;
        proxy_set_header  Host              $http_host;
        proxy_set_header  X-Real-IP         $remote_addr;
        proxy_set_header  X-Forwarded-For   $proxy_add_x_forwarded_for;

        proxy_buffering on;
        proxy_buffers 64 128k;
        proxy_buffer_size 4m;
        proxy_busy_buffers_size 4m;

        proxy_pass http://localhost:<%= @proxy_to_port %>/;
    }
}
```

Рецепт для конфигурации nginx из файла `recipes/setup.rb`

```ruby
file '/etc/nginx/sites-enabled/default' do
  action :delete
end

template '/etc/nginx/sites-enabled/app.conf' do
  source 'nginx.conf.erb'

  proxy_to_port = node['dimod-example-local-nginx']['proxy_to_port']
  raise "dimod-example-local-nginx.proxy_to_port attribute required" unless proxy_to_port
  variables(proxy_to_port: proxy_to_port)

  action :create
end
```

В итоге после подключения данного dimod в образе будет nginx и конфиг.

Следующим шагом добавляем nginx в команду запуска контейнера и выставляем 80 порт для наружного использования контейнера:

```ruby
dimg do
  docker.from 'ubuntu:16.04'

  git.add('/').to('/app')
  docker.workdir '/app'

  docker.cmd ['/bin/bash', '-lec', 'nginx && bundle exec ruby app.rb']
  docker.expose 80

  chef do
    attributes['dimod-example-ruby']['ruby_version'] = '2.3.1'
    dimod 'dimod-example-ruby', github: 'flant/dimod-example-ruby'

    attributes['dimod-example-local-nginx']['proxy_to_port'] = '4567'
    dimod 'dimod-example-local-nginx', github: 'flant/dimod-example-local-nginx'

    recipe 'bundle_gems'
    recipe 'app_config'
  end
end
```

Пересобираем образ, меняем порт 4567 на 80 при запуске контейнера и проверяем результат:

```shell
$ dapp dimg build
Before install ...                                                          [OK] 345.89 sec
Git artifacts: create archive ...                                           [OK] 1.36 sec
Install group
  Git artifacts: apply patches (before install) ...                         [OK] 1.3 sec
  Install ...                                                               [OK] 12.23 sec
  Git artifacts: apply patches (after install) ...                          [OK] 1.38 sec
Setup group
  Git artifacts: apply patches (before setup) ...                           [OK] 1.34 sec
  Setup ...                                                                 [OK] 5.31 sec
  Git artifacts: apply patches (after setup) ...                            [OK] 1.33 sec
Docker instructions ...                                                     [OK] 1.61 sec
$ docker rm -f dapp-example-chef-for-advanced-build-2
$ dapp dimg run --detach -p 80:80 --name dapp-example-chef-for-advanced-build-2
b47ff0786013132963a9cea7a6a2917ee309b274a9379a6c9b4540092962ba4e
$ curl localhost/message
Hello from setup/app_config.rb recipe
```

### Пересборка при изменении атрибутов

Изменение атрибутов, установленных через директиву `chef.attributes`, ведет к пересборке образа со стадии before\_install. Чтобы оптимизировать процесс сборки, если атрибуты используются только на определенной стадии, dapp поддерживает определение атрибутов для конкретной стадии.

В нашем примере атрибут `dimod-example-local-nginx.proxy_to_port` установлен директивой `chef.attributes`, однако используется только на стадии setup. Если попытаться изменить значение proxy\_to\_port и перезапустить сборку — произойдет пересборка всех пользовательских стадий, начиная с before\_install, и это займет долгое время. Исправим установку этого атрибута на использование директивы `chef._setup_attributes` и пересоберем приложение:

```ruby
dimg do
  docker.from 'ubuntu:16.04'

  git.add('/').to('/app')
  docker.workdir '/app'

  docker.cmd ['/bin/bash', '-lec', 'nginx && bundle exec ruby app.rb']
  docker.expose 80

  chef do
    attributes['dimod-example-ruby']['ruby_version'] = '2.3.1'
    dimod 'dimod-example-ruby', github: 'flant/dimod-example-ruby'

    _setup_attributes['dimod-example-local-nginx']['proxy_to_port'] = '4567'
    dimod 'dimod-example-local-nginx', github: 'flant/dimod-example-local-nginx'

    recipe 'bundle_gems'
    recipe 'app_config'
  end
end
```

```shell
$ dapp dimg build
Before install ...                                                          [OK] 351.04 sec
Git artifacts: create archive ...                                           [OK] 1.36 sec
Install group
  Git artifacts: apply patches (before install) ...                         [OK] 2.08 sec
  Install ...                                                               [OK] 12.95 sec
  Git artifacts: apply patches (after install) ...                          [OK] 1.39 sec
Setup group
  Git artifacts: apply patches (before setup) ...                           [OK] 1.43 sec
  Setup ...                                                                 [OK] 6.61 sec
  Git artifacts: apply patches (after setup) ...                            [OK] 1.38 sec
Docker instructions ...                                                     [OK] 1.43 sec
```

При данном изменении произошла полная пересборка, т.к. был удален глобальный атрибут proxy\_to\_port и добавлен атрибут специально для стадии setup. Изменим значение атрибута proxy\_to\_port на 9000 и запустим пересборку:

```ruby
dimg do
  docker.from 'ubuntu:16.04'

  git.add('/').to('/app')
  docker.workdir '/app'

  docker.cmd ['/bin/bash', '-lec', 'nginx && bundle exec ruby app.rb']
  docker.expose 80

  chef do
    attributes['dimod-example-ruby']['ruby_version'] = '2.3.1'
    dimod 'dimod-example-ruby', github: 'flant/dimod-example-ruby'

    _setup_attributes['dimod-example-local-nginx']['proxy_to_port'] = '9000'
    dimod 'dimod-example-local-nginx', github: 'flant/dimod-example-local-nginx'

    recipe 'bundle_gems'
    recipe 'app_config'
  end
end
```

```shell
$ dapp dimg build
Setup group
  Git artifacts: apply patches (before setup) ...                           [OK] 1.31 sec
  Setup ...                                                                 [OK] 4.65 sec
  Git artifacts: apply patches (after setup) ...                            [OK] 1.32 sec
Docker instructions ...                                                     [OK] 1.37 sec
```

Видим, что сборка произошла со стадии setup. Подобное изменение может значительно сократить время сборки для тех случаев, когда параметры-атрибуты требуется менять. Однако по умолчанию рекомендуется использовать директиву `chef.attributes` для любых атрибутов, а оптимизировать только по надобности.

### Дальнейшие действия

В данном контейнере не используется никакого supervisor'а — следующим шагом было бы логично написать dimod-example-supervisor с настройкой через атрибуты, который бы устанавливал supervisor и генерировал для него конфиги для указанных приложений. Эта задача остается в качестве практики читателям.

Итоговый пример со всеми подключенными dimod'ами можно найти тут: [https://github.com/flant/dapp-example-chef-for-advanced-build-2](https://github.com/flant/dapp-example-chef-for-advanced-build-2).
