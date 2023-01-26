---
title: Charts and dependencies
permalink: usage/deploy/charts.html
---

## What is a chart?

The chart in werf is a Helm chart with some additional features. In turn, a Helm chart is a distributable package with Helm templates, values files, and some metadata. The chart is used to generate ready-to-use Kubernetes manifests for deploying the application.

A typical chart looks as follows:

```
chartname/
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

- `templates/*.yaml` — contains to Helm templates that are used to generate Kubernetes manifests;

- `templates/*.tpl` — contains Helm pattern files for use in other Helm templates. The result of templating these files is ignored;

- `templates/NOTES.txt` — werf displays the contents of this file in the terminal at the end of each successful deployment;

- `crds/*.yaml` — [Custom Resource Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#customresourcedefinitions) to deploy before deploying manifests in `templates`;

- `secret/*` — (werf only) encrypted files; their decrypted contents can be inserted into Helm templates;

- `values.yaml` — contains files with declarative configurations to use in Helm templates. The configurations can be overridden by environment variables, command line arguments, or other values files;

- `values.schema.json` — the JSON schema to validate `values.yaml`;

- `secret-values.yaml` — (werf only) an encrypted declarative configuration file similar to `values.yaml`. Its decrypted contents are merged with `values.yaml` during manifest generation;

- `Chart.yaml` — contains the main configuration and the chart metadata;

- `Chart.lock` — the lock file protects from unwanted changes/updates to dependent charts;

- `README.md` — the chart documentation;

- `LICENSE` — the chart license;

- `.helmignore` — a list of files in the chart directory not to be included in the chart when publishing.

### The main chart

When running commands like `werf converge` or `werf render`, werf by default uses a chart in the `<root Git repository>/.helm` directory. This chart is called the main chart. The main chart directory can be changed using the `deploy.helmChartDir` directive in `werf.yaml`.

The main chart, unlike a typical chart, may not include a `Chart.yaml` file. In this case, the following `.helm/Chart.yaml` is used:

```yaml
# .helm/Chart.yaml:
apiVersion: v2
name: <werf project name>
version: 1.0.0
```

To modify the directives in `.helm/Chart.yaml' or to add new ones, create the file `.helm/Chart.yaml' manually by adding/redefining the directives of interest.

To use a chart from the OCI/HTTP repository instead of a local one or to deploy multiple charts at once, specify the charts of interest as dependencies for the main chart. Use the `dependencies` directive of the `Chart.yaml` file to do this.

### Chart configuration

`Chart.yaml` is the main configuration file of the chart. It contains the name, version, and other chart parameters, as well as references to dependent charts, for example: 

```yaml
# Chart.yaml:
apiVersion: v2
name: mychart
version: 1.0.0-anything-here

# Optional:
type: application
kubeVersion: "~1.20.0"
dependencies:
  - name: nginx
    version: 1.2.3
    repository: https://example.com/charts

# Optional info directives:
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

Mandatory directives:

* `apiVersion` — the format of `Chart.yaml`; either `v2` (recommended) or `v1`;

* `name` — the chart name;

* `version` — the chart version as per ["Semantic Versioning 2.0"](https://semver.org/spec/v2.0.0.html);

Optional directives:

* `type` — the chart type;

* `kubeVersion` — compatible Kubernetes versions;

* `dependencies` — the description of the dependent charts;

Optional info directives:

* `appVersion` — the version of the application the chart installs (in an arbitrary format);

* `deprecated` — tells whether the chart is deprecated, `false` (default) or `true`;

* `icon` — the URL to the chart icon;

* `description` — the chart description;

* `home` — the chart website;

* `sources` — a list of repositories with the chart contents;

* `keywords` — a list of keywords associated with the chart;

* `annotations` — extra optional information about the chart;

* `maintainers`  — a list of chart maintainers;

### The chart type

The chart type is specified in the `type` directive of the `Chart.yaml` file. The valid types are:

* `application` — the regular chart (no limitations);

* `library` — this chart contains templates only and does not generate any manifests on its own. Such a chart cannot be set, it can only be used as a dependent one.

### The compatible Kubernetes versions

You can specify which Kubernetes versions this chart is compatible with using the `kubeVersion` directive, e.g.:

```yaml
# Chart.yaml:
name: mychart
kubeVersion: "1.20.0"   # Compatible with 1.20.0
kubeVersion: "~1.20.3"  # Compatible with all 1.20.x starting with 1.20.3
kubeVersion: "~1"       # Compatible with all 1.x.x
kubeVersion: ">= 1.20.3 < 2.0.0 && != 1.20.4"
```

## Dependencies on other charts

A chart can depend on other charts. In this case, manifests are generated for both parent and dependent charts and then merged together for deployment. Dependent charts can be configured using the `dependencies` directive of the `Chart.yaml` file, e.g.:

```yaml
# Chart.yaml:
dependencies:
- name: backend
  version: "~1.2.3"
  repository: https://example.com/charts
```

We do not recommend specifying dependent charts in `requirements.yaml` instead of `Chart.yaml`. The `requirements.yaml` file has become obsolete.

### Name, alias

`name` — the original name of the dependent chart as set by the chart developer.

You can connect multiple charts with the same `name` or the same chart several times using the `alias` directive to change the names of the connected charts, for example:

```yaml
# Chart.yaml:
dependencies:
- name: backend
  alias: main-backend
- name: backend
  alias: secondary-backend
```

### Version

`version` — limits suitable versions of the dependent chart, e.g.:

```yaml
# Chart.yaml:
dependencies:
- name: backend
  version: "1.2.3"   # Use version 1.2.3
  version: "~1.2.3"  # Use the latest 1.2.x version but no lower than 1.2.3
  version: "~1"      # Use the latest 1.x.x version
  version: "^1.2.3"  # Use the latest 1.x.x version but no lower than 1.2.3
  version: ">= 1.2.3 < 2.0.0 && != 1.2.4"
```

For information on syntax, see [checking version constraints](https://github.com/Masterminds/semver#checking-version-constraints).

### Repository

`repository` — the path to the chart source where the specified dependent chart can be found, for example:

```yaml
# Chart.yaml:
dependencies:  
- name: mychart
  # OCI repository (recommended):
  repository: oci://example.org/myrepo
  # HTTP repository:
  repository: https://example.org/myrepo
  # Short repository name (if the repository is added using
  # `werf helm repo add shortreponame`):
  repository: alias:myrepo
  # Short repository name (alternative syntax)
  repository: "@myrepo"
  # Local chart from an arbitrary location:
  repository: file://../mychart
  # For a local chart in "charts/" the repository is omitted:
  # repository:
```

### Condition, tags

By default, all dependent charts are enabled. You can use the `condition` directive to enable/disable charts:

```yaml
# Chart.yaml:
dependencies:
- name: backend 
  # The chart will be enabled if .Values.backend.enabled == true:
  condition: backend.enabled  
```

You can also use the `tags` directive to enable/disable groups of charts all at once:

```yaml
# Chart.yaml:
dependencies:
# Enable backend if .Values.tags.app == true:
- name: backend
  tags: ["app"]
# Enable frontend if .Values.tags.app == true:
- name: frontend
  tags: ["app"]
# Enable database if .Values.tags.database == true:
- name: database
  tags: ["database"]
```

### Export-values, import-values

`export-values` — (werf only) automatically passes specified Values from the parent chart to the dependent one; for example:

```yaml
# Chart.yaml:
dependencies:
- name: backend
  export-values:
  # Substitute .Values.werf into .Values.backend.werf automatically:
  - parent: werf
    child: werf
  # Substitute .Values.exports.myMap.* into .Values.backend.* automatically:
  - parent: exports.myMap
    child: .
  # A shorter version of the above export expression:  
  - myMap
```

`import-values` — automatically passes the specified Values from the dependent chart to the parent one, i.e., in the opposite direction; for example:

```yaml
# Chart.yaml:
dependencies:
- name: backend
  import-values:
  # Substitute .Values.backend.from into .Values.to automatically:
  - child: from
    parent: to
  # Substitute .Values.backend.exports.myMap.* into .Values.* automatically:
  - child: exports.myMap
    parent: .
  # A shorter version of the above import expression:  
  - myMap
```

### Adding/updating dependent charts

The recommended way to add/update dependent charts is as follows:

1. Add/update the dependent chart configuration to `Chart.yaml`;

2. If a dependent chart is in the private OCI or HTTP repository, add the OCI or HTTP repository manually using `werf helm repo add` and provide all the options required to access it;

3. Run `werf helm dependency update` — it will update `Chart.lock`;

4. Commit the updated `Chart.yaml` and `Chart.lock` to Git.

It is also recommended to add `.helm/charts/**.tgz` to `.gitignore`.
