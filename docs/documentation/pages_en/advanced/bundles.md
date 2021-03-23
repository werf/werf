---
title: Bundles
permalink: advanced/bundles.html
change_canonical: true
---

Standard werf usage flow implies usage of [werf-converge]({{ "/reference/cli/werf_converge.html" | true_relative_url }}) command to deploy a new version of your application into kubernetes. werf-converge command consists of:
 - building and publishing needed images into the container registry;
 - actualizing application running in the kubernetes cluster.

werf allows splitting the process of releasing a new application version and deploy process of this new version into the kubernetes with so called **bundles**.

**Bundle** is a collection of images and deployment configurations of the application version published into the container registry, which is ready to be deployed into the kubernetes cluster. Deployment of an application with bundles implies 2 steps:
 1. Publishing of a bundle version.
 2. Deploying of a bundle into the kubernetes.

Bundles allow implementing hybrid push-pull model to deploy an application. Publishing of a bundle version is the push part, while deploying of a bundle into the kubernetes is the pull part.

**Note**. At the moment there is no bundle operator for kubernetes, to allow fully automatic bundles deployment, but it will be added. Also, there will be an ability to auto update a bundle [by semver constraint](#deployment-by-semver-constraint).

## Bundles publication

werf publishes bundles with the [werf-bundle-publish]({{ "/reference/cli/werf_bundle_apply.html" | true_relative_url }}) command. This command accepts arguments analogous to the [werf-converge]({{ "/reference/cli/werf_converge.html" | true_relative_url }}) command and **requires project git directory** to run.

werf-bundle-publish command consists of:
 - building and publishing needed images into the container registry (the same way werf-converge does);
 - publishing of project's [helm chart]({{ "/advanced/helm/configuration/chart.html" | true_relative_url }}) files into the container registry;
 - saving of passed [helm chart values]({{ "/advanced/helm/configuration/values.html" | true_relative_url }}) params in the published bundle;
 - saving of passed [annotations and labels]({{ "/advanced/helm/deploy_process/annotating_and_labeling.html" | true_relative_url }}) in the published bundle.

### Versioning during publication

User may choose a version of published bundle with `--tag` parameter. By default, bundle published with the `latest` tag.

When publishing a bundle by the tag which already exists in the container registry, werf will replace an existing bundle with the newer version. This ability could be used for automatic updates of the application when a newer application bundle version is available in the container registry. More info in the [auto updates of bundles](#auto-updates-of-bundles).

## Bundles deployment

Bundle published into the container registry could be deployed into the kubernetes by the werf.

[werf-bundle-apply]({{ "/reference/cli/werf_bundle_apply.html" | true_relative_url }}) command used to deploy a published bundle version into the kubernetes. This command **does not need a project git directory** to run, because bundle contains all needed files and images to deploy an application. This command accepts params which is analogous to [werf-converge]({{ "/reference/cli/werf_converge.html" | true_relative_url }}) command.

[Values for helm chart]({{ "/advanced/helm/configuration/values.html" | true_relative_url }}), [annotations and labels]({{ "/advanced/helm/deploy_process/annotating_and_labeling.html" | true_relative_url }}) which has been passed to the [werf-bundle-apply]({{ "/reference/cli/werf_bundle_apply.html" | true_relative_url }}) command will be united with the values, annotations and labels, which has been passed during publication of the bundle being applied.

### Versioning during deployment

User may choose a version of deployed bundle with `--tag` parameter. By default, bundle with the `latest` tag will be deployed.

werf will check that bundle has been updated in the container registry for the specified tag (or `latest`) and update an application to the latest published bundle version. This ability could be used for automatic updates of the application when a newer application bundle version is available in the container registry. More info in the [auto updating bundles](#auto-updates-of-bundles).

## Other commands to work with bundles

### Export bundle into the directory

[werf-bundle-export]({{ "/reference/cli/werf_bundle_export.html" | true_relative_url }}) command will create bundle directory in the same way as it will be published into the container registry.

This command works the same way as [werf-bundle-publish]({{ "/reference/cli/werf_bundle_publish.html" | true_relative_url }}), has the same parameters and **requires project git directory** to run.

This command is useful to inspect a bundle before publishing and for debug purposes.

### Download published bundle

[werf-bundle-download]({{ "/reference/cli/werf_bundle_download.html" | true_relative_url }}) command allows downloading published bundle into the directory.

This command works the same way as [werf-bundle-apply]({{ "/reference/cli/werf_bundle_apply.html" | true_relative_url }}) does, has the same parameters and **does not need a project git directory** to run.

This command is useful to inspect a published bundle and for debug purposes.

## Examples

Let's publish bundle of the application by some semver version, run in the project git directory:

```
git checkout v3.5.0
werf bundle publish --repo registry.mydomain.io/project --tag v3.5.0 --add-label "project-version=v3.5.0"
```

Let's publish bundle of the application by static tag `main`, run in the project git directory:

```
git checkout main
werf bundle publish --repo registry.mydomain.io/project --tag main --set "myvalue.x=150"
```

Let's deploy published bundle by the `v3.4.9` tag, run on an arbitrary host with access to the kubernetes and container registry:

```
werf bundle apply --repo registry.mydomain.io/project --tag v3.4.9 --env production --set "sentry.url=sentry.mydomain.io"
```

## Auto updates of bundles

werf supports automatical redeployment of the bundle, which has been published by some statical tag like `latest` or `main`. werf will update bundle by such tag in the container registry each time publish command has been called.

When deploying bundle by such tag werf will download the latest actual version of the bundle published into the container registry.

### Deployment by semver constraint

Auto updates by semver constraint is not supported yet, but [planned](https://github.com/werf/werf/issues/3169):

```
werf bundle apply --repo registry.mydomain.io/project --tag-mask v3.5.*
```

This bundle apply should check available versions by specified mask `v3.5.*`, select latest version then deploy this version into the kubernetes.

## Supported container registries

The [Open Container Initiative (OCI) Image Spec](https://github.com/opencontainers/image-spec) support in the container registry is sufficient to work with bundles.

You can read more about supported container registries in a separate [article]({{ "/advanced/supported_container_registries.html" | true_relative_url }}).
