---
title: Bundles
permalink: usage/distribute/bundles.html
change_canonical: true
published: false
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
 - publishing of project's [helm chart]({{ "/usage/deploy/charts.html" | true_relative_url }}) files into the container registry;
 - saving of passed [helm chart values]({{ "/usage/deploy/values.html" | true_relative_url }}) params in the published bundle;
 - saving of passed annotations and labels in the published bundle.

### Bundle structure

Before publication we have a regular chart in werf project, which looks like follows:

```
.helm/
  Chart.yaml
  LICENSE
  README.md
  values.yaml
  values.schema.json
  templates/
  crds/
  charts/
  files/
```

Only specified above files and directories will be packaged into bundle during publication by default.

`.helm/files` is conventional directory which should contain configuration files which used in the `.Files.Get` and `.Files.Glob` directives.

`.helm/values.yaml` will be merged with werf's internal service values and another custom values and sets passed to the publish command and then saved into resulting bundle file `values.yaml`. Using of default `.helm/values.yaml` can be disabled with `--disable-default-values` option passed to `werf bundle publish` (in such case resulting bundle `values.yaml` would still exist in the published bundle and contain merged service values and another custom values and sets passwed to the publish command).

Published bundle file structure looks like follows:

```
Chart.yaml
LICENSE
README.md
values.yaml
values.schema.json
templates/
crds/
charts/
files/
extra_annotations.json
extra_labels.json
```

Note `extra_annotations.json` and `extra_labels.json` service werf's files, which used to store default extra annotations and labels that should be appended to each deployed resource when applying bundle with the `werf bundle apply` command.

To download and inspect published bundle use `werf bundle copy --from REPO:TAG --to archive:PATH_TO_ARCHIVE.tar.gz` command.

### Versioning during publication

User may choose a version of published bundle with `--tag` parameter. By default, bundle published with the `latest` tag.

When publishing a bundle by the tag which already exists in the container registry, werf will replace an existing bundle with the newer version. This ability could be used for automatic updates of the application when a newer application bundle version is available in the container registry. More info in the [auto updates of bundles](#auto-updates-of-bundles).

## Bundles deployment

Bundle published into the container registry could be deployed into the kubernetes by the werf with one of the following ways:
 1. With `werf bundle apply` command directly from container registry.
 2. From another werf project as helm chart dependency.

### Deploy with werf bundle apply

[werf-bundle-apply]({{ "/reference/cli/werf_bundle_apply.html" | true_relative_url }}) command used to deploy a published bundle version into the kubernetes. This command **does not need a project git directory** to run, because bundle contains all needed files and images to deploy an application. This command accepts params which is analogous to [werf-converge]({{ "/reference/cli/werf_converge.html" | true_relative_url }}) command. For `werf bundle apply` you should explicitly specify Helm release name (`--release`) and namespace to be used for deployment (`--namespace`).

[Values for helm chart]({{ "/usage/deploy/values.html" | true_relative_url }}), annotations and labels which has been passed to the [werf-bundle-apply]({{ "/reference/cli/werf_bundle_apply.html" | true_relative_url }}) command will be united with the values, annotations and labels, which has been passed during publication of the bundle being applied.

### Deploy as helm chart dependency

Configure dependency in the target werf project `.helm/Chart.yaml`:

```yaml
apiVersion: v2
dependencies:
- name: project
  repository: "oci://ghcr.io/group"
  version: 1.4.x
```

Update project chart dependencies (this command will actualize `.helm/Chart.lock`):

```
werf helm dependency update .helm
```

Deploy project with dependencies:

```
werf converge --repo ghcr.io/group/otherproject
```

### Versioning during deployment

User may choose a version of deployed bundle with `--tag` parameter. By default, bundle with the `latest` tag will be deployed.

werf will check that bundle has been updated in the container registry for the specified tag (or `latest`) and update an application to the latest published bundle version. This ability could be used for automatic updates of the application when a newer application bundle version is available in the container registry. More info in the [auto updating bundles](#auto-updates-of-bundles).

## Other commands to work with bundles

### Render bundle manifests

[`werf bundle render`]({{ "/reference/cli/werf_bundle_render.html" | true_relative_url }}) command renders Kubernetes manifests that can be passed to other software which will handle the deployment part (e.g. ArgoCD) or can be used just for the debugging.

Command **does not require project git directory**. One of the two options has to be specified:
1. `--repo`, to render bundle from the remote OCI-repository.
2. or `--bundle-dir`, to render bundle from the directory containing the extracted bundle.

If you are going to deploy resulting manifests, then you should provide values for the options `--release` and `--namespace`, otherwise stub values will be used for them.

Option `--output` might be used to save rendered manifests to an arbitrary file.

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
werf bundle apply --repo registry.mydomain.io/project --tag v3.4.9 --env production --release myproject --namespace myproject --set "sentry.url=sentry.mydomain.io"
```

## Auto updates of bundles

werf supports automatical redeployment of the bundle, which has been published by some statical tag like `latest` or `main`. werf will update bundle by such tag in the container registry each time publish command has been called.

When deploying bundle by such tag werf will download the latest actual version of the bundle published into the container registry.

### Deployment by semver constraint

Auto updates by semver constraint is not supported yet, but [planned](https://github.com/werf/werf/issues/3169):

```
werf bundle apply --repo registry.mydomain.io/project --tag-mask v3.5.* --release myproject --namespace myproject
```

This bundle apply should check available versions by specified mask `v3.5.*`, select latest version then deploy this version into the kubernetes.

## Supported container registries

The [Open Container Initiative (OCI) Image Spec](https://github.com/opencontainers/image-spec) support in the container registry is sufficient to work with bundles.

|                             |                   |
|-----------------------------|:-----------------:|
| _AWS ECR_                   |      **ok**       |
| _Azure CR_                  |      **ok**       |
| _Default_                   |      **ok**       |
| _Docker Hub_                |      **ok**       |
| _GCR_                       |      **ok**       |
| _GitHub Packages_           |      **ok**       |
| _GitLab Registry_           |      **ok**       |
| _Harbor_                    |      **ok**       |
| _JFrog Artifactory_         |      **ok**       |
| _Nexus_                     |  **not tested**   |
| _Quay_                      | **not supported** |
| _Yandex Container Registry_ |      **ok**       |
| _Selectel CRaaS_            |  **not tested**   |

