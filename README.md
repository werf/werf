# dapp [![Gem Version](https://badge.fury.io/rb/dapp.svg)](https://badge.fury.io/rb/dapp) [![Build Status](https://travis-ci.org/flant/dapp.svg)](https://travis-ci.org/flant/dapp) [![Code Climate](https://codeclimate.com/github/flant/dapp/badges/gpa.svg)](https://codeclimate.com/github/flant/dapp) [![Test Coverage](https://codeclimate.com/github/flant/dapp/badges/coverage.svg)](https://codeclimate.com/github/flant/dapp/coverage)

## Reference

### Dappfile

#### Основное
*TODO*

#### Артифакты
*TODO*

#### Docker
*TODO*

#### Shell
*TODO*

#### Chef

##### chef.module \<mod\>[, \<mod\>, \<mod\> ...]
Включить переданные модули для chef builder в данном контексте.

* Для каждого переданного модуля может существовать по одному рецепту на каждый из stage.
* Файл рецепта для \<stage\>: recipes/\<stage\>.rb
* Рецепт модуля будет добавлен в runlist для данного stage если существует файл рецепта.
* Порядок вызова рецептов модулей в runlist совпадает порядком их описания в конфиге.
* При сборке stage, для каждого из включенных модулей, при наличии файла рецепта, будут скопированы:
  * files/\<stage\>/ -> files/default/
  * templates/\<stage\>/ -> templates/default/
  * metadata.json

##### chef.skip_module \<mod\>[, \<mod\>, \<mod\> ...]
Выключить переданные модули для chef builder в данном контексте.

##### chef.reset_modules
Выключить все модули для chef builder в данном контексте.

##### chef.recipe \<recipe\>[, \<recipe\>, \<recipe\> ...]
Включить переданные рецепты из проекта для chef builder в данном контексте.

* Для каждого преданного рецепта может существовать файл рецепта в проекте на каждый из stage.
* Файл рецепта для \<stage\>: recipes/\<stage\>/\<recipe\>.rb
* Рецепт будет добавлен в runlist для данного stage если существует файл рецепта.
* Порядок вызова рецептов в runlist совпадает порядком их описания в конфиге.
* При сборке stage, при наличии хотя бы одного файла рецепта из включенных, будут скопированы:
  * files/\<stage\> -> files/default/
  * templates/\<stage\>/ -> templates/default/
  * metadata.json

##### chef.remove_recipe \<recipe\>[, \<recipe\>, \<recipe\> ...]
Выключить переданные рецепты из проекта для chef builder в данном контексте.

##### chef.reset_recipes
Выключить все рецепты из проекта для chef builder в данном контексте.

##### chef.reset_all
Выключить все рецепты из проекта и все модули для chef builder в данном контексте.

##### Примеры
* [Dappfile](doc/example/Dappfile.chef.1)

### Команды

#### dapp build
*TODO*

#### dapp push
*TODO*

#### dapp smartpush
*TODO*

## Architecture

### Стадии
*TODO*

### Хранение данных
*TODO*
