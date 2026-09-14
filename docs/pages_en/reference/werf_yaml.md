---
title: werf.yaml
permalink: reference/werf_yaml.html
description: werf.yaml config
toc: false
---

{% include reference/werf_yaml/table.html %}

{% include pages/en/json_schema.md.liquid file="werf.yaml" schema="werf.json" file_match='"werf.yaml", "werf.yml"' %}

`werf.yaml` is rendered as a Go template before parsing, so the schema describes the rendered YAML. An editor validates the file only where it is valid YAML before rendering; the sections containing template actions are skipped or reported as syntax errors by the editor, not by the schema.
