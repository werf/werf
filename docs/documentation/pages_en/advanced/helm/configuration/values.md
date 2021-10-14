---
title: Values
permalink: advanced/helm/configuration/values.html
---

The Values is an arbitrary YAML file filled with parameters that you can use in [templates]({{ "/advanced/helm/configuration/templates.html" | true_relative_url }}). All values passed to the chart can be divided by following categories:

 - User-defined regular values.
 - User-defined secret values.
 - Service values.

## User-defined regular values

You can (optionally) place user-defined regular values into the `.helm/values.yaml` chart file. For example:

```yaml
global:
  names:
    - alpha
    - beta
    - gamma
  mysql:
    staging:
      user: mysql-staging
    production:
      user: mysql-production
    _default:
      user: mysql-dev
      password: mysql-dev
```

Values placed under the `global` key will be available both in the current chart and in all [subcharts]({{ "/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}).

Values placed under the arbitrary `SOMEKEY` key will be available in the current chart and in the `SOMEKEY` [subchart]({{ "advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}).

The `.helm/values.yaml` file is the default place to store values. You can also pass additional user-defined regular values via:

* Separate value files by specifying `--values=PATH_TO_FILE` (you can use it repeatedly to pass multiple files) as a werf option.
* Options `--set key1.key2.key3.array[0]=one`, `--set key1.key2.key3.array[1]=two` (can be used multiple times, [see also]({{ "/reference/cli/werf_helm_install.html" | true_relative_url }}) `--set-string key=forced_string_value`, and `--set-file key1.data.notes=notes.txt`).

**NOTE** All values files, including default `.helm/values.yaml` or any custom values file passed with `--values` option — all of these should be stored in the git repo of the project. More info in the [giterminism article]({{ "advanced/helm/configuration/giterminism.html" | true_relative_url }}).

## User-defined secret values

Secret values are perfect to store passwords and other sensitive data directly in the project repository.

You can (optionally) place user-defined secret values into the default `.helm/secret-values.yaml` chart file or any number of files with an arbitrary name (`--secret-values`). For example:

```yaml
global:
  mysql:
    production:
      password: 100024fe29e45bf00665d3399f7545f4af63f09cc39790c239e16b1d597842161123
    staging:
      password: 100024fe29e45bf00665d3399f7545f4af63f09cc39790c239e16b1d597842161123
```

Each value (like `100024fe29e45bf00665d3399f7545f4af63f09cc39790c239e16b1d597842161123`) in the secret value map is encoded by werf. The structure of the secret value map is the same as that of a regular value map (for example, in `values.yaml`). See more info [about secret value generation and working with secrets]({{ "/advanced/helm/configuration/secrets.html" | true_relative_url }}).

The `.helm/secret-values.yaml` file is the default place for storing secret values. You can also pass additional user-defined secret values via separate secret value files by specifying `--secret-values=PATH_TO_FILE` (can be used repeatedly to pass multiple files). All such files should be stored in the git repo of the project due to [giterminism]({{ "advanced/helm/configuration/giterminism.html" | true_relative_url }}).

**NOTE** All values files, including default `.helm/values.yaml` or any custom values file passed with `--values` option — all of these should be stored in the git repo of the project. More info in the [giterminism article]({{ "advanced/helm/configuration/giterminism.html" | true_relative_url }}).

### Set parameters

There is an ability to redefine or pass new values using cli params:

 - `--set KEY=VALUE`;
 - `--set-string KEY=VALUE`;
 - `--set-file=PATH`;
 - `--set-docker-config-json-value=true|false`.

**NOTE** All files, specified with `--set-file` option should be stored in the git repo of the project. More info in the [giterminism article]({{ "advanced/helm/configuration/giterminism.html" | true_relative_url }}).

#### set-docker-config-json-value

When `--set-docker-config-json-value` has been specified werf will set special value `.Values.dockerconfigjson` using current docker config from the environment where werf running (`DOCKER_CONFIG` is supported).

This value `.Values.dockerconfigjson` contains base64 encoded docker config, which is ready to use for example to create a secret to access a container registry:

{% raw %}
```
{{- if .Values.dockerconfigjson -}}
apiVersion: v1
kind: Secret
metadata:
  name: regsecret
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: {{ .Values.dockerconfigjson }}
{{- end -}}
```
{% endraw %}

**IMPORTANT** Current docker config may contain temporal login credentials created using temporal short-lived token (`CI_JOB_TOKEN` in GitLab for example) and in such case should not be used as `imagePullSecrets`, because registry will stop being available from inside Kubernetes as soon as the token is terminated.

## Service values

Service values are set by the werf to pass additional data when rendering chart templates.

Here is an example of service data:

```yaml
werf:
  name: myapp
  env: production
  repo: registry.domain.com/apps/myapp
  image:
    assets: registry.domain.com/apps/myapp/assets:a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
    rails: registry.domain.com/apps/myapp/rails:e760e9311f938e3d92681e93da3a81e176aa7f7e684ee06d092ec199-1598269478292

global:
  werf:
    name: myapp
    version: v1.2.7
```

There are following service values:
 - The name of the werf project: `.Values.werf.name`.
 - Version of the werf cli util: `.Values.werf.version`.
 - Name of a CI/CD environment used during the current deploy process: `.Values.werf.env`.
 - Container registry repo used during the current deploy process: `.Values.werf.repo`.
 - Full images names used during the current deploy process: `.Values.werf.image.NAME`. More info about using this available in [the templates article]({{ "/advanced/helm/configuration/templates.html#integration-with-built-images" | true_relative_url }}).

### Service values in the subcharts

If you are using [subcharts]({{ "/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}) and you want to use regular werf service values (those that defined outside of `global` scope) in subchart, you need to explicitly export these parent-scoped service values from the parent to the subchart:

```yaml
# .helm/Chart.yaml
apiVersion: v2
dependencies:
  - name: rails
    version: 1.0.0
    export-values:
    - parent: werf
      child: werf
```

Now the service values initially available only at `.Values.werf` of the parent chart will be also available on the same path `.Values.werf` in the subchart named "rails". Refer to these service values in the subchart like this:

{% raw %}
```yaml
# .helm/charts/rails/app.yaml
...
spec:
  template:
    spec:
      containers:
      - name: rails
        image: {{ .Values.werf.image.rails }}  # Will result in: `image: registry.domain.com/apps/myapp/rails:e760e931...`
```
{% endraw %}

Path to where you want to export service values can be changed:

```yaml
    export-values:
    - parent: werf
      child: definitely.not.werf  # Service values will become available on `.Values.definitely.not.werf` in the subchart.
```

Or pass service values one by one to the subchart:

```yaml
# .helm/Chart.yaml
apiVersion: v2
dependencies:
  - name: postgresql
    version: "10.9.4"
    repository: "https://charts.bitnami.com/bitnami"
    export-values:
    - parent: werf.repo
      child: image.repository
    - parent: werf.tag.my-postgresql
      child: image.tag
```

More info about `export-values` can be found [here]({{ "advanced/helm/configuration/chart_dependencies.html#mapping-values-from-parent-chart-to-subcharts" | true_relative_url }}).

## Merging the resulting values

<!-- This section could be in internals -->

During the deployment process, werf merges all user-defined regular, secret, and service values into the single value map, which is then passed to the template rendering engine to be used in the templates (see [how to use values in the templates](#using-values-in-the-templates)). Values are merged in the following order of priority (the more recent value overwrites the previous one):

 1. User-defined regular values from the `.helm/values.yaml`.
 2. User-defined regular values from all cli options `--values=PATH_TO_FILE` in the order specified.
 3. User-defined secret values from `.helm/secret-values.yaml`.
 4. User-defined secret values from all cli options `--secret-values=PATH_TO_FILE` in the order specified.
 5. User-defined values passed with the `--set*` params.
 6. Service values.

## Using values in the templates

werf uses the following syntax for accessing values contained in the chart templates:

{% raw %}
```yaml
{{ .Values.key.key.arraykey[INDEX].key }}
```
{% endraw %}

The `.Values` object contains the [merged map of resulting values](#merging-the-resulting-values) map.
