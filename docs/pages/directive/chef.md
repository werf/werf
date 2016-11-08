---
title: Chef директивы
sidebar: doc_sidebar
permalink: chef_directives.html
folder: directive
---

### chef.dimod \<mod\>[, \<mod\>, \<mod\> ...]
Включить переданные [модули](definitions.html#mdapp-модуль) для chef builder в данном контексте.

* Для каждого переданного модуля может существовать по одному рецепту на каждую из стадий.
* При отсутствии файла рецепта в runlist для данной стадии используется пустой рецепт \<mod\>::void.

Подробнее см.: [mdapp модуль](definitions.html#mdapp-модуль) и [установка стадии cookbook\`а](definitions.html#установка-стадии-cookbook-а).

### chef.recipe \<recipe\>[, \<recipe\>, \<recipe\> ...]
Включить переданные рецепты из [приложения](definitions.html#cookbook-приложения) для chef builder в данном контексте.

* Для каждого преданного рецепта может существовать файл рецепта в проекте на каждую из стадий.
* При отсутствии хотя бы одного файла рецепта из включенных, в runlist для данной стадии используется пустой рецепт \<projectname\>::void.
* Порядок вызова рецептов в runlist совпадает порядком их описания в конфиге.

Подробнее см.: [cookbook приложения](definitions.html#cookbook-приложения) и [установка стадии cookbook\`а](definitions.html#установка-стадии-cookbook-а).

### chef.attributes
Хэш атрибутов, доступных на всех стадиях сборки, для chef builder в данном контексте.

* Вложенные хэши создаются автоматически при первом обращении к методу доступа по ключу (см. пример).

#### Пример (строчная запись)

```ruby
dimg_group do
  docker.from 'image:tag'
  
  chef.attributes['mdapp-test']['nginx']['package_name'] = 'nginx-common'
  chef.attributes['mdapp-test']['nginx']['package_version'] = '1.4.6-1ubuntu3.5'

  dimg do
    chef.attributes['mdapp-test']['nginx']['package_version'] = '1.4.6-1ubuntu3'
  end
end
```

См.: [установка стадии cookbook\`а](definitions.html#установка-стадии-cookbook-а).

### chef.\<стадия\>_attributes
Хэш атрибутов, доступных на стадии сборки, для chef builder в данном контексте.

См.: [установка стадии cookbook\`а](definitions.html#установка-стадии-cookbook-а).

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
