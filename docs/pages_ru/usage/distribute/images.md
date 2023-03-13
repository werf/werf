---
title: Образы
permalink: usage/distribute/images.html
---

## О дистрибуции образов

Дистрибуция собираемых werf образов для использования сторонними пользователями и/или ПО осуществляется командой `werf export`. Эта команда соберёт и опубликует образы в container registry, при этом убрав все ненужные для стороннего ПО метаданные, чем полностью выведет образы из-под контроля werf, позволив организовать их дальнейший жизненный цикл сторонними средствами.

> Опубликованные командой `werf export` образы *никогда* не будут удаляться командой `werf cleanup`, в отличие от образов, опубликованных обычным способом. Очистка экспортированных образов должна быть реализована сторонними средствами.

## Дистрибуция образа

```shell
werf export \
    --repo example.org/myproject \
    --tag other.example.org/myproject/myapp:latest
```

Результат: образ собран и сначала опубликован с content-based тегом в container registry `example.org/myproject`, а затем опубликован в другой container registry `other.example.org/myproject` как конечный экспортированный образ `other.example.org/myproject/myapp:latest`.

В параметре `--tag` можно указать тот же репозиторий, что и в `--repo`, таким образом используя один и тот же container registry и для сборки, и для экспортированного образа.

## Дистрибуция нескольких образов

В параметре `--tag` можно использовать шаблоны `%image%`, `%image_slug%` и `%image_safe_slug%` для подставления имени образа из `werf.yaml`, основанном на его содержимом, например:

```shell
werf export \
    --repo example.org/mycompany/myproject \
    --tag example.org/mycompany/myproject/%image%:latest
```

## Дистрибуция произвольных образов

Используя позиционные аргументы и имена образов из `werf.yaml` можно выбрать произвольные образы, например:

```shell
werf export backend frontend \
    --repo example.org/mycompany/myproject \
    --tag example.org/mycompany/myproject/%image%:latest
```

## Использование content-based тега при формировании тега

В параметре `--tag` можно использовать шаблон `%image_content_based_tag%` для использования тега образа, основанном на его содержимом, например:

```shell
werf export \
    --repo example.org/mycompany/myproject \
    --tag example.org/mycompany/myproject/myapp:%image_content_based_tag%
```

> Опубликованные командой `werf export` образы *никогда* не будут удаляться командой `werf cleanup`, в отличие от образов, опубликованных обычным способом. Очистка экспортированных образов может быть реализована сторонними средствами.
