---
title: Release management
permalink: usage/deploy/releases.html
---

While the [chart]({{ "/usage/deploy/charts.html" | true_relative_url }}) is a collection of configuration files of your application, the **release** is a runtime object that represents a running instance of your application deployed via werf.

Each release has a single name and multiple versions. A new release version is created with each invocation of werf converge. For each release only latest version is the actual version, previous versions of the release are stored for history.

## Storing releases

Each release version is stored in the Kubernetes cluster itself. werf supports storing release data in Secret and ConfigMap objects in the arbitrary namespace.

By default, werf stores releases in the Secrets in the namespace of your deployed application. This is fully compatible with the default [Helm 3](https://helm.sh) configuration.

The [`werf helm list -A`]({{ "reference/cli/werf_helm_list.html" | true_relative_url }}) command lists releases created by werf. Also, the user can fetch the history of a specific release with the [`werf helm history`]({{ "reference/cli/werf_helm_history.html" | true_relative_url }}) command.

### Helm compatibility notice

werf is fully compatible with the existing Helm 3 installations since the release data is stored similarly to Helm.

Furthermore, you can use werf and Helm 3 simultaneously in the same cluster and at the same time.

You can inspect and browse releases created by werf using commands such as `helm list` and `helm get`. werf can also upgrade existing releases created by the Helm.

### Compatibility with Helm 2

Existing helm 2 releases could be converted to helm 3 with the [`werf helm migrate2to3` command]({{ "/reference/cli/werf_helm_migrate2to3.html" | true_relative_url }}).

Also [werf converge command]({{ "/reference/cli/werf_converge.html" | true_relative_url }}) detects existing helm 2 release for your project and converts it to the helm 3 automatically. Existing helm 2 release could exist in the case when your project has previously been deployed by the werf v1.1.

## Resources adoption

By default Helm and werf can only manage resources that was created by the Helm or werf itself. If you try to deploy chart with the new resource in the `.helm/templates` that already exists in the cluster and was created outside werf then the following error will occur:

```
Error: helm upgrade have failed: UPGRADE FAILED: rendered manifests contain a resource that already exists. Unable to continue with update: KIND NAME in namespace NAMESPACE exists and cannot be imported into the current release: invalid ownership metadata; label validation error: missing key "app.kubernetes.io/managed-by": must be set to "Helm"; annotation validation error: missing key "meta.helm.sh/release-name": must be set to RELEASE_NAME; annotation validation error: missing key "meta.helm.sh/release-namespace": must be set to RELEASE_NAMESPACE
```

This error prevents destructive behaviour when some existing resource would occasionally become part of helm release.

However, if it is the desired outcome, then edit this resource in the cluster and set the following annotations and labels:

```yaml
metadata:
  annotations:
    meta.helm.sh/release-name: "RELEASE_NAME"
    meta.helm.sh/release-namespace: "NAMESPACE"
  labels:
    app.kubernetes.io/managed-by: Helm
```

After the next deploy this resource will be adopted into the new release revision. Resource will be updated to the state defined in the chart manifest.

## Release name

The release name is constructed using the template `[[ project ]]-[[ env ]]` by default. Here, `[[ project ]]` refers to the [project name]({{ "reference/werf_yaml.html#project-name" | true_relative_url }}) and `[[ env ]]` refers to the specified or detected environment.

For example, for the project named `symfony-demo`, the following Helm Release names can be constructed depending on the environment:
* `symfony-demo-stage` for the `stage` environment;
* `symfony-demo-test` for the `test` environment;
* `symfony-demo-prod` for the `prod` environment.

You can redefine the release using the `--release NAME` deploy option. In that case werf would use the specified name as is.

You can also define the custom release name in the werf.yaml configuration [by setting `deploy.helmRelease`]({{ "/reference/werf_yaml.html#release-name" | true_relative_url }}).

### Slugging the release name

The name of the Helm Release constructed using the template will be slugified to fit the requirements for the release name. This procedure generates a unique and valid Helm Release name.

This is the default behavior. You can disable it by [setting `deploy.helmReleaseSlug=false`]({{ "/reference/werf_yaml.html#release-name" | true_relative_url }}) in the `werf.yaml` configuration.

## Auto annotations

werf automatically sets the following built-in annotations to all deployed chart resources:

* `"werf.io/version": FULL_WERF_VERSION` — version of werf used when running the `werf converge` command;
* `"project.werf.io/name": PROJECT_NAME` — project name specified in the `werf.yaml`;
* `"project.werf.io/env": ENV` — environment name specified via the `--env` param or `WERF_ENV` variable; optional, will not be set if env is not used.

werf also sets auto annotations containing information from the CI/CD system used (for example, GitLab CI) when running the `werf ci-env` command prior to the `werf converge` command.

## Custom annotations and labels

The user can pass arbitrary additional annotations and labels using `--add-annotation annoName=annoValue` (can be used repeatedly) and `--add-label labelName=labelValue` (can be used repeatedly) CLI options when invoking werf converge.

For example, you can use the following werf converge invocation to set `commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57`, `gitlab-user-email=vasya@myproject.com`  annotations/labels to all Kubernetes resources in a chart:

```shell
werf converge \
  --add-annotation "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-label "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-annotation "gitlab-user-email=vasya@mydomain.com" \
  --add-label "gitlab-user-email=vasya@mydomain.com" \
  --repo REPO
```
