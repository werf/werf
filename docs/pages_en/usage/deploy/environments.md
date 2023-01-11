---
title: Multiple environments
permalink: usage/deploy/environments.html
---

## Environment

By default, werf assumes that each release should be tainted with some environment, such as `staging`, `test` or `production`.

Using this environment, werf determines:

1. Release name.
2. Kubernetes namespace.

The environment is a required parameter for deploying and should be specified either with an `--env` option or determined automatically using the data for the CI/CD system used.

### Using environment in templates

Current werf environment can be used in templates.

For example, you can use it to generate different templates for different environments:

{% raw %}
```
apiVersion: v1
kind: Secret
metadata:
  name: regsecret
type: kubernetes.io/dockerconfigjson
data:
{{ if eq .Values.werf.env "dev" }}
  .dockerconfigjson: UmVhbGx5IHJlYWxseSByZWVlZWVlZWVlZWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWxsbGxsbGxsbGxsbGxsbGxsbGxsbGxsbGxsbGxsbGx5eXl5eXl5eXl5eXl5eXl5eXl5eSBsbGxsbGxsbGxsbGxsbG9vb29vb29vb29vb29vb29vb29vb29vb29vb25ubm5ubm5ubm5ubm5ubm5ubm5ubm5ubmdnZ2dnZ2dnZ2dnZ2dnZ2dnZ2cgYXV0aCBrZXlzCg==
{{ else }}
  .dockerconfigjson: {{ .Values.dockerconfigjson }}
{{ end }}
```
{% endraw %}

It is important that `--env ENV` param value available not only in helm templates, but also [in `werf.yaml` templates]({{ "/reference/werf_yaml_template_engine.html#env" | true_relative_url }}).

More info about service values available [in the article]({{ "/usage/deploy/values.html" | true_relative_url }}).

## Kubernetes namespace

The Kubernetes namespace is constructed using the template `[[ project ]]-[[ env ]]` by default. Here, `[[ project ]]` refers to the [project name]({{ "reference/werf_yaml.html#project-name" | true_relative_url }}) and `[[ env ]]` refers to the environment.

For example, for the project named `symfony-demo`, there can be the following Kubernetes namespaces depending on the environment specified:

* `symfony-demo-stage` for the `stage` environment;
* `symfony-demo-test` for the `test` environment;
* `symfony-demo-prod` for the `prod` environment.

You can redefine the Kubernetes Namespace using the `--namespace NAMESPACE` deploy option. In that case, werf would use the specified name as is.

You can also define the custom Kubernetes Namespace in the werf.yaml configuration [by setting `deploy.namespace`]({{ "/reference/werf_yaml.html#kubernetes-namespace" | true_relative_url }}) parameter.

### Slugging Kubernetes namespace

The Kubernetes namespace that is constructed using the template will be slugified to fit the [DNS Label](https://www.ietf.org/rfc/rfc1035.txt) requirements that generates a unique and valid Kubernetes Namespace.

This is default behavior. It can be disabled by [setting `deploy.namespaceSlug=false`]({{ "/reference/werf_yaml.html#kubernetes-namespace" | true_relative_url }}) in the werf.yaml configuration.

## Multiple Kubernetes clusters

There are cases when separate Kubernetes clusters are required for a different environments. You can [configure access to multiple clusters](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters) using kube contexts in a single kube config.

In that case, the `--kube-context=CONTEXT` deploy option should be set manually along with the environment.
