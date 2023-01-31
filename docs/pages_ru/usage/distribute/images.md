---
title: Образы
permalink: usage/distribute/images.html
---

## Дистрибуция образов для развертывания сторонним ПО

Дистрибуция собираемых werf образов для использования сторонним ПО осуществляется командой `werf export`. Эта команда соберёт и опубликует образы в Container Registry, при этом убрав с них все ненужные для стороннего ПО метаданные, чем полностью выведет образы из под контроля werf, позволив организовать их дальнейший жизненный цикл сторонними средствами.

Пример:

```shell
werf export --repo example.org/myproject --tag other.example.org/myproject/myapp:latest
```

Результат: образ собран и сначала опубликован с content-based тегом в Container Registry `example.org/myproject`, а затем опубликован в другой Container Registry `other.example.org/myproject` как конечный экспортированный образ `other.example.org/myproject/myapp:latest`.

В параметре `--tag` можно указать тот же репозиторий, что и в `--repo`, таким образом используя один и тот же Container Registry и для сборки, и для экспортированного образа.

Также в параметре `--tag` можно использовать шаблоны `%image%`, `%image_slug%` и `%image_safe_slug%` для подставления имени образа и `%image_content_based_tag%` для подставления тега образа, основанном на его содержимом, например:

```shell
werf export --repo example.org/mycompany/myproject --tag example.org/mycompany/myproject/%image%:%image_content_based_tag%
```

> Опубликованные командой `werf export` образы *никогда* не будут удаляться командой `werf cleanup`, в отличие от образов, опубликованных обычным способом. Очистка экспортированных образов может быть реализована сторонними средствами.
