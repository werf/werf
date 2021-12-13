---
title: Templates
permalink: advanced/helm/configuration/templates.html
---

Kubernetes resources definitions are placed in the `.helm/templates` directory and called templates.

This directory contains `*.yaml` YAML files, arbitrary nested folders are acceptable. Each YAML file describes one or several Kubernetes resources separated by three hyphens `---`, for example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mydeploy
  labels:
    service: mydeploy
spec:
  selector:
    matchLabels:
      service: mydeploy
  template:
    metadata:
      labels:
        service: mydeploy
    spec:
      containers:
      - name: main
        image: ubuntu:18.04
        command: [ "/bin/bash", "-c", "while true; do date ; sleep 1 ; done" ]
---
apiVersion: v1
kind: ConfigMap
  metadata:
    name: mycm
  data:
    node.conf: |
      port 6379
      loglevel notice
```

Each YAML file is also preprocessed using [Go templates](https://golang.org/pkg/text/template/#hdr-Actions).

Using go templates, the user can:

* generate different Kubernetes resources and their components depending on arbitrary conditions;
* parameterize templates with [values]({{ "/advanced/helm/configuration/values.html" | true_relative_url }}) for different environments;
* extract common text parts, save them as Go templates, and reuse in several places;
* etc.

You can use [Sprig functions](https://masterminds.github.io/sprig/) and [advanced functions](https://helm.sh/docs/howto/charts_tips_and_tricks/) such as `include` and `required` in addition to basic functions of Go templates.

Also, the user can place `*.tpl` files into the `.helm/templates` directory or any subdirectories. These will not be rendered into the Kubernetes object. These files can be used to store Go templates. You can insert templates contained the `*.tpl` files into the `*.yaml` files.

## Integration with built images

Kubernetes resources needs a full docker image name, including the docker repo and the docker tag, in order to use the docker image in the chart resource specifications. But how do you designate an image contained in the `werf.yaml` file given that the full docker image name for such an image depends on the specified image repository and werf content based tagging?

werf provides a pack of service values, which contain `.Values.werf.image` map, which contain mapping of docker images names by `werf.yaml` image short name. Full description of werf's service values is available in the [values article]({{ "/advanced/helm/configuration/values.html" | true_relative_url }}).

### .Values.werf.image

```
map[string]string

SHORT_IMAGE_NAME => FULL_DOCKER_IMAGE_NAME
```

This map contains full image name which can be used as a value for `image` key in the container section of a pod spec.

##### Examples

To retrieve docker image for an image named `backend` in the `werf.yaml`, use:

```
.Values.werf.image.backend
```

To retrieve docker image for an image named `nginx-assets` in the `werf.yaml`, use:

```
index .Values.werf.image "nginx-assets"
```

Here is how you can refer to an image called `backend` described in `werf.yaml`:

{% raw %}
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  labels:
    service: backend
spec:
  selector:
    matchLabels:
      service: backend
  template:
    metadata:
      labels:
        service: backend
    spec:
      containers:
      - name: main
        command: [ ... ]
        image: {{ .Values.werf.image.backend }}
```
{% endraw %}

## Builtin templates and params

{% raw %}
* `{{ .Chart.Name }}` — contains the [project name] specified in the `werf.yaml` config or chart name explicitly defined in the `.helm/Chart.yaml`.
* `{{ .Release.Name }}` — {% endraw %}contains the [release name]({{ "/advanced/helm/releases/release.html" | true_relative_url }}).{% raw %}
* `{{ .Files.Get }}` — a function to read file contents into templates; requires the path to file as an argument. The path should be specified relative to the `.helm` directory (files outside the `.helm` folder are ignored).
{% endraw %}

### Environment

All main werf commands (such as [`werf converge`]({{ "reference/cli/werf_converge.html" | true_relative_url }}), [`werf build`]({{ "reference/cli/werf_build.html" | true_relative_url }}), [`werf render`]({{ "reference/cli/werf_render.html" | true_relative_url }}) etc.).

This param could affect [release name]({{ "/advanced/helm/releases/naming.html" | true_relative_url }}) and [kubernetes namespace]({{ "/advanced/helm/releases/naming.html" | true_relative_url }}). Also it is possible to access environment in templates:

{% raw %}
```
{{ .Values.werf.env }}
```
{% endraw %}

It is acceptable to use this param in templates to generate different templates for different environments. For example:

{% raw %}
```
{{ if eq .Values.werf.env "dev" }}
apiVersion: v1
kind: Secret
metadata:
  name: regsecret
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: UmVhbGx5IHJlYWxseSByZWVlZWVlZWVlZWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWxsbGxsbGxsbGxsbGxsbGxsbGxsbGxsbGxsbGxsbGx5eXl5eXl5eXl5eXl5eXl5eXl5eSBsbGxsbGxsbGxsbGxsbG9vb29vb29vb29vb29vb29vb29vb29vb29vb25ubm5ubm5ubm5ubm5ubm5ubm5ubm5ubmdnZ2dnZ2dnZ2dnZ2dnZ2dnZ2cgYXV0aCBrZXlzCg==
---
{{ else }}
apiVersion: v1
kind: Secret
metadata:
  name: regsecret
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: {{ .Values.dockerconfigjson }}
---
{{ end }}
```
{% endraw %}

It is important that `--env ENV` param value available not only in helm templates, but also [in `werf.yaml` templates]({{ "/reference/werf_yaml_template_engine.html#env" | true_relative_url }}).

More info about service values available [in the article]({{ "/advanced/helm/configuration/values.html" | true_relative_url }}).
