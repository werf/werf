---
title: Release management
permalink: usage/deploy/releases.html
---

## About releases

The deployment results in a *release* — a set of resources and service information deployed in the cluster. 

From a technical point of view, werf releases are Helm 3 releases and therefore are fully compatible with them. Service information is stored in a special Secret resource by default.

## Automatic release name generation (werf only)

By default, the release name is generated automatically using a special pattern `[[ project ]]-[[ env ]]`, where `project` is the werf project name and `env` is the environment name, for example:

```yaml
# werf.yaml:
project: myapp
```

```shell
werf converge --env staging
werf converge --env production
```

In this case, the `myapp-staging` and `myapp-production` releases will be created.

## Changing the release name pattern (werf only)

If you are not happy with the pattern werf uses to generate the release name, you can modify it:

```yaml
# werf.yaml:
project: myapp
deploy:
  helmRelease: "backend-[[ env ]]"
```

```shell
werf converge --env production
```

In this case, the `backend-production` release will be created.

## Specifying the release name explicitly

Instead of generating the release name using a special pattern, you can specify the release name explicitly:

```shell
werf converge --release backend-production  # or $WERF_RELEASE=...
```

In this case, the `backend-production` release will be created.

## Formatting the release name

A release name generated using a special pattern or specified by the `--release` flag is automatically converted to match the [RFC 1123 Label Names](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names) format. You can disable automatic formatting by setting the `deploy.helmReleaseSlug` directive in the `werf.yaml` file.

You can manually format any string to match the RFC 1123 Label Names format using the `werf slugify -f helm-release` command.

## Adding resources that already exist in the cluster to the release

werf does not allow a new release resource to be deployed on top of an existing resource in the cluster if the resource in the cluster *isn't part of the current release*. This behavior prevents accidental updates to resources belonging to another release or deployed without werf. Trying to do so will result in the following error:

```
Error: helm upgrade have failed: UPGRADE FAILED: rendered manifests contain a resource that already exists...
```

To add a cluster resource to the current release and allow it to be updated, add the `meta.helm.sh/release-name: <name of the current release>`, `meta.helm.sh/release-namespace: <Namespace of the current release>` annotations and the `app.kubernetes.io/managed-by: Helm` label to that cluster resource, for example:

```shell
kubectl annotate deploy/myapp meta.helm.sh/release-name=myapp-production
kubectl annotate deploy/myapp meta.helm.sh/release-namespace=myapp-production
kubectl label deploy/myapp app.kubernetes.io/managed-by=Helm
```

... and then restart the deployment:

```shell
werf converge
```

Running the above commands will result in the `myapp` release resource successfully updating the `myapp` resource in the cluster. On top of that, the cluster resource will become part of the current release.

## Auto-annotating the release resources being deployed

During deployment, werf automatically adds the following annotations to all chart resources:

* `"werf.io/version": FULL_WERF_VERSION` — the werf version used when running the `werf converge` command;
* `"project.werf.io/name": PROJECT_NAME` — the project name specified in the `werf.yaml` configuration file;
* `"project.werf.io/env": ENV` — the environment name specified with the `--env` parameter or the `WERF_ENV` environment variable (the annotation is not set if the environment was not set at startup).

The `werf ci-env` command, if run with [supported CI/CD systems]({{"usage/integration_with_ci_cd_systems.html" | true_relative_url }}), adds annotations that allow the user to navigate to the related pipeline, job, and commit if necessary.

## Adding arbitrary annotations and labels to the release resources being deployed

During deployment, the user can attach arbitrary annotations and labels using the CLI parameters `--add-annotation annoName=annoValue` (supports multi-use) and `--add-label labelName=labelValue` (supports multi-use). Annotations and labels can also be specified with the `WERF_ADD_LABEL*` and `WERF_ADD_ANNOTATION*` variables (example: `WERF_ADD_ANNOTATION_1=annoName1=annoValue1` and `WERF_ADD_LABEL_1=labelName1=labelValue1`).

For example, use the following command to add `commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57` and `gitlab-user-email=vasya@myproject.com` annotations/labels to all Kubernetes resources in the chart:

```shell
werf converge \
  --add-annotation "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-label "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-annotation "gitlab-user-email=vasya@myproject.com" \
  --add-label "gitlab-user-email=vasya@myproject.com" \
  --env dev \
  --repo REPO
```
