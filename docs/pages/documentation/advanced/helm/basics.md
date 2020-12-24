---
title: Basics
sidebar: documentation
permalink: documentation/advanced/helm/basics.html
---

You can easily start using werf for deploying your projects via the existing [Helm](https://helm.sh) charts since they are fully compatible with werf. The configuration has a format similar to that of [Helm charts](#chart)

werf includes all the existing Helm functionality (the latter is integrated into werf) as well as some custom solutions:

- werf has several configurable modes for tracking the resources being deployed, including processing logs and events;
- you can integrate images built by werf into Helm charts’ [templates](#templates);
- you can assign arbitrary annotations and labels to all resources being deployed to Kubernetes;
- also, werf has some unique features which we will discuss below.

werf uses the following two commands to deal with an application in the Kubernetes cluster:
- [converge]({{ "documentation/reference/cli/werf_converge.html" | true_relative_url: page.url }}) — to install or update an application;
- [dismiss]({{ "documentation/reference/cli/werf_dismiss.html" | true_relative_url: page.url }}) — to delete an application from the cluster.

## Chart

The chart is a set of configuration files that describe an application. Chart files reside in the `.helm` directory under the root directory of the project:

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

### Templates

Templates are placed in the `.helm/templates` directory.

Directory contains `*.yaml` YAML files. Each YAML file describes one or several Kubernetes resources separated by three hyphens `---`, for example:

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
 * parameterize templates with [values](#values) for different environments;
 * extract common text parts, save them as Go templates, and reuse in several places;
 * etc.

You can use [Sprig functions](https://masterminds.github.io/sprig/) and [advanced functions](https://helm.sh/docs/howto/charts_tips_and_tricks/) such as `include` and `required` in addition to basic functions of Go templates.

Also, the user can place `*.tpl` files into the directory. They will not be rendered into the Kubernetes object. These files can be used to store Go templates. You can insert templates contained the `*.tpl` files into the `*.yaml` files.

#### Integration with built images

The user must specify the full docker image name, including the docker repo and the docker tag, in order to use the docker image in the chart resource specifications. But how do you designate an image contained in the `werf.yaml` file given that the full docker image name for such an image depends on the specified image repository?

werf provides runtime function to address this issue: [`werf_image`](#werf_image). The user just needs to specify image which is described in the `werf.yaml` by a short name, and full image name will be generated as output.

##### werf_image

This template function generates full image name which can be used as a value for `image` key for the container in the pod. Function takes image name argument.

Here is an example of using the function with the **named** image:
* `tuple <image-name> . | werf_image`

Here is an example of using the function with the **unnamed** image:
* `tuple . | werf_image`
* `werf_image .` (additional simplified entry format)

##### Examples

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
        image: {{ tuple "backend" . | werf_image }}
```
{% endraw %}

Refer to a single unnamed image described `werf.yaml`:

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
        image: {{ werf_image . }}
```
{% endraw %}

#### Secret files

Secret files are excellent for storing sensitive data such as certificates and private keys in the project repository.

Secret files are placed in the `.helm/secret` directory. The user can create an arbitrary files structure in this directory. [This article about secrets]({{ "documentation/advanced/helm/working_with_secrets.html#secret-file-encryption" | true_relative_url: page.url }}) describes how to encrypt them.

##### werf_secret_file

`werf_secret_file` is a runtime template function that allows the user to fetch the contents of a secret file and insert them in a chart template. This template function reads file context that is usually placed in the resource yaml manifest of those resources as Secrets. The template function requires a relative path to the file inside the `.helm/secret` directory as an argument.

Here is an example of inserting the decrypted contents of `.helm/secret/backend-saml/stage/tls.key` and `.helm/secret/backend-saml/stage/tls.crt` files into the template:

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

Note that `backend-saml/stage/` is an arbitrary file structure. The user can place all files into the single directory `.helm/secret` or create subdirectories at his own discretion.

#### Builtin templates and params

{% raw %}
 * `{{ .Chart.Name }}` — contains the [project name] specified in the `werf.yaml` config.
 * `{{ .Release.Name }}` — contains the [release name](#release).
 * `{{ .Files.Get }}` — a function to read file contents into templates; requires the path to file as an argument. The path should be specified relative to the `.helm` directory (files outside the `.helm` folder are ignored).
{% endraw %}

### Values

The Values is an arbitrary YAML file filled with parameters that you can use in [templates](#templates).

There are different types of values in werf:

 * User-defined regular values.
 * User-defined secret values.
 * Service values.

#### User-defined regular values

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

Values placed under the `global` key will be available both in the current chart and in all [subcharts]({{ "documentation/advanced/helm/working_with_chart_dependencies.html" | true_relative_url: page.url }}).

Values placed under the arbitrary `SOMEKEY` key will be available in the current chart and in the `SOMEKEY` [subchart]({{ "documentation/advanced/helm/working_with_chart_dependencies.html" | true_relative_url: page.url }}).

The `.helm/values.yaml` file is the default place to store values. You can also pass additional user-defined regular values via:

 * Separate value files by specifying `--values=PATH_TO_FILE` (you can use it repeatedly to pass multiple files) as a werf option.
 * Options `--set key1.key2.key3.array[0]=one`, `--set key1.key2.key3.array[1]=two` (can be used multiple times, see also `--set-string key=forced_string_value`).

#### User-defined secret values

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

Each value (like `100024fe29e45bf00665d3399f7545f4af63f09cc39790c239e16b1d597842161123`) in the secret value map is encoded by werf. The structure of the secret value map is the same as that of a regular value map (for example, in `values.yaml`). See more info [about secret value generation and working with secrets]({{ "documentation/advanced/helm/working_with_secrets.html#secret-values-encryption" | true_relative_url: page.url }}).

The `.helm/secret-values.yaml` file is the default place for storing secret values. You can also pass additional user-defined secret values via separate secret value files by specifying `--secret-values=PATH_TO_FILE` (can be used repeatedly to pass multiple files).

#### Service values

Service values are generated by werf automatically to pass additional data when rendering chart templates.

Here is a comprehensive example of service data:

```yaml
global:
  env: stage
  namespace: myapp-stage
  werf:
    image:
      assets:
        docker_image: registry.domain.com/apps/myapp/assets:a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
        docker_tag: a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
      rails:
        docker_image: registry.domain.com/apps/myapp/rails:e760e9311f938e3d92681e93da3a81e176aa7f7e684ee06d092ec199-1598269478292
        docker_image_id: e760e9311f938e3d92681e93da3a81e176aa7f7e684ee06d092ec199-1598269478292
    name: myapp
    repo: registry.domain.com/apps/myapp
```

There are the following service values:
 * Name of a CI/CD environment used during the deployment:: `.Values.global.env`.
 * Kubernetes namespace used during the deployment: `.Values.global.namespace`.
 * Full docker image name and tag for each image contained in the `werf.yaml` config: `.Values.global.werf.image.IMAGE_NAME.docker_image` and `.Values.global.werf.image.IMAGE_NAME.docker_tag`.
 * `.Values.global.werf.is_nameless_image` indicates whether there is a nameless image defined in the `werf.yaml` config.
 * Project name as specified in `werf.yaml`: `.Values.global.werf.name`.
 * Repo used during the deployment: `.Values.global.werf.repo`.

#### Merging the resulting values

During the deployment process, werf merges all user-defined regular, secret, and service values into the single value map, which is then passed to the template rendering engine to be used in the templates (see [how to use values in the templates](#using-values-in-the-templates)). Values are merged in the following order of priority (the more recent value overwrites the previous one):

 1. User-defined regular values from `.helm/values.yaml`.
 2. User-defined regular values from all cli options `--values=PATH_TO_FILE` in the order specified.
 3. User-defined secret values from `.helm/secret-values.yaml`.
 4. User-defined secret values from all cli options `--secret-values=PATH_TO_FILE` in the order specified.
 5. Service values.

### Using values in the templates

werf uses the following syntax for accessing values contained in the chart templates:

{% raw %}
```yaml
{{ .Values.key.key.arraykey[INDEX].key }}
```
{% endraw %}

The `.Values` object contains the [merged map of resulting values](#merging-the-resulting-values) map.

## Release

While the chart is a collection of configuration files of your application, the release is a runtime object that represents a running instance of your application deployed via werf.

Each release has a single name and multiple versions. A new release version is created with each invocation of werf converge.

### Storing releases

Each release version is stored in the Kubernetes cluster itself. werf supports storing release data in ConfigMap or Secret objects in the arbitrary namespace.

By default, werf stores releases in the ConfigMaps in the `kube-system` namespace, and this is fully compatible with the default [Helm 2](https://helm.sh) configuration. You can set the release storage by werf converge cli options: `--helm-release-storage-namespace=NS` and `--helm-release-storage-type=configmap|secret`.

The [werf helm list]({{ "documentation/reference/cli/werf_helm_list.html" | true_relative_url: page.url }}) command lists releases created by werf. Also, the user can fetch the history of a specific release with the [werf helm history]({{ "documentation/reference/cli/werf_helm_history.html" | true_relative_url: page.url }}) command.

#### Helm compatibility notice

werf is fully compatible with the existing Helm 2 installations since the release data is stored similarly to Helm. If the existing Helm installation uses non-default release storage, then you might need to configure the release storage correspondingly using `--helm-release-storage-namespace` and `--helm-release-storage-type` options.

You can inspect and browse releases created by werf using commands such as `helm list` and `helm get`. werf can also upgrade existing releases created by Helm.

Furthermore, you can use werf and Helm 2 simultaneously in the same cluster and at the same time.

### Environment

By default, werf assumes that each release should be tainted with some environment, such as `staging`, `test` or `production`.

Using this environment, werf determines:

 1. Release name.
 2. Kubernetes namespace.

The environment is a required parameter for deploying and should be specified either with an `--env` option or determined automatically using the data for the CI/CD system used. See the [CI/CD configuration integration]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#cicd-configuration-integration" | true_relative_url: page.url }}) for more info.

### Release name

The release name is constructed using the template `[[ project ]]-[[ env ]]` by default. Here, `[[ project ]]` refers to the [project name]({{ "documentation/reference/werf_yaml.html#project-name" | true_relative_url: page.url }}) and `[[ env ]]` refers to the specified or detected environment.

For example, for the project named `symfony-demo`, the following Helm Release names can be constructed depending on the environment:
* `symfony-demo-stage` for the `stage` environment;
* `symfony-demo-test` for the `test` environment;
* `symfony-demo-prod` for the `prod` environment.

You can redefine the release using the `--release NAME` deploy option. In that case werf would use the specified name as is.

You can also define the custom release name in the werf.yaml configuration [by setting `deploy.helmRelease`]({{ "documentation/advanced/helm/basics.html#release-name" | true_relative_url: page.url }}).

#### Slugging the release name

The name of the Helm Release constructed using the template will be slugified according to the [*release slug procedure*]({{ "documentation/internals/names_slug_algorithm.html#basic-algorithm" | true_relative_url: page.url }}) to fit the requirements for the release name. This procedure generates a unique and valid Helm Release name.

This is the default behavior. You can disable it by [setting `deploy.helmReleaseSlug=false`]({{ "documentation/advanced/helm/basics.html#release-name" | true_relative_url: page.url }}) in the `werf.yaml` configuration.

### Kubernetes namespace

The Kubernetes namespace is constructed using the template `[[ project ]]-[[ env ]]` by default. Here, `[[ project ]]` refers to the [project name]({{ "documentation/reference/werf_yaml.html#project-name" | true_relative_url: page.url }}) and `[[ env ]]` refers to the environment.

For example, for the project named `symfony-demo`, there can be the following Kubernetes namespaces depending on the environment specified:

* `symfony-demo-stage` for the `stage` environment;
* `symfony-demo-test` for the `test` environment;
* `symfony-demo-prod` for the `prod` environment.

You can redefine the Kubernetes Namespace using the `--namespace NAMESPACE` deploy option. In that case, werf would use the specified name as is.

You can also define the custom Kubernetes Namespace in the werf.yaml configuration [by setting `deploy.namespace`]({{ "documentation/advanced/helm/basics.html#kubernetes-namespace" | true_relative_url: page.url }}) parameter.

#### Slugging Kubernetes namespace

The Kubernetes namespace that is constructed using the template will be slugified to fit the [DNS Label](https://www.ietf.org/rfc/rfc1035.txt) requirements according to the [*namespace slugging procedure*]({{ "documentation/internals/names_slug_algorithm.html#basic-algorithm" | true_relative_url: page.url }}) that generates a unique and valid Kubernetes Namespace.

This is default behavior. It can be disabled by [setting `deploy.namespaceSlug=false`]({{ "documentation/advanced/helm/basics.html#kubernetes-namespace" | true_relative_url: page.url }}) in the werf.yaml configuration.

## Deploy process

When running the `werf converge` command, werf starts the deployment process that includes the following steps:

 1. Rendering chart templates into a single list of Kubernetes resource manifests and validating them.
 2. Running `pre-install` or `pre-upgrade` [hooks](#helm-hooks) and tracking each hook until successful or failed termination; printing logs and other information in the process.
 3. Applying changes to Kubernetes resources: creating new, deleting old, updating the existing.
 4. Creating a new release version and saving the current state of resource manifests into this release’s data.
 5. Tracking all release resources until the readiness state is reached; printing logs and other information in the process.
 6. Running `post-install` or `post-upgrade` [hooks](#helm-hooks) and tracking each hook until successful or failed termination; printing logs and other information in the process.

**NOTE:** werf would delete all newly created resources immediately during the ongoing deploy process if this process fails at any of the steps described above!

When executing helm hooks at the step 2 and 6, werf would track these hooks resources until successful termination. Tracking [can be configured](#configure-resource-tracking) for each hook resource.

On step 5, werf would track all release resources until each resource reaches the "ready" state. All resources are tracked simultaneously. During tracking, werf aggregates information obtained from all release resources into the single text output in real-time, and periodically prints the so-called status progress table. Tracking [can be configured](#configure-resource-tracking) for each resource.

werf displays logs of resource Pods until those pods reach the "ready" state. In the case of Job pods, logs are shown until Pods are terminated.

werf uses the [kubedog library](https://github.com/werf/kubedog) to track resources. Currently, tracking is implemented for Deployments, StatefulSets, DaemonSets, and Jobs. We plan to implement support for tracking Service, Ingress, PVC, and other resources [in the near future](https://github.com/werf/werf/issues/1637).

### If the deploy failed

In the case of failure during the release process, werf would create a new release having the FAILED state. This state can then be inspected by the user to find the problem and solve it on the next deploy invocation.

### Helm hooks

The helm hook is an arbitrary Kubernetes resource marked with the `helm.sh/hook` annotation. For example:

```yaml
kind: Job
metadata:
  name: somejob
  annotations:
    "helm.sh/hook": pre-upgrade,pre-install
    "helm.sh/hook-weight": "1"
```
A lot of various helm hooks come into play during the deploy process. We have already discussed `pre|post-install|upgrade` hooks during the [deploy process](#deploy-process). These hooks are often used to perform tasks such as migrations (in the case of `pre-upgrade` hooks) or some post-deploy actions. The full list of available hooks can be found in the [helm docs](https://helm.sh/docs/topics/charts_hooks/).

Hooks are sorted in the ascending order specified by the `helm.sh/hook-weight` annotation (hooks with the same weight are sorted by the name). After that, hooks are created and executed sequentially. werf recreates the Kubernetes resource for each hook if that resource already exists in the cluster. Hooks of Kubernetes resources are not deleted after executing.

### Configure resource tracking

Tracking can be configured for each resource using [resource annotations]({{ "documentation/reference/deploy_annotations.html" | true_relative_url: page.url }}), which should be set in the chart templates.

### Annotating and labeling of chart resources

#### Auto annotations

werf automatically sets the following built-in annotations to all deployed chart resources:

 * `"werf.io/version": FULL_WERF_VERSION` — version of werf used when running the `werf converge` command;
 * `"project.werf.io/name": PROJECT_NAME` — project name specified in the `werf.yaml`;
 * `"project.werf.io/env": ENV` — environment name specified via the `--env` param or `WERF_ENV` variable; optional, will not be set if env is not used.

werf also sets auto annotations containing information from the CI/CD system used (for example, GitLab CI)  when running the `werf ci-env` command prior to the `werf converge` command. For example, [`project.werf.io/git`]({{ "documentation/internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_project_git" | true_relative_url: page.url }}), [`ci.werf.io/commit`]({{ "documentation/internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_ci_commit" | true_relative_url: page.url }}), [`gitlab.ci.werf.io/pipeline-url`]({{ "documentation/internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_gitlab_ci_pipeline_url" | true_relative_url: page.url }}) and [`gitlab.ci.werf.io/job-url`]({{ "documentation/internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_gitlab_ci_job_url" | true_relative_url: page.url }}).

For more information about the CI/CD integration, please refer to the following pages:

 * [plugging into CI/CD overview]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html" | true_relative_url: page.url }});
 * [plugging into GitLab CI]({{ "documentation/internals/how_ci_cd_integration_works/gitlab_ci_cd.html" | true_relative_url: page.url }}).

#### Custom annotations and labels

The user can pass arbitrary additional annotations and labels using `--add-annotation annoName=annoValue` (can be used repeatedly) and `--add-label labelName=labelValue` (can be used repeatedly) CLI options when invoking werf converge.

For example, you can use the following werf converge invocation to set `commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57`, `gitlab-user-email=vasya@myproject.com`  annotations/labels to all Kubernetes resources in a chart:

```shell
werf converge \
  --add-annotation "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-label "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-annotation "gitlab-user-email=vasya@myproject.com" \
  --add-label "gitlab-user-email=vasya@myproject.com" \
  --env dev \
  --repo REPO
```

### Validating resource manifests

If the resource manifest in the chart contains logic or syntax errors, then werf would print the validation warning to the output during the deployment process. Also, all validation errors will be written to the `debug.werf.io/validation-messages`. These errors typically do not affect the exit status of the deploy process since Kubernetes apiserver can accept incorrect manifests containing certain typos or errors without any warnings.

Let us suppose there are the following typos in the chart template (`envs` in place of `env` and `redinessProbe` in lieu of `readinessProbe`):

```yaml
containers:
- name: main
  command: [ "/bin/bash", "-c", "while true; do date ; sleep 1 ; done" ]
  image: ubuntu:18.04
  redinessProbe:
    tcpSocket:
      port: 8080
    initialDelaySeconds: 5
    periodSeconds: 10
envs:
- name: MYVAR
  value: myvalue
```

Validation output will be like:

```
│   WARNING ### Following problems detected during deploy process ###
│   WARNING Validation of target data failed: deployment/mydeploy1: [ValidationError(Deployment.spec.template.spec.containers[0]): unknown field               ↵
│ "redinessProbe" in io.k8s.api.core.v1.Container, ValidationError(Deployment.spec.template.spec): unknown field "envs" in io.k8s.api.core.v1.PodSpec]
```

As a result, the resource will contain the `debug.werf.io/validation-messages` annotation with the following contents:

```
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    debug.werf.io/validation-messages: 'Validation of target data failed: deployment/mydeploy1:
      [ValidationError(Deployment.spec.template.spec.containers[0]): unknown field
      "redinessProbe" in io.k8s.api.core.v1.Container, ValidationError(Deployment.spec.template.spec):
      unknown field "envs" in io.k8s.api.core.v1.PodSpec]'
...
```

## Multiple Kubernetes clusters

There are cases when separate Kubernetes clusters are required for a different environments. You can [configure access to multiple clusters](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters) using kube contexts in a single kube config.

In that case, the `--kube-context=CONTEXT` deploy option should be set manually along with the environment.
