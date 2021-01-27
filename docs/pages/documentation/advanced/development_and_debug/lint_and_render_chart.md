---
title: Lint and render chart
sidebar: documentation
permalink: documentation/advanced/development_and_debug/lint_and_render_chart.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

During development of [chart templates]({{ "documentation/advanced/helm/basics.html#templates" | true_relative_url }}) user need check these templates for errors before running full deploy process.

There are 2 tools to do that:

 1. Render templates.
 2. Lint templates.

## Render

During templates rendering werf will create a single text stream of all manifests defined in the chart expanding Go templates.

Use [`werf helm template` command]({{ "documentation/reference/cli/werf_helm_template.html" | true_relative_url }}) to get rendered manifests. The same params as in [`werf converge` command]({{ "documentation/reference/cli/werf_converge.html" | true_relative_url }}) can be passed (such as additional [values]({{ "documentation/advanced/helm/basics.html#values" | true_relative_url }}), repo, environment and other).

Render command aimed to help in debugging of problems related to the wrong usage of Go templates or inspect yaml format of Kubernetes manifests.

## Lint

Lint checks a [chart]({{ "documentation/advanced/helm/basics.html#chart" | true_relative_url }}) for different issues, such as:
 * errors in Go templates;
 * errors in YAML syntax;
 * errors in Kubernetes manifest syntax: wrong kind, missed resource fields, etc.;
 * errors in Kubernetes runtime logic [coming soon](https://github.com/werf/werf/issues/1187): missed resources labels, wrong names of related resources specified, check resources api version, etc.;
 * security risk analysis [coming soon](https://github.com/werf/werf/issues/1317).

[`werf helm lint` command]({{ "documentation/reference/cli/werf_helm_lint.html" | true_relative_url }}) runs all of these checks and can be used either in local development or in the CI/CD pipeline to automate chart checking procedure. The same params as in [`werf converge` command]({{ "documentation/reference/cli/werf_converge.html" | true_relative_url }}) can be passed (such as additional [values]({{ "documentation/advanced/helm/basics.html#values" | true_relative_url }}), repo, environment and other).
