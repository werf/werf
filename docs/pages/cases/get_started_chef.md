---
title: Сборка приложения на dapp с помощью Chef
sidebar: doc_sidebar
permalink: get_started_chef.html
---

**Внимание, chef-сборщик более активно не развивается. Переходите на ansible.**

Разберем hello-world приложение на ruby, собираемое с помощью chef-сборщика.

```shell
git clone https://github.com/flant/dapp-example-chef-for-advanced-build-1.git
cd dapp-example-chef-for-advanced-build-1
```

Dappfile приложения:

```ruby
dimg do
  docker.from 'ubuntu:16.04'

  git.add('/').to('/app')
  docker.workdir '/app'

  docker.cmd ['/bin/bash', '-lec', 'bundle exec ruby app.rb']
  docker.expose 4567

  chef do
    cookbook 'apt'
    cookbook 'rvm'

    recipe 'ruby'
    recipe 'bundle_gems'
    recipe 'app_config'
  end
end
```

В образ на базе ubuntu:16.04 добавляется исходный код приложения в директорию `/app`. Все правила сборки для chef в данном случае описаны в блоке директивы `chef`. Конфигурация chef-сборщика в общем виде представляет собой включение рецептов и модулей, определение атрибутов, определение зависимых cookbook'ов в Dappfile и создание рецептов (recipes), шаблонов (templates) и подготовленных заранее файлов (files) в директории `.dapp_chef`.

Для сборки образа включено 3 рецепта: ruby, bundle\_gems и app\_config — директивой `chef.recipe <recipe-name>`.

Один логический рецепт, включенный таким образом через Dappfile, может относиться к нескольким файлам-рецептам для разных стадий сборки.

Например, файл `.dapp_chef/recipes/before_install/ruby.rb` запускается на стадии before\_install, а файл `.dapp_chef/recipes/setup/ruby.rb` — на стадии setup, но оба этих файла включаются одновременно через указание директивы `chef.recipe 'ruby'`. Однако создавать или нет файл рецепта для конкретной стадии решает пользователь. Если для включенного в Dappfile рецепта не нашлось файла в некоторой стадии — этот рецепт просто игнорируется при сборке этой стадии.

Последовательность включения нескольких рецептов в Dappfile определяет последовательность их запуска в рамках одной стадии (опять же, в случае, если файлы рецептов для этой стадии существуют).

Рецепт ruby отвечает за установку rvm, ruby и bundler. Переустановку данного софта не требуется производить часто, и для него не требуется наличие исходного кода приложения в образе. Поэтому логичнее всего запускать рецепт на первичной стадии, в которой еще не добавлен исходный код приложения — на стадии before\_install. Рецепт с именем ruby для стадии before\_install располагается в `.dapp_chef/recipes/before_install/ruby.rb`.

```ruby
include_recipe 'apt'

node.default['rvm']['gpg'] = {}
node.default['rvm']['install_rubies'] = true
node.default['rvm']['rubies'] = ['2.3.1']
node.default['rvm']['default_ruby'] = node['rvm']['rubies'].first
node.default['rvm']['global_gems'] = [{name: 'bundler'}]
include_recipe 'rvm::system'
```

Данный рецепт использует внешние cookbook'и apt и rvm. Для указания внешних зависимостей не требуется создавать Berksfile, Berksfile.lock и metadata.rb. Чтобы эти cookbook'и были доступны, необходимо указать их в Dappfile с помощью директивы `chef.cookbook`. Все параметры директивы полностью совпадают с параметрами директивы `cookbook` из Berksfile. 

Следующий рецепт bundle\_gems используется для установки зависимостей целевого ruby-приложения. Эти зависимости определены в Gemfile и Gemfile.lock, которые располагаются в git-репозитории. Первая пользовательская стадия сборки, в которой доступен описанный в Dappfile git-репозиторий — это стадия install. Рецепт с именем bundle\_gems для стадии install располагается в `.dapp_chef/recipes/install/bundle_gems.rb` и просто запускает bundle install.

```ruby
execute 'install bundle gems' do
  cwd '/app'
  command 'bundle install --deployment --path .vendor'
end
```

Рецепт app\_config отвечает за генерацию конфига приложения. Генерация/установка конфигов как правило происходит на стадии setup. Рецепт с именем app\_config для стадии setup располагается в `.dapp_chef/recipes/setup/app_config.rb`

```ruby
file "/app/config.yml" do
  mode 0644
  action :create
  content YAML.dump(
    'message' => "Hello from setup/app_config.rb recipe\n"
  )
end
```

Целевое приложение представляет собой web-сервер, который отдает сообщение из конфига по запросу /message.

Собираем образ и запускаем контейнер:

```shell
$ dapp dimg build
From ...                                                                              [OK] 1.01 sec
Before install ...                                                                    [OK] 260.06 sec
Git artifacts: create archive ...                                                     [OK] 0.92 sec
Install group
  Git artifacts: apply patches (before install) ...                                   [OK] 0.87 sec
  Install ...                                                                         [OK] 15.95 sec
  Git artifacts: apply patches (after install) ...                                    [OK] 0.97 sec
Setup group
  Git artifacts: apply patches (before setup) ...                                     [OK] 0.98 sec
  Setup ...                                                                           [OK] 3.91 sec
  Git artifacts: apply patches (after setup) ...                                      [OK] 0.98 sec
Docker instructions ...                                                               [OK] 1.89 sec
$ dapp dimg run --detach -p 4567:4567 --name dapp-example-chef-for-advanced-build-1
f89d8357dd4a9c79076e741a5713a4147f71516651a58318fa9269b4b0f48172
```

Проверяем работоспособность приложения:

```shell
$ curl localhost:4567/message
Hello from setup/app_config.rb recipe
```

Разделение рецептов и файлов cookbook'а по стадиям в текущей реализации dapp открывает возможность беспроблемного кэширования стадий сборки. Изменение рецептов, шаблонов или файлов связанных со стадией ведет не к полной пересборке, а к пересборке начиная с этой стадии. Поменяем поле конфигурации message в рецепте, который генерирует конфигурационный файл, `.dapp_chef/recipes/setup/app_config.rb`:

```ruby
file "/app/config.yml" do
  mode 0644
  action :create
  content YAML.dump(
    'message' => "New message\n"
  )
end
```

Запускаем пересборку:

```shell
$ dapp dimg build
Setup group
  Git artifacts: apply patches (before setup) ...                           [OK] 0.96 sec
  Setup ...                                                                 [OK] 3.57 sec
  Git artifacts: apply patches (after setup) ...                            [OK] 0.92 sec
Docker instructions ...                                                     [OK] 0.99 sec
```

Пересборка прошла, начиная со стадии setup, что соответствует нашим изменениям в файле `.dapp_chef/recipes/setup/app_config.rb`. Перезапускаем контейнер и проверяем сообщение:

```shell
$ docker rm -f dapp-example-chef-for-advanced-build-1
f89d8357dd4a9c79076e741a5713a4147f71516651a58318fa9269b4b0f48172
$ dapp dimg run --detach -p 4567:4567 --name dapp-example-chef-for-advanced-build-1
5b34ff3fe4e6f5456f1df3d7c5339566bacf8a0df2225229a542b56f1f6c026e
$ curl localhost:4567/message
New message
```

### Добавление файлов и шаблонов для chef

В директории .dapp\_chef можно определять файлы и шаблоны для использования в рецептах. Файлы и шаблоны обязательно привязаны либо к одной стадии — в этом случае они доступны для всех рецептов стадии, либо к одному рецепту стадии — в этом случае они доступны только одному рецепту конкретной стадии. Общая структура файлов:

* files
  * \<стадия\>
    * common
      * \<имя файла\>
    * \<рецепт\>
      * \<имя файла\>
* templates
  * \<стадия\>
    * common
      * \<имя файла шаблона\>
    * \<рецепт\>
      * \<имя файла шаблона\>

В директории common лежат файлы и шаблоны, доступные независимо от включенных рецептов для сборки образа. Файлы и шаблоны из директории рецепта доступны при сборке только если рецепт включен для сборки образа директивой `chef.recipe`.

Технически сборщик устроен так, что при сборке стадии файлы из директории common и директории рецепта совмещаются в единую директорию `{files|templates}/default`. Поэтому указывать в рецептах параметр source для ресурсов cookbook\_file и template для простейших случаев не обязательно. Файлы в common и директории рецепта не могут иметь одинаковых имен — это приведет к ошибке сборки.

Для примера, переделаем генерацию конфига /app/config.yml приведенного выше на использование шаблона. Создаем файл шаблона для рецепта app\_config на стадии setup `.dapp_chef/templates/setup/app_config/config.yml.erb`:

```yaml
message: "<%= @message %>\n"
```

Правим рецепт `.dapp_chef/recipes/setup/app_config.rb` на использование этого шаблона:

```ruby
template '/app/config.yml' do
  mode 0644
  action :create
  variables(message: 'Passed through variable for template')
end
```

Собираем новую версию образа, перезапускаем приложение и проверяем сообщение:

```shell
$ dapp dimg build
Setup group
  Git artifacts: apply patches (before setup) ...                           [OK] 0.93 sec
  Setup ...                                                                 [OK] 3.64 sec
  Git artifacts: apply patches (after setup) ...                            [OK] 0.9 sec
Docker instructions ...                                                     [OK] 0.95 sec
$ docker rm -f dapp-example-chef-for-advanced-build-1
$ dapp dimg run --detach -p 4567:4567 --name dapp-example-chef-for-advanced-build-1
be7e96bf0da16a280dac800df76740507dfdaa859eca0c412086e13c1c1c5c9e
$ curl localhost:4567/message
Passed through variable for template
```

### Зачем нужно разделение файлов и шаблонов по рецептам

Разделение файлов по рецептам условно и фактически ничего не дает в случае, если все рецепты включаются одновременно для сборки одного образа. Однако, когда происходит сборка нескольких образов в рамках одного Dappfile, и для сборки разных образов используются разные рецепты — работает разделение файлов по используемым рецептам. Рассмотрим Dappfile:

```ruby
dimg_group do
  dimg 'A' do
    chef.recipe 'X'
  end

  dimg 'B' do
    chef.recipe 'Y'
  end
end
```

В данной конфигурации:

* все файлы, определенные, например, для стадии before\_install в `.dapp_chef/{files|templates}/before_install/common` будут доступны при сборкe стадии before\_install обоих образов A и B;
* файлы, определенные в `.dapp_chef/{files|templates}/before_install/X` будут доступны только при сборке стадии before\_install образа A, т.к. рецепт X используется только для сборки образа A;
* файлы, определенные в `.dapp_chef/{files|templates}/before_install/Y` будут доступны только при сборке стадии before\_install образа B, т.к. рецепт Y используется только для сборки образа B.

### Атрибуты

Атрибуты для сборочного cookbook'а определяются не через файлы атрибутов, а прямо в Dappfile. Для определения атрибутов пользователь заполняет hash через определенные директивы. Поддерживается автоматическое создание вложенных hash'ей если идет обращение к ранее не существующему ключу (по ключу 'a' в примере ниже автоматически создается hash).

```ruby
chef do
  attributes['a']['k1'] = 'value'
  attributes['a']['k2'] = ['one', 'two', 'three']

  _before_install_attributes['a']['k1'] = 'value_for_before_install'
  _before_install_attributes['a']['k3'] = 'value'
  _setup_attributes['b']['c']['d']['e'] = 'value'
end
```

Для определения общих атрибутов для всех стадий сборки используется директива `chef.attributes`.

Атрибуты как правило требуются для настройки подключаемых модулей chef сборки — dimod'ов. Об этом — в следующем разделе документации: [Dimod — модули сборки для chef](chef_dimod_for_advanced_build.html).

Чтобы переопределить или дополнить атрибуты для какой-то конкретной стадии используются директивы `chef._before_install_attributes`, `chef._install_attributes`, `chef._before_setup_attributes`, `chef._setup_attributes`. Соответственно эти директивы имеют приоритет и перетирают определенные через директиву `chef.attributes` атрибуты. Однако их использование по умолчанию не рекомендуется. В разделе [Пересборка при изменении атрибутов](chef_dimod_for_advanced_build.html#пересборка-при-изменении-атрибутов) описан пример такого определения атрибутов.


