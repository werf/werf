---
title: Naming
permalink: advanced/helm/releases/naming.html
---

## Environment

By default, werf assumes that each release should be tainted with some environment, such as `staging`, `test` or `production`.

Using this environment, werf determines:

1. Release name.
2. Kubernetes namespace.

The environment is a required parameter for deploying and should be specified either with an `--env` option or determined automatically using the data for the CI/CD system used. See the [CI/CD configuration integration]({{ "internals/how_ci_cd_integration_works/general_overview.html#cicd-configuration-integration" | true_relative_url }}) for more info.

## Release name

The release name is constructed using the template `[[ project ]]-[[ env ]]` by default. Here, `[[ project ]]` refers to the [project name]({{ "reference/werf_yaml.html#project-name" | true_relative_url }}) and `[[ env ]]` refers to the specified or detected environment.

For example, for the project named `symfony-demo`, the following Helm Release names can be constructed depending on the environment:
* `symfony-demo-stage` for the `stage` environment;
* `symfony-demo-test` for the `test` environment;
* `symfony-demo-prod` for the `prod` environment.

You can redefine the release using the `--release NAME` deploy option. In that case werf would use the specified name as is.

You can also define the custom release name in the werf.yaml configuration [by setting `deploy.helmRelease`]({{ "/reference/werf_yaml.html#release-name" | true_relative_url }}).

### Slugging the release name

The name of the Helm Release constructed using the template will be slugified according to the [*release slug procedure*]({{ "internals/names_slug_algorithm.html#basic-algorithm" | true_relative_url }}) to fit the requirements for the release name. This procedure generates a unique and valid Helm Release name.

This is the default behavior. You can disable it by [setting `deploy.helmReleaseSlug=false`]({{ "/reference/werf_yaml.html#release-name" | true_relative_url }}) in the `werf.yaml` configuration.

## Kubernetes namespace

The Kubernetes namespace is constructed using the template `[[ project ]]-[[ env ]]` by default. Here, `[[ project ]]` refers to the [project name]({{ "reference/werf_yaml.html#project-name" | true_relative_url }}) and `[[ env ]]` refers to the environment.

For example, for the project named `symfony-demo`, there can be the following Kubernetes namespaces depending on the environment specified:

* `symfony-demo-stage` for the `stage` environment;
* `symfony-demo-test` for the `test` environment;
* `symfony-demo-prod` for the `prod` environment.

You can redefine the Kubernetes Namespace using the `--namespace NAMESPACE` deploy option. In that case, werf would use the specified name as is.

You can also define the custom Kubernetes Namespace in the werf.yaml configuration [by setting `deploy.namespace`]({{ "/reference/werf_yaml.html#kubernetes-namespace" | true_relative_url }}) parameter.

### Slugging Kubernetes namespace

The Kubernetes namespace that is constructed using the template will be slugified to fit the [DNS Label](https://www.ietf.org/rfc/rfc1035.txt) requirements according to the [*namespace slugging procedure*]({{ "internals/names_slug_algorithm.html#basic-algorithm" | true_relative_url }}) that generates a unique and valid Kubernetes Namespace.

This is default behavior. It can be disabled by [setting `deploy.namespaceSlug=false`]({{ "/reference/werf_yaml.html#kubernetes-namespace" | true_relative_url }}) in the werf.yaml configuration.
