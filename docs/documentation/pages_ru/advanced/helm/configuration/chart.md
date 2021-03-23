---
title: Chart
permalink: advanced/helm/configuration/chart.html
---

Чарт — набор конфигурационных файлов описывающих приложение. Файлы чарта находятся в папке `.helm`, в корневой папке проекта:

```
.helm/
  templates/
    <name>.yaml
    <name>.tpl
    <some_dir>/
      <name>.yaml
      <name>.tpl
  charts/
  secret/
  values.yaml
  secret-values.yaml
```

Чарт werf может содержать опциональный файл `.helm/Chart.yaml` с описанием чарта, который полностью совместим с [`Chart.yaml`](https://helm.sh/docs/topics/charts/) и может содержать примерно следующее:

```yaml
apiVersion: v2
name: mychart
version: 1.0.0
dependencies:
 - name: redis
   version: "12.7.4"
   repository: "https://charts.bitnami.com/bitnami" 
```

По умолчанию werf будет использовать [имя проекта]({{ "/reference/werf_yaml.html#имя-проекта" | true_relative_url }}) из `werf.yaml` в качестве имени чарта. Версия чарта по умолчанию: `1.0.0`. Это можно переопределить создав файл `.helm/Chart.yaml` с явным переопределением имени чарта или его версии:

```yaml
name: mychart
version: 2.4.6
```

`.helm/Chart.yaml` также требуется для определения [зависимостей чарта]({{ "/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}).
