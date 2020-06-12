---
title: Lint And Render Chart
sidebar: documentation
permalink: documentation/reference/development_and_debug/lint_and_render_chart.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

During development of [chart templates]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#templates) user need check these templates for errors before running full deploy process.

There are 2 tools to do that:

 1. Render templates.
 2. Lint templates.

## Render

During templates rendering werf will create a single text stream of all manifests defined in the chart expanding Go templates.

Use [`werf helm render` command]({{ site.baseurl }}/documentation/cli/management/helm/render.html) to get rendered manifests. The same params as in [`werf deploy` command]({{ site.baseurl }}/documentation/cli/main/deploy.html) can be passed (such as additional [values]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#values), images repo, environment and other).

Render command aimed to help in debugging of problems related to the wrong usage of Go templates or inspect yaml format of Kubernetes manifests.

## Lint

Lint checks a [chart]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#chart) for different issues, such as:
 * errors in Go templates;
 * errors in YAML syntax;
 * errors in Kubernetes manifest syntax: wrong kind, missed resource fields, etc.;
 * errors in Kubernetes runtime logic [coming soon](https://github.com/werf/werf/issues/1187): missed resources labels, wrong names of related resources specified, check resources api version, etc.;
 * security risk analysis [coming soon](https://github.com/werf/werf/issues/1317).

[`werf helm lint` command]({{ site.baseurl }}/documentation/cli/management/helm/lint.html) runs all of these checks and can be used either in local development or in the CI/CD pipeline to automate chart checking procedure. The same params as in [`werf deploy` command]({{ site.baseurl }}/documentation/cli/main/deploy.html) can be passed (such as additional [values]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#values), images repo, environment and other).
