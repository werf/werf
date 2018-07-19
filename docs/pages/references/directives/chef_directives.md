---
title: Использовать функции chef
sidebar: reference
permalink: chef_directives.html
folder: directive
---

При сборке образов можно использовать функции chef

Поддержка сборщиками

- Shell - нет
- Chef - да
- Ansible - нет

### chef.dimod \<mod\>[[, \<version-constraint\>,] \<cookbook-opts\>]
Включить указанный модуль для chef builder в данном контексте.

* Название модуля должно включать в себя префикс 'dimod-' (dimod-php, dimod-nginx).
* Для каждого переданного модуля может существовать по одному рецепту на каждую из стадий.
* При отсутствии файла рецепта в runlist для данной стадии используется пустой рецепт \<mod\>::void.
* Параметры \<version-constraint\> и \<cookbook-opts\> определяют опции cookbook'а, соответствуют параметрам директивы chef.cookbook.

Подробнее см.: [dimod модуль](chef_dimod.html) и [chef-сборщик](build_chef.html)

### chef.cookbook \<cookbook\>[[, \<version-constraint\>,] \<cookbook-opts\>]
Включить указанный cookbook в зависимость для сборочного cookbook'а.

* Опциональный параметр \<version-constraint\> определяет ограничение на версию cookbook'а.
* Опции \<cookbook-opts\> соответствуют опциям cookbook'ов из Berksfile.

### chef.recipe \<recipe\>[, \<recipe\>, \<recipe\> ...]
Включить переданные рецепты из cookbook'а dapp для chef builder в данном контексте.

* Для каждого преданного рецепта может существовать файл рецепта в проекте на каждую из стадий.
* При отсутствии хотя бы одного файла рецепта из включенных, в runlist для данной стадии используется пустой рецепт \<projectname\>::void.
* Порядок вызова рецептов в runlist совпадает порядком их описания в конфиге.

### chef.attributes
Хеш атрибутов, доступных на всех стадиях сборки, для chef builder в данном контексте.

* Вложенные хеши создаются автоматически при первом обращении к методу доступа по ключу (см. пример).

#### Пример (строчная запись)

```ruby
dimg_group do
  docker.from 'image:tag'

  chef.attributes['dimod-test']['nginx']['package_name'] = 'nginx-common'
  chef.attributes['dimod-test']['nginx']['package_version'] = '1.4.6-1ubuntu3.5'

  dimg do
    chef.attributes['dimod-test']['nginx']['package_version'] = '1.4.6-1ubuntu3'
  end
end
```


### chef.\<стадия\>_attributes
Хеш атрибутов, доступных на стадии сборки, для chef builder в данном контексте.

### Примеры

#### Собрать с несколькими модулями и рецептами

```ruby
dimg_group do
  docker.from 'image:tag'

  chef do
    dimod 'mod1', 'mod2'
    recipe 'recipe1'
  end

  dimg do
    chef do
      dimod 'mod3'
      recipe 'recipe2', 'recipe3'
    end
  end
end
```

#### Собрать с несколькими модулями и рецептами (строчная запись)

```ruby
dimg_group do
  docker.from 'image:tag'

  chef.dimod 'mod1', 'mod2'
  chef.recipe 'recipe1'

  dimg do
    chef.dimod 'mod3'
    chef.recipe 'recipe2', 'recipe3'
  end
end
```
