---
title: Resource tracking
permalink: usage/deploy/tracking.html
---

werf uses the [kubedog library](https://github.com/werf/kubedog) to track resources. Tracking behavior can be configured for each resource by setting [resource annotations]({{ "/reference/deploy_annotations.html" | true_relative_url }}) in the chart templates.

## Subcharts

During deploy process werf will render, create and track all resources of all [subcharts]({{ "/usage/deploy/charts.html#dependent-charts" | true_relative_url }}).
