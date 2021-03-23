---
title: Chart dependencies
permalink: advanced/helm/configuration/chart_dependencies.html
---

**Subchart** is a helm chart that included into the current chart as a dependency. werf allows usage of subcharts the same way [as helm](https://helm.sh/docs/topics/charts/). The chart can include arbitrary number of subcharts. Usage of werf project itself as a subchart in another werf project is not supported for now.

Subcharts are placed in the directory `.helm/charts/SUBCHART_DIR`. Each subchart in the `SUBCHART_DIR` is a chart by itself with the similar files structure (which can also have own recursive subcharts).

During deploy process werf will render, create and track all resources of all subcharts.

## Enable subchart for your project

 1. Let's include a `redis` as a dependency for our werf chart using `.helm/Chart.yaml` file:

     ```yaml
     # .helm/Chart.yaml
     apiVersion: v2
     dependencies:
       - name: redis
         version: "12.7.4"
         repository: "https://charts.bitnami.com/bitnami"
      ```

     **NOTE** It is not required to define full `Chart.yaml` with name and version fields as for the standard helm. werf will autogenerate chart name and version based on the `werf.yaml` `project` field settings. See more info in the [chart article]({{ "/advanced/helm/configuration/chart.html" | true_relative_url }}).

 2. Next it is required to generate `.helm/Chart.lock` using `werf helm dependency update` command from the root of the project:

     ```shell
     werf helm dependency update .helm
     ```

     This command will generate `.helm/Chart.lock` file as well as download all dependencies into the `.helm/charts` directory.
     
 3. `.helm/Chart.lock` file should be committed into the git repository, while `.helm/charts` could be added to the `.gitignore`.

Later during deploy process (with [`werf converge` command]({{ "/reference/cli/werf_converge.html" | true_relative_url }}) or [`werf bundle apply` command]({{ "/reference/cli/werf_bundle_apply.html" | true_relative_url }})), or during templates rendering (with [`werf render` command]({{ "/reference/cli/werf_render.html" | true_relative_url }})) werf will automatically download all dependencies specified in the lock file `.helm/Chart.lock`.

**NOTE** `.helm/Chart.lock` file must be committed into the git repo, more info in the [giterminism article]({{ "/advanced/helm/configuration/giterminism.html#subcharts-and-giterminism" | true_relative_url }}).

## Dependencies configuration

<!-- Move to reference -->

Let's describe format of the `.helm/Chart.yaml` dependencies.

* The `name` should be the name of a chart, where that name must match the name in that chart's 'Chart.yaml' file.
* The `version` field should contain a semantic version or version range.
* The `repository` URL should point to a **Chart Repository**. Helm expects that by appending `/index.yaml` to the URL, it should be able to retrieve the chart repository's index. The `repository` can be an alias. The alias must start with `alias:` or `@`.

The `.helm/Chart.lock` file lists the exact versions of immediate dependencies and their dependencies and so forth.

The `werf helm dependency` commands operate on that file, making it easy to synchronize between the desired dependencies and the actual dependencies stored in the `charts` directory:
* Use [`werf helm dependency list`]({{ "reference/cli/werf_helm_dependency_list.html" | true_relative_url }}) to check dependencies and their statuses.
* Use [`werf helm dependency update`]({{ "reference/cli/werf_helm_dependency_update.html" | true_relative_url }}) to update `.helm/charts` based on the contents of `.helm/Chart.yaml`.
* Use [`werf helm dependency build`]({{ "reference/cli/werf_helm_dependency_build.html" | true_relative_url }}) to update `.helm/charts` based on the `.helm/Chart.lock` file.

All Chart Repositories that are used in `.helm/Chart.yaml` should be configured on the system. The `werf helm repo` commands can be used to interact with Chart Repositories:
* Use [`werf helm repo add`]({{ "reference/cli/werf_helm_repo_add.html" | true_relative_url }}) to add Chart Repository.
* Use [`werf helm repo index`]({{ "reference/cli/werf_helm_repo_index.html" | true_relative_url }}).
* Use [`werf helm repo list`]({{ "reference/cli/werf_helm_repo_list.html" | true_relative_url }}) to list existing Chart Repositories.
* Use [`werf helm repo remove`]({{ "reference/cli/werf_helm_repo_remove.html" | true_relative_url }}) to remove Chart Repository.
* Use [`werf helm repo update`]({{ "reference/cli/werf_helm_repo_update.html" | true_relative_url }}) to update local Chart Repositories indexes.

werf is compatible with Helm settings, so by default `werf helm dependency` and `werf helm repo` commands use settings from **helm home folder**, `~/.helm`. But you can change it with `--helm-home` option. If you do not have **helm home folder** or want to create another one use `werf helm repo init` command to initialize necessary settings and configure default Chart Repositories.

## Subchart and values

To pass values from parent chart to subchart called `mysubchart` user must define following values in the parent chart:

```yaml
mysubchart:
  key1:
    key2:
    - key3: value
```

In the `mysubchart` these values should be specified without `mysubchart` key:

{% raw %}
```yaml
{{ .Values.key1.key2[0].key3 }}
```
{% endraw %}

Global values defined by the special toplevel values key `global` will also be available in the subcharts:

```yaml
global:
  database:
    mysql:
      user: user
      password: password
```

In the subcharts these values should be specified as always:

{% raw %}
```yaml
{{ .Values.global.database.mysql.user }}
```
{% endraw %}

Only values by keys `mysubchart` and `global` will be available in the subchart `mysubchart`.

**NOTE** `secret-values.yaml` files from subcharts will not be used during deploy process. Although secret values from main chart and additional secret values from cli params `--secret-values` will be available in the `.Values` as usually.

## Obsolete requirements.yaml and requirements.lock

An older way of storing dependencies in the `.helm/requirements.yaml` and `.helm/requirements.lock` files is also supported by the werf, but it is recommended to use `.helm/Chart.yaml` and `.helm/Chart.lock` instead. 
