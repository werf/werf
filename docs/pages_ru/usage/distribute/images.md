---
title: Образы
permalink: usage/distribute/images.html
---

## Дистрибуция образов для развертывания через werf

Дистрибуция собранных образов в Container Registry для дальнейшего развертывания через werf происходит автоматически при сборке образов, если указан параметр `--repo` (`$WERF_REPO`). При этом большинство других команд werf первым шагом автоматически производят и сборку и дистрибуцию образов, например:

```shell
werf converge --repo example.org/myrepo
```

Результат: образы будут собраны, опубликованы в Container Registry, а сразу затем использованы при развертывании в Kubernetes.

> Опубликованные таким способом образы находятся под управлением werf и не рекомендуются к развертыванию сторонним ПО, так как, например, команда `werf cleanup` может удалить эти образы, так как не имеет информацию об их стороннем использовании.

Шаг сборки/публикации образов можно выделить в отдельный шаг с помощью опции `--skip-build`, доступной для большинства команд, например:

```shell
werf build --repo example.org/myrepo
```

```shell
werf converge --skip-build --repo example.org/myrepo
```

## Дистрибуция образов для развертывания сторонним ПО

Дистрибуция собираемых werf образов для использования сторонним ПО осуществляется командой `werf export`. Эта команда соберёт и опубликует образы в Container Registry, при этом убрав с них все ненужные для стороннего ПО метаданные, чем полностью выведет образы из под контроля werf, позволив организовать их дальнейший жизненный цикл сторонними средствами.

Пример:

```shell
werf export --repo example.org/mycompany/myproject --tag example.org/mycompany/myproject/myapp:latest
```

Результат: в Container Registry `example.org/mycompany/myproject` будет опубликован образ `example.org/mycompany/myproject/myapp` с тегом `latest`.

В параметре `--tag` можно использовать шаблоны `%image%`, `%image_slug%` и `%image_safe_slug%` для подставления имени образа и `%image_content_based_tag%` для подставления тега образа, основанном на его содержимом, например:

```shell
werf export --repo example.org/mycompany/myproject --tag example.org/mycompany/myproject/%image%:%image_content_based_tag%
```

> Опубликованные командой `werf export` образы *никогда* не будут удаляться командой `werf cleanup`. Очистка таких образов может быть реализована сторонними средствами.
