---
title: Charts and dependencies
permalink: usage/deploy/charts.html
---

## What is a chart?

The chart in werf is a Helm chart with some additional features. In turn, a Helm chart is a distributable package with Helm templates, values files, and some metadata. The chart is used to generate ready-to-use Kubernetes manifests for deploying the application.

## Creating a new chart

When deploying with `werf converge` or publishing with `werf render`, werf uses a `.helm` chart in the Git repository's root. directory. This chart is called the *main* chart. The main chart directory can be changed using the `deploy.helmChartDir` parameter in `werf.yaml`.

A regular chart requires a `Chart.yaml` file with the chart name and version:

```yaml
# Chart.yaml:
apiVersion: v2
name: mychart
version: 1.2.3
```

However, the main chart does not require it, because if there is no `Chart.yaml` file available or if there is no chart name or version in the file, the following default parameters will be used:

```yaml
# .helm/Chart.yaml:
apiVersion: v2
name: <werf project name>
version: 1.0.0
```

If necessary, you can redefine the parameters of the main chart. To do so, create a file `.helm/Chart.yaml` and specify the necessary parameters in it.

In case your chart only includes named templates for use in other charts, add the `type: library` directive to `Chart.yaml`:

```yaml
# Chart.yaml:
type: library
```

You can specify a list of compatible Kubernetes versions for the cluster into which the chart is deployed using the `kubeVersion` directive:

```yaml
# Chart.yaml:
kubeVersion: "~1.20.3"
```

You may also add the following informational directives

```yaml
# Chart.yaml:
appVersion: "1.0"
deprecated: false
icon: https://example.org/mychart-icon.svg
description: This is My Chart
home: https://example.org
sources:
  - https://github.com/my/chart
keywords:
  - apps
annotations:
  anyAdditionalInfo: here
maintainters:
  - name: John Doe
    email: john@example.org
    url: https://john.example.org
```

The resulting chart may be deployed or published, although there is little use of it as it is. You will have to add templates to the `templates` directory or include dependent charts to make it more useful.

## Adding files to a chart

Add templates, parameters, dependent charts, and more to the chart as needed. A typical main chart looks as follows:

```
.helm/
  charts/
    dependent-chart/
      # ...
  templates/
    deployment.yaml
    _helpers.tpl
    NOTES.txt
  crds/
    crd.yaml
  secret/                   # only available in werf
    some-secret-file
  values.yaml
  values.schema.json
  secret-values.yaml        # only available in werf
  Chart.yaml
  Chart.lock
  README.md
  LICENSE
  .helmignore
```

Here:

- `charts/*` — contains dependent charts whose Helm templates/values files are used to generate manifests together with the Helm templates/values files of the parent chart;

- `templates/*.yaml` — contains Helm templates that are used to generate Kubernetes manifests;

- `templates/*.tpl` — contains Helm templates for use in other Helm templates. The result of templating these files is ignored;

- `templates/NOTES.txt` — werf displays the contents of this file in the terminal at the end of each successful deployment;

- `crds/*.yaml` — [Custom Resource Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#customresourcedefinitions) to deploy before deploying manifests in `templates`;

- `secret/*` — (werf only) сontains encrypted files; their decrypted contents can be inserted into the Helm templates;

- `values.yaml` — contains files with declarative configurations to use in Helm templates. The configurations can be overridden by environment variables, command line arguments, or other values files;

- `values.schema.json` — the JSON schema to validate `values.yaml`;

- `secret-values.yaml` — (werf only) an encrypted declarative configuration file similar to `values.yaml`. Its decrypted contents are merged with `values.yaml` during manifest generation;

- `Chart.yaml` — contains the main configuration and the chart metadata;

- `Chart.lock` — the lock file protects from unwanted changes/updates to dependent charts;

- `README.md` — the chart documentation;

- `LICENSE` — the chart license;

- `.helmignore` — a list of files in the chart directory not to be included in the chart when publishing.

## Including additional charts

### Including dependent local charts

You can add dependent local charts by placing them in the `charts` directory of the parent chart. In this case, manifests will be generated for both parent and dependent charts, and then merged together for further deployment. No additional configuration is required.

Here is an example of a main chart with local dependent charts:

```
.helm/
  charts/
    postgresql/
      templates/
        postgresql.yaml
      Chart.yaml
    redis/
      templates/
        redis.yaml
      Chart.yaml
  templates/
    backend.yaml
```

> Note that local dependent charts *must* have the same name as their directory name.

If the local dependent chart needs some additional parameters, specify the name of the dependent chart in the `Chart.yaml` file of the parent chart without providing `dependencies[].repository` and add the desired parameters as follows:

```yaml
# .helm/Chart.yaml:
apiVersion: v2
dependencies:
- name: redis
  condition: redis.enabled
```

Use the `dependencies[].repository` parameter to include a local chart from a location other than the `charts` directory:

```yaml
# .helm/Chart.yaml:
apiVersion: v2
dependencies:
- name: redis
  repository: file://../redis
```
### Including dependent charts that come from the repository

You can include additional charts from the OCI/HTTP repository by listing them as dependencies in the `dependencies` parameter of the `Chart.yaml` file of the parent chart. In this case, manifests will be generated for both parent and dependent charts and then merged together for further deployment.

Here is an example of using a chart that comes from a repository and depends on the main chart:

```yaml
# .helm/Chart.yaml:
apiVersion: v2
dependencies:
- name: database
  version: "~1.2.3"
  repository: https://example.com/charts
```

After each addition/update of remote dependent charts or change in their configuration you need to:

1. (If using a private OCI or HTTP repository with a dependent chart) Add the OCI or HTTP repository manually with `werf helm repo add` and provide the necessary options to access the repository.

2. Run `werf helm dependency update`, which will update `Chart.lock`.

3. Commit the updated `Chart.yaml` and `Chart.lock` to Git.

We also recommend adding `.helm/charts/**.tgz` to `.gitignore` when using charts from the repository.

### Specifying the name of the chart to include

You can use the directive `dependencies[].name` of the parent chart to specify the original name of the dependent chart as set by its developer, for example:

```yaml
# .helm/Chart.yaml:
apiVersion: v2
dependencies:
- name: backend
```

If you want to connect multiple dependent charts with the same name or connect the same dependent chart several times, use the parent chart's `dependencies[].alias' directive to add alias for the charts to be included, for example:

```yaml
# .helm/Chart.yaml:
apiVersion: v2
dependencies:
- name: backend
  alias: main-backend
- name: backend
  alias: secondary-backend
```

### Specifying the version of the chart to include

You can limit the applicable versions of the dependent chart (werf will chose the most recent one) using the `dependencies[].version` directive of the parent chart, for example:

```yaml
# .helm/Chart.yaml:
apiVersion: v2
dependencies:
- name: backend
  version: "~1.2.3"
```

In this case, the latest 1.2.x version will be used, with 1.2.3 as the lowest possible one.

### Specifying the source of the chart to include

You can specify the path to the source where the specified dependent chart can be found using the `dependencies[].repository` directive of the parent chart as follows:

```yaml
# .helm/Chart.yaml:
apiVersion: v2
dependencies:
- name: mychart
  repository: oci://example.org/myrepo
```

In this case, the werf will use the `mychart` chart from the `example.org/myrepo` OCI repository.

### Enabling/disabling dependent charts

All dependent charts are enabled by default. You can use the `dependencies[].condition` directive of the parent chart to easily enable/disable dependent charts, for example:

```yaml
# .helm/Chart.yaml:
apiVersion: v2
dependencies:
- name: backend
  condition: backend.enabled
```

In this case, the `backend` dependent cahrt will be enabled only if `$.Values.backend.enabled` is `true` (it is set to `true` by default).

You can also use the `dependencies[].tags` directive of the parent chart to enable/disable a whole group of dependent charts at once, for example:

```yaml
# .helm/Chart.yaml:
dependencies:
- name: backend
  tags: ["app"]
- name: frontend
  tags: ["app"]
```

In this case, the `backend` and `frontend` dependent charts will only be enabled if the `$.Values.tags.app` parameter is set to `true` (default is `true`).
