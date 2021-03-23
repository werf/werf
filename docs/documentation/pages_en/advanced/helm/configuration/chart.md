---
title: Chart
permalink: advanced/helm/configuration/chart.html
---

The chart is a set of configuration files which describe an application. Chart files reside in the `.helm` directory under the root directory of the project:

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

werf chart has an optional `.helm/Chart.yaml` description file, which is fully compatible with [helm`s `Chart.yaml`](https://helm.sh/docs/topics/charts/) and could contain following content:

```yaml
apiVersion: v2
name: mychart
version: 1.0.0
dependencies:
 - name: redis
   version: "12.7.4"
   repository: "https://charts.bitnami.com/bitnami" 
```

By default, werf will use [project name]({{ "/reference/werf_yaml.html#project-name" | true_relative_url }}) from the `werf.yaml` as a chart name, and default version is always `1.0.0`. You can redefine this by placing own `.helm/Chart.yaml` with overrides for chart name or version:

```yaml
name: mychart
version: 2.4.6
```

`.helm/Chart.yaml` is also needed to define [chart dependencies]({{ "/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}).
