---
title: Deploy to kubernetes
sidebar: reference
permalink: reference/deploy/deploy_to_kubernetes.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Werf uses [Helm](https://helm.sh/) Kubernetes package manager to deploy applications into kubernetes. Before using Werf for deploy you should install Helm and its back-end Tiller, [see installation for details]({{ site.baseurl }}/how_to/installation.html).

## Helm chart

The `.helm` directory in the project root describes a helm chart. Helm chart provides a description of the application configuration and application component used to release into the kubernetes cluster using the helm toolset. The chart for werf, i.e. `.helm` folder, has the following structure:

```
.helm/
  templates/
    <name>.yaml
    <name>.tpl
  charts/
  secret/
  values.yaml
  secret-values.yaml
```

The `Chart.yaml` file, which is required for the standard helm chart, is not needed. Werf will autogenerate `Chart.yaml` during deploy process. Chart name (and `.Chart.Name` value) will be the same as [project name]({{ site.baseurl }}/reference/config.html#meta-configuration-doc).

In the `templates` directory, YAML file (chart element) templates are stored, which describe the resources for their publishing in the cluster. Detailed information about template creation is available in a separate [section]({{ site.baseurl }}/reference/deploy/chart_configuration.html#features-of-chart-template-creation). The `charts` directory is used when you need to work with external charts.

Chart structure includes additional elements that are not contained within the structure of a standard helm chart – these are the `secret-values.yaml` file and the `secret` directory that is described in detail in the [working with secrets section]({{ site.baseurl }}/reference/deploy/secrets.html).

## Connection settings

Werf uses standard configuration file `~/.kube/config` to connect to Kubernetes cluster.

## Additions to helm

Werf has several additions to the helm.

### Chart generation

During werf deploy a temporary helm chart is created.

This chart contains:

* Additional generated go-templates: `werf_container_image`, `werf_container_env` and other. These templates are described in [the templates article]({{ site.baseurl }}/reference/deploy/chart_configuration.html#features-of-chart-template-creation).
* Decoded secret values yaml file. The secrets are described in [the secrets article]({{ site.baseurl }}/reference/deploy/secrets.html).

The temporary chart then passed to the helm. Werf deletes this chart on the werf deploy command termination.

### Watch resources

Werf watches resources statuses and logs during the deploy process. More info is available in the [watch resources article]({{ site.baseurl }}/reference/deploy/track_kubernetes_resources.html).

## Environment

Application can be deployed to multiple environments, like staging, testing, production, development, etc.

Werf has basic support for environments to automate generation of external names, such as Helm Release name or Kubernetes Namespace.

Environment is a required parameter for deploy and should be specified either with option `--env` or automatically determined for the used CI system. Werf currently support only [Gitlab CI environments integration](#integration-with-gitlab).

### Integration with Gitlab

Gitlab has [environments support](https://docs.gitlab.com/ce/ci/environments.html). Werf will detect current environment for the pipeline in gitlab and use it as environment parameter.

`CI_ENVIRONMENT_SLUG` gitlab variable used in Werf to determine environment name in gitlab pipeline.

## Multiple Kubernetes clusters

There are cases when separate Kubernetes clusters are needed for a different environments. You can [configure access to multiple clusters](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters) using kube contexts in a single kube config.

In that case deploy option `--kube-context=CONTEXT` should be specified manually along with the environment.

## Helm Release name

By default Helm Release name will be constructed by template `[[ project ]]-[[ env ]]`. Where `[[ project ]]` refers to the [project name]({{ site.baseurl }}/reference/config.html#meta-configuration-doc) and `[[ env ]]` refers to the specified or detected environment.

For example for project named `symfony-demo` there will be following Helm Release names depending on the specified environment:
* `symfony-demo-stage` for the `stage` environment;
* `symfony-demo-test` for the `test` environment;
* `symfony-demo-prod` for the `prod` environment.

Helm Release name could be redefined by deploy option `--release NAME`. In that case Werf will use specified name as is.

Custom Helm Release template can also be defined in the [meta configuration doc]({{ site.baseurl }}/reference/config.html#meta-configuration-doc) of `werf.yaml`:

```yaml
project: PROJECT_NAME
deploy:
  helmRelease: TEMPLATE
```

`deploy.helmRelease` is a Go template with `[[` and `]]` delimiters. There are `[[ project ]]`, `[[ env ]]` functions support. [Sprig functions](https://masterminds.github.io/sprig/) can also be used (for example function `env` to retrieve environment variables).

### Slug

Helm Release name constructed by template will be slugified to fit release name requirements by [*release slug procedure*]({{ site.baseurl }}/reference/toolbox/slug.html#basic-algorithm), which generates unique valid Helm Release name.

This is default behaviour, which can be disabled by [meta configuration doc]({{ site.baseurl }}/reference/config.html#meta-configuration-doc) option `deploy.helmReleaseSlug`:

```
project: PROJECT_NAME
deploy:
  helmReleaseSlug: false
```

Werf will not apply release slug procedure for the release name specified with `--release NAME` option.

## Kubernetes Namespace

By default Kubernetes Namespace will be constructed by template `[[ project ]]-[[ env ]]`. Where `[[ project ]]` refers to the [project name]({{ site.baseurl }}/reference/config.html#meta-configuration-doc) and `[[ env ]]` refers to the determined environment.

For example for project named `symfony-demo` there will be following Kubernetes Namespaces depending on the specified environment:
* `symfony-demo-stage` for the `stage` environment;
* `symfony-demo-test` for the `test` environment;
* `symfony-demo-prod` for the `prod` environment.

Kubernetes Namespace could be redefined by deploy option `--namespace NAMESPACE`. In that case Werf will use specified name as is.

Custom Kubernetes Namespace template can also be defined in the [meta configuration doc]({{ site.baseurl }}/reference/config.html#meta-configuration-doc) of `werf.yaml`:

```yaml
project: PROJECT_NAME
deploy:
  namespace: TEMPLATE
```

`deploy.namespace` is a Go template with `[[` and `]]` delimiters. There are `[[ project ]]`, `[[ env ]]` functions support. [Sprig functions](https://masterminds.github.io/sprig/) can also be used (for example function `env` to retrieve environment variables).

### Slug

Kubernetes Namespace constructed by template will be slugified to fit [DNS Label](https://www.ietf.org/rfc/rfc1035.txt) requirements by [*namespace slug procedure*]({{ site.baseurl }}/reference/toolbox/slug.html#basic-algorithm), which generates unique valid Kubernetes Namespace.

This is default behaviour, which can be disabled by [meta configuration doc]({{ site.baseurl }}/reference/config.html#meta-configuration-doc) option `deploy.namespaceSlug`:

```
project: PROJECT_NAME
deploy:
  namespaceSlug: false
```

Werf will not apply namespace slug procedure for the namespace specified with `--namespace NAMESPACE` option.

## Deploy command

{% include /cli/werf_deploy.md %}

## Dismiss command

{% include /cli/werf_dismiss.md %}
