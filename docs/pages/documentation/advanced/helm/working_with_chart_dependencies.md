---
title: Working with chart dependencies
sidebar: documentation
permalink: documentation/advanced/helm/working_with_chart_dependencies.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

## Managing subcharts with requirements.yaml

For chart developers, it is often easier to manage a single dependency file which declares all dependencies.

A `requirements.yaml` file is a YAML file in which developers can declare chart dependencies, along with the location of the chart and the desired version. For example, this requirements file declares two dependencies:

```yaml
# requirements.yaml
dependencies:
- name: nginx
  version: "1.2.3"
  repository: "https://example.com/charts"
- name: sysdig
  version: "1.4.12"
  repository: "@stable"
```
