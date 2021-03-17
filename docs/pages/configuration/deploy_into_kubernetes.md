---
title: Deploy into Kubernetes
sidebar: documentation
permalink: configuration/deploy_into_kubernetes.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

## Release name

werf allows to define a custom release name template, which [used during deploy process]({{ site.baseurl }}/reference/deploy_process/deploy_into_kubernetes.html#release-name) to generate a release name.

Custom release name template is defined in the [meta configuration section]({{ site.baseurl }}/configuration/introduction.html#meta-config-section) of `werf.yaml`:

```yaml
project: PROJECT_NAME
configVersion: 1
deploy:
  helmRelease: TEMPLATE
  helmReleaseSlug: false
```

`deploy.helmRelease` is a Go template with `[[` and `]]` delimiters. There are `[[ project ]]`, `[[ env ]]` functions support. Default: `[[ project ]]-[[ env ]]`.

`deploy.helmReleaseSlug` defines whether to apply or not [slug]({{ site.baseurl }}/reference/deploy_process/deploy_into_kubernetes.html#slugging-the-release-name) to generated helm release name. Default: `true`.

`TEMPLATE` as well as any value of the config can include [werf Go templates functions]({{ site.baseurl }}/configuration/introduction.html#go-templates). E.g. you can mix the value with an environment variable:

{% raw %}
```yaml
deploy:
  helmRelease: >-
    [[ project ]]-{{ env "HELM_RELEASE_EXTRA" }}-[[ env ]]
```
{% endraw %}

## Kubernetes namespace

werf allows to define a custom Kubernetes namespace template, which [used during deploy process]({{ site.baseurl }}/reference/deploy_process/deploy_into_kubernetes.html#kubernetes-namespace) to generate a Kubernetes Namespace.

Custom Kubernetes Namespace template is defined in the [meta configuration section]({{ site.baseurl }}/configuration/introduction.html#meta-config-section) of `werf.yaml`:

```yaml
project: PROJECT_NAME
configVersion: 1
deploy:
  namespace: TEMPLATE
  namespaceSlug: true|false
```

`deploy.namespace` is a Go template with `[[` and `]]` delimiters. There are `[[ project ]]`, `[[ env ]]` functions support. Default: `[[ project ]]-[[ env ]]`.

`deploy.namespaceSlug` defines whether to apply or not [slug]({{ site.baseurl }}/reference/deploy_process/deploy_into_kubernetes.html#slugging-kubernetes-namespace) to generated kubernetes namespace. Default: `true`.
