---
title: Директивы монтирования
sidebar: doc_sidebar
permalink: mount_directives.html
folder: directive
---

### mount \<to\>

Определяет [внешнюю директорию сборки](definitions.html#внешняя-директория-сборки) для переданного абсолютного пути \<to\>.

Поддиректива from определяет место размещения директории.

* при 'tmp_dir', во [временной директории приложения](definitions.html#временная-директория-приложения).
* при 'build_dir', в [директории сборки проекта](definitions.html#директория-сборки-проекта).

#### Примеры

##### Собрать с несколькими внешними директориями
```ruby
dimg do
  mount '/any_path' do
    from 'tmp_dir'
  end

  mount '/cache' do
    from 'build_dir'
  end
end
```
