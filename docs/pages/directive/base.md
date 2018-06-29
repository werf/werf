---
title: Базовые директивы в Dappfile
sidebar: doc_sidebar
permalink: base_directives.html
folder: directive
---

### Особенности

* В одном Dappfile может быть только один безымянный образ.
* Можно определять несколько образов в одном Dappfile.
* На верхнем уровне могут использоваться только директивы dimg и dimg\_group.
* Часть директив имеют несколько возможных форм записи.

### dimg \<name\>

Определяет образ dimg для сборки.

* Наследуются все настройки родительского контекста.
* Можно дополнять или переопределять настройки родительского контекста.

#### Примеры

##### Собирать образ, не указывая имени
```ruby
dimg do
  docker.from 'image:tag'
end
```

##### Собирать образы X и Y
```ruby
dimg 'X' do
  docker.from 'image1:tag'
end

dimg 'Y' do
  docker.from 'image2:tag'
end
```

### dimg\_group

* При использовании блока создается новый контекст.
  * Можно использовать директиву dimg\_group внутри нового контекста.

#### Примеры

##### Собирать образы X, Y и Z
```ruby
dimg_group do
  docker.from 'image:tag'

  dimg 'X'

  dimg_group do
    dimg 'Y'

    dimg_group do
      dimg 'Z'
    end
  end
end
```
