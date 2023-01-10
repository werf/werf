---
title: Templates parametrization
permalink: usage/deploy/values.html
---

The Values is an arbitrary YAML file filled with parameters that you can use in [templates]({{ "/usage/deploy/templates.html" | true_relative_url }}). All values passed to the chart can be divided by following categories:

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

Values placed under the `global` key will be available both in the current chart and in all [subcharts]({{ "/usage/deploy/charts.html#dependencies-on-other-charts" | true_relative_url }}).

Values placed under the arbitrary `SOMEKEY` key will be available in the current chart and in the `SOMEKEY` [subchart]({{ "usage/deploy/charts.html#dependencies-on-other-charts" | true_relative_url }}).

The `.helm/values.yaml` file is the default place to store values. You can also pass additional user-defined regular values via:

* Separate value files by specifying `--values=PATH_TO_FILE` (you can use it repeatedly to pass multiple files) as a werf option.
* Options `--set key1.key2.key3.array[0]=one`, `--set key1.key2.key3.array[1]=two` (can be used multiple times, [see also]({{ "/reference/cli/werf_helm_install.html" | true_relative_url }}) `--set-string key=forced_string_value`, and `--set-file key1.data.notes=notes.txt`).

**NOTE** All values files, including default `.helm/values.yaml` or any custom values file passed with `--values` option — all of these should be stored in the git repo of the project due to giterminism.

### Set parameters

There is an ability to redefine or pass new values using cli params:

 - `--set KEY=VALUE`;
 - `--set-string KEY=VALUE`;
 - `--set-file=PATH`;
 - `--set-docker-config-json-value=true|false`.

**NOTE** All files, specified with `--set-file` option should be stored in the git repo of the project due to giterminism.

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

Each value (like `100024fe29e45bf00665d3399f7545f4af63f09cc39790c239e16b1d597842161123`) in the secret value map is encoded by werf. The structure of the secret value map is the same as that of a regular value map (for example, in `values.yaml`). See more info [about secret value generation and working with secrets]({{ "/usage/deploy/values.html#secret-values-and-files" | true_relative_url }}).

The `.helm/secret-values.yaml` file is the default place for storing secret values. You can also pass additional user-defined secret values via separate secret value files by specifying `--secret-values=PATH_TO_FILE` (can be used repeatedly to pass multiple files). All such files should be stored in the git repo of the project due to giterminism.

**NOTE** All values files, including default `.helm/values.yaml` or any custom values file passed with `--values` option — all of these should be stored in the git repo of the project due to giterminism.

## Service values

Service values are set by the werf to pass additional data when rendering chart templates.

Here is an example of service data:

```yaml
werf:
  name: myapp
  namespace: myapp-production
  env: production
  repo: registry.domain.com/apps/myapp
  image:
    assets: registry.domain.com/apps/myapp:a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
    rails: registry.domain.com/apps/myapp:e760e9311f938e3d92681e93da3a81e176aa7f7e684ee06d092ec199-1598269478292
  tag:
    assets: a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
    rails: e760e9311f938e3d92681e93da3a81e176aa7f7e684ee06d092ec199-1598269478292
  commit:
    date:
      human: 2022-01-21 18:51:39 +0300 +0300
      unix: 1642780299
    hash: 1b28e6843a963c5bdb3579f6fc93317cc028051c

global:
  werf:
    name: myapp
    version: v1.2.7
```

There are following service values:
 - The name of the werf project: `.Values.werf.name`.
 - Version of the werf cli util: `.Values.werf.version`.
 - Resources will be deployed in `.Values.werf.namespace` namespace.
 - Name of a CI/CD environment used during the current deploy process: `.Values.werf.env`.
 - Container registry repo used during the current deploy process: `.Values.werf.repo`.
 - Full images names used during the current deploy process: `.Values.werf.image.NAME`.
 - Only tags of built images. Usually used in combination with `.Values.werf.repo` to pass image repos and image tags separately.
 - Info about commit from which werf was executed: `.Values.werf.commit.hash`, `.Values.werf.commit.date.human`, `.Values.werf.commit.date.unix`.

### Service values in the subcharts

If you are using [subcharts]({{ "/usage/deploy/charts.html#dependencies-on-other-charts" | true_relative_url }}) and you want to use regular werf service values (those that defined outside of `global` scope) in subchart, you need to explicitly export these parent-scoped service values from the parent to the subchart:

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

More info about `export-values` can be found [here]({{ "usage/deploy/charts.html#export-values-import-values" | true_relative_url }}).

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

## Secret values and files

werf secrets engine is recommended for storing database passwords, files with encryption certificates, etc.

The idea is that sensitive data must be stored in a repository served by an application and remain independent of any specific server.

werf supports passing secrets as:
 - separate [secret values]({{ "/usage/deploy/values.html#user-defined-secret-values" | true_relative_url }}) yaml file (`.helm/secret-values.yaml` by default, or any file passed by the `--secret-values` option);
 - secret files — raw encoded files, which can be used in the templates.

## Encryption key

A key is required for encryption and decryption of data. There are three locations from which werf can read the key:
* from the `WERF_SECRET_KEY` environment variable
* from a special `.werf_secret_key` file in the project root
* from `~/.werf/global_secret_key` (globally)

> Encryption key must be **hex dump** of either 16, 24, or 32 bytes long to select AES-128, AES-192, or AES-256. [werf helm secret generate-secret-key command]({{ "reference/cli/werf_helm_secret_generate_secret_key.html" | true_relative_url }}) returns AES-128 encryption key

You can promptly generate a key using the [werf helm secret generate-secret-key command]({{ "reference/cli/werf_helm_secret_generate_secret_key.html" | true_relative_url }}).

> **ATTENTION! Do not save the file into the git repository. If you do it, the entire sense of encryption is lost, and anyone who has source files at hand can retrieve all the passwords. `.werf_secret_key` must be kept in `.gitignore`!**

## Secret key rotation

To regenerate secret files and values with new secret key use [werf helm secret rotate-secret-key command]({{ "reference/cli/werf_helm_secret_rotate_secret_key.html" | true_relative_url }}).

## Secret values

The secret values file is designed for storing secret values. **By default** werf uses `.helm/secret-values.yaml` file, but user can specify arbitrary number of such files.

Secret values file may look like:
```yaml
mysql:
  host: 10005968c24e593b9821eadd5ea1801eb6c9535bd2ba0f9bcfbcd647fddede9da0bf6e13de83eb80ebe3cad4
  user: 100016edd63bb1523366dc5fd971a23edae3e59885153ecb5ed89c3d31150349a4ff786760c886e5c0293990
  password: 10000ef541683fab215132687a63074796b3892d68000a33a4a3ddc673c3f4de81990ca654fca0130f17
  db: 1000db50be293432129acb741de54209a33bf479ae2e0f53462b5053c30da7584e31a589f5206cfa4a8e249d20
```

To manage secret values files use the following commands:
- [`werf helm secret values edit` command]({{ "reference/cli/werf_helm_secret_values_edit.html" | true_relative_url }})
- [`werf helm secret values encrypt` command]({{ "reference/cli/werf_helm_secret_values_encrypt.html" | true_relative_url }})
- [`werf helm secret values decrypt` command]({{ "reference/cli/werf_helm_secret_values_decrypt.html" | true_relative_url }})

### Using in a chart template

The secret values files are decoded in the course of deployment and used in helm as [additional values](https://helm.sh/docs/chart_template_guide/values_files/). Thus, given the following secret values yaml:

```yaml
# .helm/secret-values.yaml
mysql:
  user: 10003c7f513b1ba1a0eb3d2cfb8294c93fddda8701850aa8adc1d9032229ddb4fd3b
  password: 1000cd6674285b65f55b739ee2e5130cfc6d01d87772c9e62c1c917d9b10194f14ef
```

— usage of these values is the same as regular values:

{% raw %}
```yaml
...
env:
- name: MYSQL_USER
  value: {{ .Values.mysql.user }}
- name: MYSQL_PASSWORD
  value: {{ .Values.mysql.password }}
```
{% endraw %}

## Secret files

Secret files are excellent for storing sensitive data such as certificates and private keys in the project repository. For these files, the `.helm/secret` directory is allocated where encrypted files must be stored.

To use secret data in helm templates, you must save it to an appropriate file in the `.helm/secret` directory.

To manage secret files use the following commands:
 - [`werf helm secret file edit` command]({{ "reference/cli/werf_helm_secret_file_edit.html" | true_relative_url }})
 - [`werf helm secret file encrypt` command]({{ "reference/cli/werf_helm_secret_file_encrypt.html" | true_relative_url }})
 - [`werf helm secret file decrypt` command]({{ "reference/cli/werf_helm_secret_file_decrypt.html" | true_relative_url }})

> **NOTE** werf will decrypt all files in the `.helm/secret` directory prior rendering helm chart templates. Make sure that `.helm/secret` contains valid encrypted files.

### Using in a chart template

<!-- Move to reference -->

The `werf_secret_file` runtime function allows using decrypted file content in a template. The required function argument is a secret file path relative to `.helm/secret` directory.

Using the decrypted secret `.helm/secret/backend-saml/tls.key` in a template may appear as follows:

{% raw %}
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: myproject-backend-saml
type: kubernetes.io/tls
data:
  tls.crt: {{ werf_secret_file "backend-saml/stage/tls.crt" | b64enc }}
  tls.key: {{ werf_secret_file "backend-saml/stage/tls.key" | b64enc }}
```
{% endraw %}

Note that `backend-saml/stage/` is an arbitrary file structure. User can place all files into the single directory `.helm/secret` or create subdirectories at his own discretion.
