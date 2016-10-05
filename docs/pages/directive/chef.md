---
title: Chef директивы
sidebar: doc_sidebar
permalink: chef_directives.html
folder: directive
---

### chef.module \<mod\>[, \<mod\>, \<mod\> ...]
Включить переданные [модули](definitions.html#mdapp-модуль) для chef builder в данном контексте.

* Для каждого переданного модуля может существовать по одному рецепту на каждую из стадий.
* При отсутствии файла рецепта в runlist для данной стадии используется пустой рецепт \<mod\>::void.

Подробнее см.: [mdapp модуль](definitions.html#mdapp-модуль) и [установка стадии cookbook\`а](definitions.html#установка-стадии-cookbookа).

### chef.skip_module \<mod\>[, \<mod\>, \<mod\> ...]
Выключить переданные модули для chef builder в данном контексте.

### chef.reset_modules
Выключить все модули для chef builder в данном контексте.

### chef.recipe \<recipe\>[, \<recipe\>, \<recipe\> ...]
Включить переданные рецепты из [приложения](definitions.html#cookbook-приложения) для chef builder в данном контексте.

* Для каждого преданного рецепта может существовать файл рецепта в проекте на каждую из стадий.
* При отсутствии хотя бы одного файла рецепта из включенных, в runlist для данной стадии используется пустой рецепт \<projectname\>::void.
* Порядок вызова рецептов в runlist совпадает порядком их описания в конфиге.

Подробнее см.: [cookbook приложения](definitions.html#cookbook-приложения) и [установка стадии cookbook\`а](definitions.html#установка-стадии-cookbookа).

### chef.remove_recipe \<recipe\>[, \<recipe\>, \<recipe\> ...]
Выключить переданные рецепты из [приложения](definitions.html#cookbook-приложения) для chef builder в данном контексте.

### chef.reset_recipes
Выключить все рецепты из проекта для chef builder в данном контексте.

### chef.attributes
Хэш атрибутов, доступных на всех стадиях сборки, для chef builder в данном контексте.

* Вложенные хэши создаются автоматически при первом обращении к методу доступа по ключу (см. пример).

#### Пример

```ruby
chef.attributes['mdapp-test']['nginx']['package_name'] = 'nginx-common'
chef.attributes['mdapp-test']['nginx']['package_version'] = '1.4.6-1ubuntu3.5'

app 'X' do
  chef.attributes['mdapp-test']['nginx']['package_version'] = '1.4.6-1ubuntu3'
end
```

См.: [установка стадии cookbook\`а](definitions.html#установка-стадии-cookbookа).

### chef.\<стадия\>_attributes
Хэш атрибутов, доступных на стадии сборки, для chef builder в данном контексте.

См.: [установка стадии cookbook\`а](definitions.html#установка-стадии-cookbookа).

### chef.reset_attributes
Выключить атрибуты, доступные на всех стадиях сборки, для chef builder в данном контексте.

### chef.reset_\<стадия\>_attributes
Выключить атрибуты, доступные на стадии сборки, для chef builder в данном контексте.

### chef.reset_all_attributes
Выключить все атрибуты для chef builder в данном контексте.

### chef.reset_all
Выключить все рецепты из [приложения](definitions.html#cookbook-приложения), все модули для chef builder и все атрибуты в данном контексте.
