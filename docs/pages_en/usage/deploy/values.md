---
title: Templates parametrization
permalink: usage/deploy/values.html
---

## Parameterization basics

The contents of the `$.Values` dictionary can be used to parameterize templates. Each chart has its own `$.Values` dictionary. The dictionary is compiled by merging parameters obtained from parameter files, command line options and other sources.

Below is a quick example of parameterization using `values.yaml`:

```yaml
# values.yaml:
myparam: myvalue
```

{% raw %}

```
# templates/example.yaml:
{{ $.Values.myparam }}
```

{% endraw %}

Output:

```yaml
myvalue
```

And here is a more complex example:

```yaml
# values.yaml:
myparams:
- value: original
```

{% raw %}

```
# templates/example.yaml:
{{ (index $.Values.myparams 0).value }}
```

{% endraw %}

```
werf render --set myparams[0].value=overriden
```

Output:

```yaml
overriden
```

## Parameter sources and their priority

The `$.Values` dictionary is compiled by combining the parameters from the parameter sources in the following order:

1. `values.yaml` of the current chart.
2. `secret-values.yaml` of the current chart (werf only).
3. The dictionary in `values.yaml` of the parent chart, whose key is the alias or name of the current chart.
4. The dictionary in `secret-values.yaml` of the parent chart (werf only), whose key is the alias or name of the current chart.
5. Parameter files stored in the `WERF_VALUES_*` variable.
6. Parameter files set by the `--values` option.
7. Secret parameter files stored in the `WERF_SECRET_VALUES_*` variable.
8. Secret parameter files set by the `-secret-values` option.
9. Parameters in the set files stored in the `WERF_SET_FILE_*` variable.
10. Parameters in the set files set by the `-set-file` option.
11. Parameters stored in the `WERF_SET_STRING_*` variable.
12. Parameters set by the `--set-string` option.
13. Parameters stored in the `WERF_SET_*` variable.
14. Parameters set by the `--set` option.
15. werf auxiliary parameters.
16. Parameters taken from the `export-values` directive of the parent chart (werf only).
17. Parameters taken from the `import-values` directive of the child charts.

The rules for combining parameters:

* basic data types are overwritten;

* lists are overwritten;

* dictionaries are merged;

* in case of a conflict, the parameters from the sources located higher in the list are overwritten by the parameters from the sources located lower in the list.

## Parameterizing the chart

The chart can be parameterized using its parameter file:

```yaml
# values.yaml:
myparam: myvalue
```

{% raw %}

```
# templates/example.yaml:
{{ $.Values.myparam }}
```

{% endraw %}

Output:

```
myvalue
```

You can also add/override the chart parameters using command line arguments:

```shell
werf render --set myparam=overriden  # or WERF_SET_MYPARAM=myparam=overriden werf render
```

```shell
werf render --set-string myparam=overriden  # or WERF_SET_STRING_MYPARAM=myparam=overriden werf render
```

... as well as by additional parameter files:

```yaml
# .helm/values-production.yaml:
myparam: overriden
```

```shell
werf render --values .helm/values-production.yaml  # or WERF_VALUES_PROD=.helm/values-production.yaml werf render
```

... or by a secret parameter file of the main chart (werf only):

```yaml
# .helm/secret-values.yaml:
myparam: <encrypted>
```

```shell
werf render
```

... or by additional secret parameter files of the main chart (werf only):

```yaml
# .helm/secret-values-production.yaml:
myparam: <encrypted>
```

```shell
werf render --secret-values .helm/secret-values-production.yaml  # or WERF_SECRET_VALUES_PROD=.helm/secret-values-production.yaml werf render
```

... or by set files:

```
# myparam.txt:
overriden
```

```shell
werf render --set-file myparam=myparam.txt  # or WERF_SET_FILE_PROD=myparam=myparam.txt werf render
```

The output is the same:

```
overriden
```

## Parameterizing dependent charts

A dependent chart can be parameterized both using its own parameter file and the parameter file of the parent chart.

In the example below, the parameters from the `mychild` dictionary in the `values.yaml` file of the `myparent` chart override the parameters in the `values.yaml` file of the `mychild` chart:

```yaml
# Chart.yaml:
name: myparent
dependencies:
- name: mychild
```

```yaml
# values.yaml:
mychild:
  myparam: overriden
```

```yaml
# charts/mychild/values.yaml:
myparam: original
```

{% raw %}

```
# charts/mychild/templates/example.yaml:
{{ $.Values.myparam }}
```

{% endraw %}

Output:

```
overriden
```

Note that the dictionary in `values.yaml` of the parent chart that contains parameters for the dependent chart must have `alias` (if any) or `name` of the dependent chart as its name.

You can also add/override parameters of a dependent chart using command line arguments:

```shell
werf render --set mychild.myparam=overriden  # or WERF_SET_MYPARAM=mychild.myparam=overriden werf render
```

```shell
werf render --set-string mychild.myparam=overriden  # or WERF_SET_STRING_MYPARAM=mychild.myparam=overriden werf render
```

... as well as by additional parameter files:

```yaml
# .helm/values-production.yaml:
mychild:
  myparam: overriden
```

```shell
werf render --values .helm/values-production.yaml  # or WERF_VALUES_PROD=.helm/values-production.yaml werf render
```

... or by a secret parameter file of the main chart (werf only):

```yaml
# .helm/secret-values.yaml:
mychild:
  myparam: <encrypted>
```

```shell
werf render
```

or by additional secret parameter files of the main chart (werf only):

```yaml
# .helm/secret-values-production.yaml:
mychild:
  myparam: <encrypted>
```

```shell
werf render --secret-values .helm/secret-values-production.yaml  # or WERF_SECRET_VALUES_PROD=.helm/secret-values-production.yaml werf render
```

... or using set files:

```
# mychild-myparam.txt:
overriden
```

```shell
werf render --set-file mychild.myparam=mychild-myparam.txt  # or WERF_SET_FILE_PROD=mychild.myparam=mychild-myparam.txt werf render
```

... or vie the `export-values` directive (werf only):

```yaml
# Chart.yaml:
name: myparent
dependencies:
- name: mychild
  export-values:
  - parent: myparam
    child: myparam
```

```yaml
# values.yaml:
myparam: overriden
```

```shell
werf render
```

The output will be the same:

```
overriden
```

## Using the dependent chart parameters in the parent chart

You can use the `import-values` directive in the parent chart to pass the parameters of the dependent chart to the parent one:

```yaml
# Chart.yaml:
name: myparent
dependencies:
- name: mychild
  import-values:
  - child: myparam
    parent: myparam
```

```yaml
# values.yaml:
myparam: original
```

```yaml
# charts/mychild/values.yaml:
myparam: overriden
```

{% raw %}

```
# templates/example.yaml:
{{ $.Values.myparam }}
```

{% endraw %}

Output:

```
overriden
```

## Global parameters

Chart parameters are only accessible in that chart (and have limited availability in the dependent charts). One easy way to access the parameters of one chart in other linked charts is to use global parameters.

**A global parameter has global scope** — a parameter declared in a parent, child, or other linked chart becomes available *in all linked charts* at the same path:

```yaml
# Chart.yaml:
name: myparent
dependencies:
- name: mychild1
- name: mychild2
```

```yaml
# charts/mychild1/values.yaml:
global:
  myparam: myvalue
```

{% raw %}

```
# templates/example.yaml:
myparent: {{ $.Values.global.myparam }}
```

{% endraw %}

{% raw %}

```
# charts/mychild1/templates/example.yaml:
mychild1: {{ $.Values.global.myparam }}
```

{% endraw %}

{% raw %}

```
# charts/mychild2/templates/example.yaml:
mychild2: {{ $.Values.global.myparam }}
```

{% endraw %}

Output:

```yaml
myparent: myvalue
---
mychild1: myvalue
---
mychild2: myvalue
```

## Secret parameters (werf only)

You can use secret parameter files to store secret parameters. These files are encrypted and stored in the Git repository.

By default, werf tries to locate the `.helm/secret-values.yaml` file containing the encrypted parameters. When werf finds the file, it decrypts it and merges the decrypted parameters with the others:

```yaml
# .helm/values.yaml:
plainParam: plainValue
```

```yaml
# .helm/secret-values.yaml:
secretParam: 1000625c4f1d874f0ab853bf1db4e438ad6f054526e5dcf4fc8c10e551174904e6d0
```

{% raw %}

```
{{ $.Values.plainParam }}
{{ $.Values.secretParam }}
```

{% endraw %}

Output:

```
plainValue
secretValue
```

### Working with secret parameter files

The order in which the secret parameters files are processed:

1. Use an existing secret key or create a new one by running `werf helm secret generate-secret-key`.

2. Save the secret key to the `WERF_SECRET_KEY` environment variable, or to `<Git repository root>/.werf_secret_key` or `<home directory>/.werf/global_secret_key`.

3. Use the `werf helm secret values edit .helm/secret-values.yaml` command to open the secret parameters file and add the decrypted parameters to it or modify them.

4. Save the file — it will be encrypted and saved in an encrypted form.

5. Commit the added/modified file `.helm/secret-values.yaml` to Git.

6. On subsequent werf invocations, the secret key must be present in the aforementioned environment variable or files, or werf will not be able to decrypt the secret parameter file.

> Note that someone with access to the secret key can decrypt the contents of the secret settings file, so **keep the secret key in a safe place!**

When using the `<Git repository root>/.werf_secret_key` file, be sure to add it to `.gitignore` to avoid unintentionally saving it to the Git repository.

Many werf commands can also be run without a secret key (the `--ignore-secret-key` flag enables this), in which case the parameters will be available for use in an encrypted rather than decrypted form.

### Additional secret parameter files

You can create and use extra secret files in addition to the `.helm/secret-values.yaml` file:

```yaml
# .helm/secret-values-production.yaml:
secret: 1000625c4f1d874f0ab853bf1db4e438ad6f054526e5dcf4fc8c10e551174904e6d0
```

```shell
werf --secret-values .helm/secret-values-production.yaml
```

## Information about the built images (werf only)

werf stores information about the built images in the `$.Values.werf` parameters of the main chart:

```yaml
werf:
  image:
    # The full path to the built Docker image for the "backend" werf image:
    backend: example.org/apps/myapp:a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
  # The address of the container registry for the built images:
  repo: example.org/apps/myapp
  tag:
    # The tag of the built Docker image for the "backend" werf image:
    backend: a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
```

Example of use:

{% raw %}

```
image: {{ $.Values.werf.image.backend }}
```

{% endraw %}

Output:

```yaml
image: example.org/apps/myapp:a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
```

The `export-values` directive allows you to use `$.Values.werf` in the dependent charts (werf only):

```yaml
# .helm/Chart.yaml:
dependencies:
- name: backend
  export-values:
  - parent: werf
    child: werf
```

{% raw %}

```
# .helm/charts/backend/templates/example.yaml:
image: {{ $.Values.werf.image.backend }}
```

{% endraw %}

Output:

```yaml
image: example.org/apps/myapp:a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
```

## Release information

werf stores release information in the properties of the `$.Release` object:

```yaml
# Is the release being installed for the first time?
IsInstall: true
# Is the existing release being updated?
IsUpgrade: false
# Release name:
Name: myapp-production
# Kubernetes Namespace name:
Namespace: myapp-production
# Release revision:
Revision: 1
```

... and in the `$.Values.werf` parameters of the main chart (werf only):

```yaml
werf:
  # Name of the werf project:
  name: myapp
  # Environment:
  env: production
```

Example of use:

{% raw %}

```
{{ $.Release.Namespace }}
{{ $.Values.werf.env }}
```

{% endraw %}

Output:

```
myapp-production
production
```

The `export-values` directive allows you to use `$.Values.werf` in the dependent charts (werf only):

```yaml
# .helm/Chart.yaml:
dependencies:
- name: backend
  export-values:
  - parent: werf
    child: werf
```

{% raw %}

```
# .helm/charts/backend/templates/example.yaml:
{{ $.Values.werf.env }}
```

{% endraw %}

Output:

```yaml
production
```

## Chart information

werf stores information about the current chart in the `$.Chart` object:

```yaml
# Is the chart the main one?
IsRoot: true

# Contents of the Chart.yaml:
Name: mychart
Version: 1.0.0
Type: library
KubeVersion: "~1.20.3"
AppVersion: "1.0"
Deprecated: false
Icon: https://example.org/mychart-icon.svg
Description: This is My Chart
Home: https://example.org
Sources:
  - https://github.com/my/chart
Keywords:
  - apps
Annotations:
  anyAdditionalInfo: here
Dependencies:
- Name: redis
  Condition: redis.enabled
```

Example of use:

{% raw %}

```
{{ $.Chart.Name }}
```

{% endraw %}

Output:

```
mychart
```

## Template information

werf stores information about the current template in the properties of the `$.Template` object:

```yaml
# The relative path to the chart's templates directory:
BasePath: mychart/templates
# The relative path to the current template file:
Name: mychart/templates/example.yaml
```

Example of use:

{% raw %}

```
{{ $.Template.Name }}
```

{% endraw %}

Output:

```
mychart/templates/example.yaml
```

## Git commit information (werf only)

werf stores information about the Git commit that triggered its run in the `$.Values.werf.commit` parameters of the main chart:

```yaml
werf:
  commit:
    date:
      # The date of the Git commit that triggered werf (the human-readable form):
      human: 2022-01-21 18:51:39 +0300 +0300
      # The date of the Git commit that triggered werf (Unix time):
      unix: 1642780299
    # The hash of the Git commit that triggered werf:
    hash: 1b28e6843a963c5bdb3579f6fc93317cc028051c
```

Example of use:

{% raw %}

```
{{ $.Values.werf.commit.hash }}
```

{% endraw %}

Output:

```
1b28e6843a963c5bdb3579f6fc93317cc028051c
```

The `export-values` directive allows you to use `$.Values.werf.commit` in the dependent charts (werf only):

```yaml
# .helm/Chart.yaml:
dependencies:
- name: backend
  export-values:
  - parent: werf
    child: werf
```

{% raw %}

```
# .helm/charts/backend/templates/example.yaml:
{{ $.Values.werf.commit.hash }}
```

{% endraw %}

Output:

```yaml
1b28e6843a963c5bdb3579f6fc93317cc028051c
```

## Information about the Kubernetes cluster

werf provides information about the capabilities of the target Kubernetes cluster via the properties of the `$.Capabilities` object:

```yaml
KubeVersion:
  # The full version of the Kubernetes cluster:
  Version: v1.20.0
  # The major version of the Kubernetes cluster:
  Major: "1"
  # The minor version of the Kubernetes cluster:
  Minor: "20"
# A list of APIs supported by the Kubernetes cluster:
APIVersions:
- apps/v1
- batch/v1
- # ...
```

... as well as methods of the `$.Capabilities` object:

* `APIVersions.Has <arg>` — tells whether the Kubernetes cluster supports the API (e.g., `apps/v1`) or resource (e.g., `apps/v1/Deployment`) specified by the argument.

Example of use:

{% raw %}

```
{{ $.Capabilities.KubeVersion.Version }}
{{ $.Capabilities.APIVersions.Has "apps/v1" }}
```

{% endraw %}

Output:

```
v1.20.0
true
```
