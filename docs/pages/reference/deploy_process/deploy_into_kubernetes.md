---
title: Deploy into Kubernetes
sidebar: documentation
permalink: reference/deploy_process/deploy_into_kubernetes.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

You can easily start using werf for deploying your projects via the existing [Helm](https://helm.sh) charts since they are fully compatible with werf. The configuration has a format similar to that of [Helm charts](#chart)

werf includes all the existing Helm functionality (the latter is integrated into werf) as well as some custom solutions:

- werf has several configurable modes for tracking the resources being deployed, including processing logs and events;
- you can integrate images built by werf into Helm charts’ [templates](#templates);
- you can assign arbitrary annotations and labels to all resources being deployed to Kubernetes;
- also, werf has some unique features which we will discuss below.

werf uses the following two commands to deal with an application in the Kubernetes cluster:
- [deploy]({{ site.baseurl }}/cli/main/deploy.html) — to install or update an application;
- [dismiss]({{ site.baseurl }}/cli/main/dismiss.html) — to delete an application from the cluster.

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

The user must specify the full docker image name, including the docker repo and the docker tag, in order to use the docker image in the chart resource specifications. But how do you designate an image contained in the `werf.yaml` file given that the full docker image name for such an image depends on the selected tagging strategy and the specified image repository?

The next set of questions is: “How do I use the [`imagePullPolicy` Kubernetes parameter](https://kubernetes.io/docs/concepts/containers/images/#updating-images) together with images from `werf.yaml`: should I set `imagePullPolicy` to `Always`?” and “How do I pull images only when it is really necessary?”

werf provides two runtime functions to address these issues: [`werf_container_image`](#werf_container_image) and [`werf_container_env`](#werf_container_env). The user just needs to specify images described in the `werf.yaml`, and everything will be set based on the options specified.

##### werf_container_image

This template function generates `image` and `imagePullPolicy` keys for the container in the pod.

The unique feature of the function is that `imagePullPolicy` is generated based on the selected tagging scheme.  If you are using a custom tag (`--tag-custom`) or tagging by a branch name (` --tag-git-branch`), that is, the tag name is constant while an image may change during the subsequent build, then the function returns `imagePullPolicy: Always`. Otherwise, the `imagePullPolicy` key is not returned.

The function may return multiple strings, which is why it must be used together with the `indent` construction.

The logic behind generating the `imagePullPolicy` key is the following:

* The `.Values.global.werf.is_branch=true` value means that an image is being deployed based on the latest logic for a branch. The same is true for the custom tag `.Values.global.werf.ci.is_custom=true`.
  * In this case, the image tagged with the appropriate docker tag must be updated through docker pull (even if it already exists) to get the `latest` version of the respective tag.
  * In this case, imagePullPolicy=Always is set.
* The `.Values.global.werf.is_branch=false` value means that a tag or a specific image commit is being deployed.
  * Thus, the image for an appropriate docker tag doesn't need to be updated through docker pull if it already exists.
  * In this case, `imagePullPolicy` is not set, which is consistent with the default value currently used in Kubernetes: `imagePullPolicy=IfNotPresent`.

> Images tagged by a custom tagging strategy (`--tag-custom`) are processed similarly to images tagged by a git branch tagging strategy (`--tag-git-branch`)

Here is an example of using the function with the **named** image:
* `tuple <image-name> . | werf_container_image | indent <N-spaces>`

Here is an example of using the function with the **unnamed** image:
* `tuple . | werf_container_image | indent <N-spaces>`
* `werf_container_image . | indent <N-spaces>` (additional simplified entry format)

##### werf_container_env

Streamlines the release process if the image remains unchanged. It generates a block with the `DOCKER_IMAGE_ID` environment variable for the pod container. An image id will be set only if the custom user-defined tag (`--tag-custom`) or the git branch tagging strategy (`--tag-git-branch`) is used – this way, the name of the tag does not change while an image may change during the subsequent build. The value of the `DOCKER_IMAGE_ID` variable contains the current ID of the Docker image, which affects the Kubernetes object update as well as pod redeployment.

The function may return multiple strings, which is why it must be used together with the `indent` construct.

An example of using the function in the case of a named **image**:
* `tuple <image-name> . | werf_container_env | indent <N-spaces>`

An example of using the function in the case of an **unnamed** image:
* `tuple . | werf_container_env | indent <N-spaces>`
* `werf_container_env . | indent <N-spaces>` (additional simplified entry format)

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
{{ tuple "backend" . | werf_container_image | indent 8 }}
        env:
{{ tuple "backend" . | werf_container_env | indent 8 }}
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
{{ werf_container_image . | indent 8 }}
        env:
{{ werf_container_env . | indent 8 }}
```
{% endraw %}

#### Secret files

Secret files are excellent for storing sensitive data such as certificates and private keys in the project repository.

Secret files are placed in the `.helm/secret` directory. The user can create an arbitrary files structure in this directory. [This article about secrets]({{ site.baseurl }}/reference/deploy_process/working_with_secrets.html#secret-file-encryption) describes how to encrypt them.

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

Values placed under the `global` key will be available both in the current chart and in all [subcharts]({{ site.baseurl }}/reference/deploy_process/working_with_chart_dependencies.html).

Values placed under the arbitrary `SOMEKEY` key will be available in the current chart and in the `SOMEKEY` [subchart]({{ site.baseurl }}/reference/deploy_process/working_with_chart_dependencies.html).

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

Each value (like `100024fe29e45bf00665d3399f7545f4af63f09cc39790c239e16b1d597842161123`) in the secret value map is encoded by werf. The structure of the secret value map is the same as that of a regular value map (for example, in `values.yaml`). See more info [about secret value generation and working with secrets]({{ site.baseurl }}/reference/deploy_process/working_with_secrets.html#secret-values-encryption).

The `.helm/secret-values.yaml` file is the default place for storing secret values. You can also pass additional user-defined secret values via separate secret value files by specifying `--secret-values=PATH_TO_FILE` (can be used repeatedly to pass multiple files).

#### Service values

Service values are generated by werf automatically to pass additional data when rendering chart templates.

Here is a comprehensive example of service data:

```yaml
global:
  env: stage
  namespace: myapp-stage
  werf:
    ci:
      branch: mybranch
      is_branch: true
      is_tag: false
      tag: '"-"'
    docker_tag: mybranch
    image:
      assets:
        docker_image: registry.domain.com/apps/myapp/assets:mybranch
        docker_image_id: sha256:ddaec322ee2c622aa0591177062a81009d9e52785be6915c5a37e822c2019755
        docker_image_digest: sha256:81009d9e52785be6915c5a37e822c2019755ddaec322ee2c622aa0591177062a
      rails:
        docker_image: registry.domain.com/apps/myapp/rails:mybranch
        docker_image_id: sha256:646c56c828beaf26e67e84a46bcdb6ab555c6bce8ebeb066b79a9075d0e87f50
        docker_image_digest: sha256:555c6bce8ebeb066b79a9075d0e87f50646c56c828beaf26e67e84a46bcdb6ab
    is_nameless_image: false
    name: myapp
    repo: registry.domain.com/apps/myapp
```

There are the following service values:
 * Name of a CI/CD environment used during the deployment:: `.Values.global.env`.
 * Kubernetes namespace used during the deployment: `.Values.global.namespace`.
 * The values of the tagging strategy used: `.Values.global.werf.ci.is_branch`, `.Values.global.werf.ci.branch`, `.Values.global.werf.ci.is_tag`, `.Values.global.werf.ci.tag`.
 * `.Values.global.ci.ref` is set to either a git branch name or a git tag name (optional).
 * Full docker image names and their IDs for each image contained in the `werf.yaml` config: `.Values.global.werf.image.IMAGE_NAME.docker_image`, `.Values.global.werf.image.IMAGE_NAME.docker_image_id` and `.Values.global.werf.image.IMAGE_NAME.docker_image_digest`.
 * `.Values.global.werf.is_nameless_image` indicates whether there is a nameless image defined in the `werf.yaml` config.
 * Project name as specified in `werf.yaml`: `.Values.global.werf.name`.
 * Docker tag for images from `werf.yaml` used during the deployment (accordingly to the selected tagging strategy): `.Values.global.werf.docker_tag`.
 * Images repo used during the deployment: `.Values.global.werf.repo`.

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

Each release has a single name and multiple versions. A new release version is created with each invocation of werf deploy.

### Storing releases

Each release version is stored in the Kubernetes cluster itself. werf supports storing release data in ConfigMap or Secret objects in the arbitrary namespace.

By default, werf stores releases in the ConfigMaps in the `kube-system` namespace, and this is fully compatible with the default [Helm 2](https://helm.sh) configuration. You can set the release storage by werf deploy cli options: `--helm-release-storage-namespace=NS` and `--helm-release-storage-type=configmap|secret`.

The [werf helm list]({{ site.baseurl }}/cli/management/helm/list.html) command lists releases created by werf. Also, the user can fetch the history of a specific release with the [werf helm history]({{ site.baseurl }}/cli/management/helm/history.html) command.

#### Helm compatibility notice

werf is fully compatible with the existing Helm 2 installations since the release data is stored similarly to Helm. If the existing Helm installation uses non-default release storage, then you might need to configure the release storage correspondingly using `--helm-release-storage-namespace` and `--helm-release-storage-type` options.

You can inspect and browse releases created by werf using commands such as `helm list` and `helm get`. werf can also upgrade existing releases created by Helm.

Furthermore, you can use werf and Helm 2 simultaneously in the same cluster and at the same time.

### Environment

By default, werf assumes that each release should be tainted with some environment, such as `staging`, `test` or `production`.

Using this environment, werf determines:

 1. Release name.
 2. Kubernetes namespace.

The environment is a required parameter for deploying and should be specified either with an `--env` option or determined automatically using the data for the CI/CD system used. See the [CI/CD configuration integration]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#cicd-configuration-integration) for more info.

### Release name

The release name is constructed using the template `[[ project ]]-[[ env ]]` by default. Here, `[[ project ]]` refers to the [project name]({{ site.baseurl }}/configuration/introduction.html#project-name) and `[[ env ]]` refers to the specified or detected environment.

For example, for the project named `symfony-demo`, the following Helm Release names can be constructed depending on the environment:
* `symfony-demo-stage` for the `stage` environment;
* `symfony-demo-test` for the `test` environment;
* `symfony-demo-prod` for the `prod` environment.

You can redefine the release using the `--release NAME` deploy option. In that case werf would use the specified name as is.

You can also define the custom release name in the werf.yaml configuration [by setting `deploy.helmRelease`]({{ site.baseurl }}/configuration/deploy_into_kubernetes.html#release-name).

#### Slugging the release name

The name of the Helm Release constructed using the template will be slugified according to the [*release slug procedure*]({{ site.baseurl }}/reference/toolbox/slug.html#basic-algorithm) to fit the requirements for the release name. This procedure generates a unique and valid Helm Release name.

This is the default behavior. You can disable it by [setting `deploy.helmReleaseSlug=false`]({{ site.baseurl }}/configuration/deploy_into_kubernetes.html#release-name) in the `werf.yaml` configuration.

### Kubernetes namespace

The Kubernetes namespace is constructed using the template `[[ project ]]-[[ env ]]` by default. Here, `[[ project ]]` refers to the [project name]({{ site.baseurl }}/configuration/introduction.html#meta-config-section) and `[[ env ]]` refers to the environment.

For example, for the project named `symfony-demo`, there can be the following Kubernetes namespaces depending on the environment specified:

* `symfony-demo-stage` for the `stage` environment;
* `symfony-demo-test` for the `test` environment;
* `symfony-demo-prod` for the `prod` environment.

You can redefine the Kubernetes Namespace using the `--namespace NAMESPACE` deploy option. In that case, werf would use the specified name as is.

You can also define the custom Kubernetes Namespace in the werf.yaml configuration [by setting `deploy.namespace`]({{ site.baseurl }}/configuration/deploy_into_kubernetes.html#kubernetes-namespace) parameter.

#### Slugging Kubernetes namespace

The Kubernetes namespace that is constructed using the template will be slugified to fit the [DNS Label](https://www.ietf.org/rfc/rfc1035.txt) requirements according to the [*namespace slugging procedure*]({{ site.baseurl }}/reference/toolbox/slug.html#basic-algorithm) that generates a unique and valid Kubernetes Namespace.

This is default behavior. It can be disabled by [setting `deploy.namespaceSlug=false`]({{ site.baseurl }}/configuration/deploy_into_kubernetes.html#kubernetes-namespace) in the werf.yaml configuration.

## Deploy process

When running the `werf deploy` command, werf starts the deployment process that includes the following steps:

 1. Rendering chart templates into a single list of Kubernetes resource manifests and validating them.
 2. Running `pre-install` or `pre-upgrade` [hooks](#helm-hooks) and tracking each hook until successful or failed termination; printing logs and other information in the process.
 3. Applying changes to Kubernetes resources: creating new, deleting old, updating the existing.
 4. Creating a new release version and saving the current state of resource manifests into this release’s data.
 5. Tracking all release resources until the readiness state is reached; printing logs and other information in the process.
 6. Running `post-install` or `post-upgrade` [hooks](#helm-hooks) and tracking each hook until successful or failed termination; printing logs and other information in the process.

**NOTE:** werf would delete all newly created resources immediately during the ongoing deploy process if this process fails at any of the steps described above!

When executing helm hooks at the step 2 and 6, werf would track these hooks resources until successful termination. Tracking [can be configured](#configuring-resource-tracking) for each hook resource.

On step 5, werf would track all release resources until each resource reaches the "ready" state. All resources are tracked simultaneously. During tracking, werf aggregates information obtained from all release resources into the single text output in real-time, and periodically prints the so-called status progress table. Tracking [can be configured](#configuring-resource-tracking) for each resource.

werf displays logs of resource Pods until those pods reach the "ready" state. In the case of Job pods, logs are shown until Pods are terminated.

werf uses the [kubedog library](https://github.com/werf/kubedog) to track resources. Currently, tracking is implemented for Deployments, StatefulSets, DaemonSets, and Jobs. We plan to implement support for tracking Service, Ingress, PVC, and other resources [in the near future](https://github.com/werf/werf/issues/1637).

### Method of applying changes

werf tries to use 3-way-merge patches to update resources in the Kubernetes cluster since it is the best possible option. However, there are different resource update methods available.

See articles for more info:
 - [resource update methods and adoption]({{ site.baseurl }}/reference/deploy_process/resources_update_methods_and_adoption.html);
 - [differences with the helm resource update method]({{ site.baseurl }}/reference/deploy_process/differences_with_helm.html#three-way-merge-patches-and-resources-adoption);
 - ["3-way merge in werf: deploying to Kubernetes via Helm “on steroids” medium article](https://medium.com/flant-com/3-way-merge-patches-helm-werf-beb7eccecdfe).

### If the deploy failed

In the case of failure during the release process, werf would create a new release having the FAILED state. This state can then be inspected by the user to find the problem and solve it on the next deploy invocation.

On the next deploy invocation, werf would roll back the release to the last successful version. During the rollback, all release resources will be restored to the previous working version before applying any new changes to the resource manifests.

This rollback step will be abandoned when a [3-way-merge method](#method-of-applying-changes) of applying changes will be implemented.

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

### Configuring resource tracking

Tracking can be configured for each resource using resource annotations:

 * [`werf.io/track-termination-mode`](#track-termination-mode);
 * [`werf.io/fail-mode`](#fail-mode);
 * [`werf.io/failures-allowed-per-replica`](#failures-allowed-per-replica);
 * [`werf.io/log-regex`](#log-regex);
 * [`werf.io/log-regex-for-CONTAINER_NAME`](#log-regex-for-container);
 * [`werf.io/skip-logs`](#skip-logs);
 * [`werf.io/skip-logs-for-containers`](#skip-logs-for-containers);
 * [`werf.io/show-logs-only-for-containers`](#show-logs-only-for-containers);
 * [`werf.io/show-service-messages`](#show-service-messages).

All these annotations can be combined and used together for a resource.

**TIP** Use `"werf.io/track-termination-mode": NonBlocking` and `"werf.io/fail-mode": IgnoreAndContinueDeployProcess` when you need to define a Job in the release that runs in the background and does not affect the deploy process.

**TIP** Use `"werf.io/track-termination-mode": NonBlocking` annotation when you need to describe a StatefulSet object with the `OnDelete` manual update strategy, and you don't want to block the entire deploy process while waiting for the StatefulSet update.

#### Examples of using annotations

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'show-service-messages')">show-service-messages</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'skip-logs')">skip-logs</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'track-termination-mode')">NonBlocking track-termination-mode</a>
</div>
<div id="show-service-messages" class="tabs__content active">
  <img src="https://raw.githubusercontent.com/werf/demos/master/deploy/werf-new-track-modes-1.gif" />
</div>
<div id="skip-logs" class="tabs__content">
  <img src="https://raw.githubusercontent.com/werf/demos/master/deploy/werf-new-track-modes-2.gif" />
</div>
<div id="track-termination-mode" class="tabs__content">
  <img src="https://raw.githubusercontent.com/werf/demos/master/deploy/werf-new-track-modes-3.gif" />
</div>

#### Track termination mode

`"werf.io/track-termination-mode": WaitUntilResourceReady|NonBlocking`

 * `WaitUntilResourceReady` (default) —  the entire deployment process would monitor and wait for the readiness of the resource having this annotation. Since this mode is enabled by default, the deployment process would wait for all resources to be ready.
 * `NonBlocking` — the resource is tracked only if there are other resources that are not yet ready.

#### Fail mode

`"werf.io/fail-mode": FailWholeDeployProcessImmediately|HopeUntilEndOfDeployProcess|IgnoreAndContinueDeployProcess`

 * `FailWholeDeployProcessImmediately` (default) — the entire deploy process will fail with an error if an error occurs for some resource.
 * `HopeUntilEndOfDeployProcess` — when an error occurred for the resource, set this resource into the "hope" mode, and continue tracking other resources. If all remained resources are ready or in the "hope" mode, transit the resource back to "normal" and fail the whole deploy process if an error for this resource occurs once again.
 * `IgnoreAndContinueDeployProcess` — resource errors do not affect the deployment process.

#### Failures allowed per replica

`"werf.io/failures-allowed-per-replica": "NUMBER"`

By default, one error per replica is allowed before considering the whole deployment process unsuccessful. This setting is related to the [fail mode](#fail-mode): it defines a threshold, after which the fail mode comes into play.

#### Log regex

`"werf.io/log-regex": RE2_REGEX`

Defines a [Re2 regex](https://github.com/google/re2/wiki/Syntax) template that applies to all logs of all containers of all Pods owned by a resource with this annotation. werf would show only those log lines that fit the specified regex template. By default, werf shows all log lines.

#### Log regex for container

`"werf.io/log-regex-for-CONTAINER_NAME": RE2_REGEX`

Defines a [Re2 regex](https://github.com/google/re2/wiki/Syntax) template that applies to logs of `CONTAINER_NAME` containers in all Pods owned by a resource with this annotation. werf would show only those log lines that fit the specified regex template. By default, werf shows all log lines.

#### Skip logs

`"werf.io/skip-logs": "true"|"false"`

Set to `"true"` to turn off printing logs of all containers of all Pods owned by a resource with this annotation. This annotation is disabled by default.

#### Skip logs for containers

`"werf.io/skip-logs-for-containers": CONTAINER_NAME1,CONTAINER_NAME2,CONTAINER_NAME3...`

The comma-separated list of containers in all Pods owned by a resource with this annotation. werf would turn off log output for those containers.

#### Show logs only for containers

`"werf.io/show-logs-only-for-containers": CONTAINER_NAME1,CONTAINER_NAME2,CONTAINER_NAME3...`

The comma-separated list of containers in all Pods owned by a resource with this annotation. werf would show logs for these containers. Logs of containers that are not included in this list will not be printed. By default, werf displays logs of all containers of all Pods of a resource.

#### Show service messages

`"werf.io/show-service-messages": "true"|"false"`

Set to `"true"` to enable additional real-time debugging info (including Kubernetes events) for a resource during tracking. By default, werf would show these service messages only if the resource has failed the entire deploy process.

### Annotating and labeling chart resources

#### Auto annotations

werf automatically sets the following built-in annotations to all deployed chart resources:

 * `"werf.io/version": FULL_WERF_VERSION` — version of werf used when running the `werf deploy` command;
 * `"project.werf.io/name": PROJECT_NAME` — project name specified in the `werf.yaml`;
 * `"project.werf.io/env": ENV` — environment name specified via the `--env` param or `WERF_ENV` variable; optional, will not be set if env is not used.

werf also sets auto annotations containing information from the CI/CD system used (for example, GitLab CI)  when running the `werf ci-env` command prior to the `werf deploy` command. For example, [`project.werf.io/git`]({{ site.baseurl }}/reference/plugging_into_cicd/gitlab_ci.html#werf_add_annotation_project_git), [`ci.werf.io/commit`]({{ site.baseurl }}/reference/plugging_into_cicd/gitlab_ci.html#werf_add_annotation_ci_commit), [`gitlab.ci.werf.io/pipeline-url`]({{ site.baseurl }}/reference/plugging_into_cicd/gitlab_ci.html#werf_add_annotation_gitlab_ci_pipeline_url) and [`gitlab.ci.werf.io/job-url`]({{ site.baseurl }}/reference/plugging_into_cicd/gitlab_ci.html#werf_add_annotation_gitlab_ci_job_url).

For more information about the CI/CD integration, please refer to the following pages:

 * [plugging into CI/CD overview]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html);
 * [plugging into GitLab CI]({{ site.baseurl }}/reference/plugging_into_cicd/gitlab_ci.html).

#### Custom annotations and labels

The user can pass arbitrary additional annotations and labels using `--add-annotation annoName=annoValue` (can be used repeatedly) and `--add-label labelName=labelValue` (can be used repeatedly) CLI options when invoking werf deploy.

For example, you can use the following werf deploy invocation to set `commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57`, `gitlab-user-email=vasya@myproject.com`  annotations/labels to all Kubernetes resources in a chart:

```shell
werf deploy \
  --add-annotation "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-label "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-annotation "gitlab-user-email=vasya@myproject.com" \
  --add-label "gitlab-user-email=vasya@myproject.com" \
  --env dev \
  --images-repo :minikube \
  --stages-storage :local
```

### Validating resource manifests

If the resource manifest in the chart contains logic or syntax errors, then werf would print the validation warning to the output during the deployment process. Also, all validation errors will be written to the `debug.werf.io/validation-messages`. These errors typically do not affect the exit status of the deploy process since Kubernetes apiserver can accept incorrect manifests containing certain typos or errors without any warnings.

Let us suppose there are the following typos in the chart template (`envs` in place of `env` and `redinessProbe` in lieu of `readinessProbe`):

```
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
