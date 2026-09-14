---
title: werf.yaml
permalink: reference/werf_yaml.html
description: Пример конфигурации werf
toc: false
---

{% include reference/werf_yaml/table.html %}

{% include pages/ru/json_schema.md.liquid file="werf.yaml" schema="werf.json" file_match='"werf.yaml", "werf.yml"' %}

`werf.yaml` перед разбором рендерится как Go-шаблон, поэтому схема описывает уже отрендеренный YAML. Редактор валидирует файл только там, где он является корректным YAML до рендеринга; участки с шаблонными конструкциями редактор пропускает или помечает как синтаксические ошибки сам, вне схемы.
