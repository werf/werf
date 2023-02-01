---
title: Multiple environments
permalink: usage/deploy/environments.html
---

## Environment-dependent template parameters (werf only)

You can set the *environment* to use in werf with the `--env` option (`$WERF_ENV`). It can also be set automatically via the `werf ci-env` command. The current environment is stored in the `$.Values.werf.env` parameter of the main chart.

The werf environment is used when generating the release name and Namespace name and can also be used to parameterize templates:

```yaml
# .helm/values.yaml:
memory:
  staging: 1G
  production: 2G
```

{% raw %}

```
# .helm/templates/example.yaml:
memory: {{ index $.Values.memory $.Values.werf.env }}
```

{% endraw %}

```shell
werf render --env production
```

Output:

```yaml
memory: 2G
```

The `export-values` (werf only) directive provides a way to use `$.Values.werf.env` in dependent charts:

```yaml
# .helm/Chart.yaml:
dependencies:
- name: child
  export-values:
  - parent: werf
    child: werf
```

{% raw %}

```
# .helm/charts/child/templates/example.yaml:
{{ $.Values.werf.env }}
```

{% endraw %}

Output:

```
production
```

## Deployment to various Kubernetes Namespaces

The name of the Kubernetes Namespace for the resources being deployed is generated automatically (werf only) using a custom pattern `[[ project ]]-[[ env ]]`, where `[ project ]]` is the werf project name, and `[[ env ]]` is the environment name.

Thus, the Namespace changes when the werf environment changes:

```yaml
# werf.yaml:
project: myapp
```

```shell
werf converge --env staging
werf converge --env production
```

In this case, the first instance of the application will be deployed to the `myapp-staging` namespace, and the second will be deployed to the `myapp-production` namespace.

Note that if the Namespace is explicitly specified in the Kubernetes resource manifest, it will be used for that resource.

### Changing the Namespace naming pattern (werf only)

You can change the custom pattern used to generate the Namespace name.

```yaml
# werf.yaml:
project: myapp
deploy:
  namespace: "backend-[[ env ]]"
```

```shell
werf converge --env production
```

In this case, the application will be deployed to the `backend-production` Namespace.

### Specifying the Namespace name explicitly

You can specify the Namespace explicitly for each command instead of generating the Namespace using a custom pattern (it is recommended to change the release name as well):

```shell
werf converge --namespace backend-production --release backend-production
```

In this case, the application will be deployed to the `backend-production` Namespace.

### Formatting the Namespace name

Namespace generated using a custom pattern or specified by the `--namespace` option is converted to the [RFC 1123 Label Names](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names) format automatically. You can disable automatic formatting using the `deploy.namespaceSlug` directive in `werf.yaml`.

You can manually format any string according to RFC 1123 Label Names format using the `werf slugify -f kubernetes-namespace` command.

## Deploying to different Kubernetes clusters

By default, werf deploys Kubernetes resources to the cluster set for the `werf kubectl` command. You can use different kube-contexts of a single kube-config file (the default is `$HOME/.kube/config`) to deploy to different clusters:

```shell
werf converge --kube-context staging  # or $WERF_KUBE_CONTEXT=...
werf converge --kube-context production
```

... or use different kube-config files:

```shell
werf converge --kube-config "$HOME/.kube/staging.config"  # or $WERF_KUBE_CONFIG=...
werf converge --kube-config-base64 "$KUBE_PRODUCTION_CONFIG_IN_BASE64"  # or $WERF_KUBE_CONFIG_BASE64=...
```

## Deploying as a different Kubernetes user

By default, werf uses the Kubernetes user set for the `werf kubectl` command for deployment. You can use several kube contexts to deploy as different users:

```shell
werf converge --kube-context admin  # or $WERF_KUBE_CONTEXT=...
werf converge --kube-context regular-user
```
