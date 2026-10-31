---
title: Bundles and charts
permalink: usage/distribute/bundles.html
---

## About bundles and charts

A *bundle* is a way to distribute a chart and its related images as a single entity.

The `werf bundle publish` command builds the images and publishes them together with the chart to be deployed later with werf. This way, access to the application's Git repository is no longer needed during deployment.

The published chart is an ordinary Helm OCI chart: the source chart as is, plus the `.werf-data/` directory at its root. `.werf-data/` holds everything werf adds at publish time:

* the images built and their tags, as werf auxiliary parameters;
* the values passed with the `--values`, `--secret-values`, `--set*`, `--set-root-json` options and the `--patches` files;
* the user and service annotations and labels to be added to the chart resources;
* the release name and Namespace resolved for the environment.

`werf bundle apply` (as well as `werf bundle plan` and `werf bundle render`) uses all of it, so publishing and then applying the chart deploys the same as `werf converge` with the same options. Helm, Argo CD, Flux and similar solutions can install the chart too, but they get only the chart's own values by default, see [Using the chart with Helm, Argo CD and Flux](#using-the-chart-with-helm-argo-cd-and-flux).

The published bundle (a chart and its related images) can be copied to another container registry or extracted from/added to the archive using the `werf bundle copy` command.

{% include /pages/en/cr_login.md.liquid %}

## Publishing a bundle

You can publish the bundle to the OCI repository as follows:

1. Create `werf.yaml` if it does not exist:

   ```yaml
   # werf.yaml:
   project: mybundle
   configVersion: 1
   ```

2. Put the files in the main chart directory (by default, `.helm` in the Git repository root). All files of the chart directory are published, except for the ones excluded by `.helmignore` (see [below](#excluding-files-or-directories-from-the-chart)).

3. The next step is to publish the bundle. Build and publish the images defined in `werf.yaml` (if any), and then publish the contents of `.helm` as the `example.org/bundles/mybundle` chart under the `latest` tag:

   ```shell
   werf bundle publish --repo example.org/bundles/mybundle
   ```

## What the published chart contains

The chart directory is published as is: werf does not merge anything into `values.yaml`, and `secret-values.yaml` and the `secret/` directory of the chart stay where they are, encrypted. werf changes only the following fields in `Chart.yaml`, keeping the rest of the file, comments and key order included:

* `name` — the last path component of the repository (see [Chart name and version](#chart-name-and-version));
* `version` — the version derived from the tag;
* `appVersion` — only if it is set in `werf.yaml`;
* the `helm.sh/images` annotation — the werf-built images and their references, added to the entries the chart author wrote. If an entry written by the author has the same name as a werf image, publishing fails: remove the entry from `Chart.yaml` or rename the werf image.

werf also vendors the dependent charts listed in `Chart.lock` into `charts/` and builds TypeScript charts.

Everything werf adds at publish time goes to `.werf-data/`:

```
.werf-data/
  metadata.yaml        # the format version (apiVersion: v2), the --set* options, annotations, labels, the release name and Namespace
  values.yaml          # the --values files, merged in order
  secret-values.yaml   # the --secret-values files, merged in order, encrypted
  extra-values.yaml    # werf auxiliary parameters: images, repository, environment, commit
  set-files/0, 1, ...  # the contents of the --set-file files
  patches/0, 1, ...    # the --patches files
```

`.werf-data/` is always published and kept when the chart is unpacked and packed again, even if `.helmignore` of the chart matches it.

> **Note**: `.werf-data/values.yaml`, the `set-files/` files and the `--set*` strings are stored in the registry in plaintext. Pass secrets only with `--secret-values`: werf checks that every value in these files is encrypted and stores them encrypted.

## Publishing multiple bundles from a single Git repository

Place the `.helm` file with the chart contents and its corresponding `werf.yaml` in a separate directory for each bundle:

```
bundle1/
  .helm/
    templates/
    # ...
  werf.yaml
bundle2/
  .helm/
    templates/
    # ...
  werf.yaml
```

You can now publish each bundle individually:

```shell
cd bundle1
werf bundle publish --repo example.org/bundles/bundle1

cd ../bundle2
werf bundle publish --repo example.org/bundles/bundle2
```

## Excluding files or directories from the chart

The `.helmignore` file of the chart filters the published chart the same way it filters any chart werf reads, see [Excluding files or directories from the chart]({{ "/usage/deploy/charts.html#excluding-files-or-directories-from-the-chart" | true_relative_url }}).

The `--disable-default-values` and `--disable-default-secret-values` options of the `werf bundle publish` command are recorded in the chart: `werf bundle apply` then ignores `values.yaml` and `secret-values.yaml` of the chart, as `werf converge` does with these options. The files themselves stay in the chart, and Helm still uses `values.yaml`.

## Using custom image tags

To use custom image tags for bundles, you need to set the naming template using the `--use-custom-tag` option (`WERF_USE_CUSTOM_TAG`).

Example usage:

```shell
werf bundle publish --repo example.org/bundles/mybundle --use-custom-tag "%image%-my-custom-tag"
```

You can read more about custom tags in the [documentation]({{ "/usage/build/process.html#adding-custom-tags" | true_relative_url }}).

## Chart name and version

The chart name is always the last path component of the repository the chart is published to: `werf bundle publish --repo example.org/bundles/mybundle` publishes the `mybundle` chart. Helm finds an OCI chart only in a repository named after the chart.

By default, the chart is published under the `latest` tag. You can specify a different tag, e.g., a semantic version, using the `--tag` option:

```shell
werf bundle publish --repo example.org/bundles/mybundle --tag v1.0.0
```

The chart version is derived from the tag:

* a strict semantic version (`MAJOR.MINOR.PATCH` with an optional prerelease and build, optionally prefixed with `v`) is used as is, without the `v` prefix: `v1.0.0` → `1.0.0`;
* any other tag (`latest`, a commit SHA, a branch name, `1.2`, a number such as a date or a pipeline ID) gets the `0.0.0-<unix time>-<tag slug>` version, where the slug is the lowercased tag with each run of characters other than letters and digits replaced by `-`: `feature/Foo_1` → `0.0.0-<unix time>-feature-foo-1`.

werf pushes the chart under its version tag and also under the given tag if it differs, with `+` replaced by `_` in both, as Helm does: `--tag 1.0.0+build.5` publishes `1.0.0_build.5`. The command above publishes `example.org/bundles/mybundle:1.0.0` and `example.org/bundles/mybundle:v1.0.0`, both pointing to the same chart.

A `0.0.0-…` version is a prerelease, so Helm, Argo CD and Flux skip it when selecting a version by a semver range: use a range like `>=0.0.0-0` or Helm's `--devel` option to select it. Two publications with a non-semver tag within the same second get the same version and overwrite each other. Such versions accumulate in the registry: werf cleanup does not remove them.

If the OCI repository finds that a chart with this tag already exists, the chart in the repository will be overwritten.

## Deploying a bundle

`werf bundle apply` deploys the published chart without access to the application's Git repository:

```shell
werf bundle apply --repo example.org/bundles/mybundle --tag v1.0.0
```

`werf bundle publish` followed by `werf bundle apply` deploys the same as `werf converge` with the options that were passed to `werf bundle publish`, and `werf release get --print-values` shows the same values. The options passed to `werf bundle apply` (`--values`, `--secret-values`, `--set*`, `--set-root-json`, `--patches`, `--add-annotation`, `--add-label`) are applied on top of the ones stored at publish time and win over them. The parameters are combined in the following order, from the lowest priority to the highest:

1. `values.yaml` and `secret-values.yaml` of the chart.
2. The `--values`, `--secret-values`, `--set*` and `--set-root-json` options of `werf bundle publish`.
3. werf auxiliary parameters computed at publish time (images, repository, environment, commit).
4. The `--values`, `--secret-values`, `--set*` and `--set-root-json` options of `werf bundle apply`.
5. werf auxiliary parameters computed at apply time (environment, Namespace and others).

The `--patches` files are applied in the same order: the chart's own `patches.yaml`, then the ones of `werf bundle publish`, then the ones of `werf bundle apply`. The `--disable-default-values` and `--disable-default-secret-values` options of `werf bundle apply` ignore only `values.yaml` and `secret-values.yaml` of the chart, not the values stored at publish time.

The `--secret-values` of `werf bundle publish` are decrypted at apply time and saved in the release in decrypted form, as `werf converge` saves its `--secret-values`, see [Secret parameters in the release]({{ "/usage/deploy/values.html#secret-parameters-in-the-release" | true_relative_url }}). Decrypting them requires the secret key they were encrypted with at publish time.

`werf bundle publish` records the release name and Namespace resolved the same way as `werf converge` resolves them for the given `--env`. `werf bundle apply`, `werf bundle plan` and `werf bundle render` use them for `--release` and `--namespace` that are not passed, also when `--env` is not passed. If you pass a non-empty `--env` that differs from the one used at publish time, the recorded ones are not used and both `--release` and `--namespace` are required.

The one known difference from `werf converge` is `.Chart.Name` and `.Chart.Version`: under `werf converge` templates get the project name and the version from `Chart.yaml` (`1.0.0` by default), under `werf bundle apply` they get the repository name and the version derived from the tag. A label such as {% raw %}`helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}`{% endraw %} therefore differs, and moving a release from `werf converge` to `werf bundle apply` restarts its Pods once.

`werf converge`, `werf plan`, `werf render` and `werf lint` also use `.werf-data/` if they find it in the chart directory (for example, when `.helm` contains an unpacked published chart): its values, annotations and labels are applied below the options of the command.

## Using a published chart as a dependent chart

When werf or Nelm deploys a chart that includes a werf-published chart as a dependent chart, the values from `.werf-data/` of the dependent chart become its default values, as if it were deployed alone with `werf converge`. The parent chart overrides them as usual. The secret values of the dependent chart are decrypted with the secret key of the deployment, and, like the parameters from `secret-values.yaml` of any dependent chart, are saved in the release in decrypted form. Its annotations and labels are ignored with a warning: the parent chart decides them.

## Changing a published chart

A published chart can be downloaded, changed and published again without the application's Git repository, for example, with the [Nelm](https://github.com/werf/nelm) CLI:

1. Download and unpack the chart:

   ```shell
   nelm chart download oci://example.org/bundles/mybundle --version 1.0.0 --untar
   ```

2. Change the chart: `values.yaml`, secrets (`nelm chart secret values-file edit mybundle/secret-values.yaml` or `mybundle/.werf-data/secret-values.yaml`), templates, or any file in `.werf-data/`.

3. Pack the chart and upload it:

   ```shell
   nelm chart pack mybundle --set replicas=3
   nelm chart upload mybundle-1.0.0.tgz oci://example.org/bundles
   ```

`nelm chart pack` keeps the version from `Chart.yaml` unless `--version` is passed. The `--values`, `--secret-values`, `--set*`, `--set-root-json` and `--patches` options of `nelm chart pack` are added on top of the ones stored in `.werf-data/`, as if they were appended to the original `werf bundle publish` command line: the values files are merged over the stored ones, the `--set*`, `--set-root-json` and `--patches` options are appended to the stored ones, and the new annotations and labels win over the stored ones with the same keys. The `--reset-values` option drops everything stored in `.werf-data/` before applying the options of this pack. To remove a single stored option, edit `.werf-data/` and pack the chart. `nelm chart pack` does not need the secret key.

A TypeScript chart is rebuilt when packed, which requires network access to download its modules. To reuse the bundled TypeScript code of the downloaded chart instead, pass `--ts-no-rebuild`.

A chart published by werf 3.12 or earlier can be repacked too: `nelm chart pack` moves its `extra_annotations.json` and `extra_labels.json` into `.werf-data/`, while its `values.yaml`, which contains the values merged at publish time, stays the chart's own `values.yaml`.

Instead of `nelm chart pack` and `nelm chart upload`, you can use `helm package` and `helm push`: they keep `.werf-data/` unless the chart's `.helmignore` matches it.

## Changing the version of the published chart

To change the tag of a published chart, copy the bundle and add the new tag to it using the `werf bundle copy` command, for example:

```shell
werf bundle copy --from example.org/bundles/mybundle:v1.0.0 --to example.org/bundles/renamedbundle:v2.0.0
```

`werf bundle copy --to REPO:TAG` sets the version of the copy from `TAG` by the [same rules](#chart-name-and-version) as `werf bundle publish` and pushes the copy under the version tag and the given tag. `werf bundle copy --to REPO` without a tag keeps the version and tag of the source chart. The copy is named after the last path component of the destination repository.

## Copying the bundle to a different repository

The `werf bundle copy` command provides a convenient way to copy the bundle to another repository. Besides copying the chart and its related images, this command regenerates the image references in `.werf-data/extra-values.yaml` and in the `helm.sh/images` annotation for the destination repository.

Example:

```shell
werf bundle copy --from example.org/bundles/mybundle:v1.0.0 --to other.example.org/bundles/mybundle:v1.0.0
```

The image references in a werf-published chart included as a dependent chart are not rewritten: it keeps the references to its original repository.

A signed chart (with a Helm provenance file) keeps its signature when copied unchanged. If the copy changes the chart, e.g. rewrites the image references, renames it or changes its version, the signature is dropped with a warning.

`werf bundle apply`, `werf bundle plan` and `werf bundle render` verify the signature of the bundle pulled from the container registry against the public keys from `--provenance-keyring` when `--provenance-strategy=verify` (an unsigned bundle fails) or `--provenance-strategy=verify-if-possible` (only a signed bundle is verified) is passed. `werf bundle render --bundle-dir` can't verify the signature of an unpacked bundle and fails with `--provenance-strategy=verify`.

## Exporting the bundle from the container registry to the archive

After publication, the bundle can be exported from the repository to a local archive for distribution by other means using the `werf bundle copy` command, for example:

```shell
werf bundle copy --from example.org/bundles/mybundle:v1.0.0 --to archive:archive.tar.gz
```

## Importing the bundle from the archive to the repository

The exported to the archive bundle can be imported back into the same or another OCI repository using the `werf bundle copy` command, for example:

```shell
werf bundle copy --from archive:archive.tar.gz --to other.example.org/bundles/mybundle:v1.0.0
```

Then the newly published bundle (a chart and its images) can be used as usual.

## Using the chart with Helm, Argo CD and Flux

Helm, Argo CD and Flux install the published chart with its own `values.yaml` and the values they are given at install time. They do not read `.werf-data/` by themselves, so the image references and the values passed to `werf bundle publish` are missing: templates that use `.Values.werf.image.*` or `.Values.global.werf.images.*` render empty references.

> **Note**: charts published by werf 3.12 and earlier merged the image references and the `--values` and `--set*` options of `werf bundle publish` into `values.yaml`. The new charts do not, so a Helm, Argo CD or Flux deployment of a newly published chart needs the values files below.

`.werf-data/values.yaml` (the `--values` of `werf bundle publish`) and `.werf-data/extra-values.yaml` (the image references and other werf auxiliary parameters) are plain YAML files inside the chart, so you can pass them as values files.

Flux, a `HelmRelease` whose chart comes from a `HelmRepository` of the `oci` type:

```yaml
spec:
  chart:
    spec:
      chart: mybundle
      sourceRef:
        kind: HelmRepository
        name: bundles
      valuesFiles:
        - values.yaml
        - .werf-data/values.yaml
        - .werf-data/extra-values.yaml
```

`valuesFiles` replaces the default list of values files, so `values.yaml` must be listed first.

Argo CD, the paths are relative to the chart root:

```yaml
spec:
  source:
    repoURL: example.org/bundles
    chart: mybundle
    targetRevision: 1.0.0
    helm:
      valueFiles:
        - .werf-data/values.yaml
        - .werf-data/extra-values.yaml
```

The following stays unavailable with Helm, Argo CD and Flux even with these values files:

* the `--set*` options of `werf bundle publish`;
* secrets: Helm cannot decrypt `secret-values.yaml`, `.werf-data/secret-values.yaml` and the `secret/` files, and templates that call `werf_secret_file` fail;
* the annotations and labels of `werf bundle publish`;
* `--disable-default-values` of `werf bundle publish`: Helm still uses `values.yaml` of the chart;
* resources generated by TypeScript: Helm does not run `ts/`;
* everything from `.werf-data/` when the werf-published chart is a dependent chart of a chart deployed by Helm, Argo CD or Flux.

Helm puts `.werf-data/` into `.Files` and saves it in the release together with the rest of the chart (`.werf-data/secret-values.yaml` stays encrypted). Templates that read all chart files, e.g. with `.Files.Glob "**"`, get the `.werf-data/` files too.

## Compatibility with other werf and Nelm versions

Charts published by werf 3.12 and earlier are deployed by the current werf unchanged.

The current chart format (`apiVersion: v2` in `.werf-data/metadata.yaml`) requires werf v3 or Nelm v2 with `.werf-data/` support. The latest versions of werf v2 and Nelm v1 refuse to deploy such a chart, as well as a chart that includes it as a dependent chart. Earlier werf v2 and Nelm v1 versions, werf 3.12 and earlier, and Nelm v2 versions without `.werf-data/` support do not refuse: they deploy the chart with its own values only, without the image references, the values, secrets, annotations and labels of `werf bundle publish`.

## Container registries that support the publication of bundles

Publishing bundles requires a container registry to support the OCI ([Open Container Initiative](https://github.com/opencontainers/image-spec)) specification. Below is a list of the most popular container registries that have been tested and found to be compatible:

| Container registry        | Supports bundle publishing      |
|---------------------------|:-------------------------------:|
| AWS ECR                   |                +                |
| Azure CR                  |                +                |
| Docker Hub                |                +                |
| GCR                       |                +                |
| GitHub Packages           |                +                |
| GitLab Registry           |                +                |
| Harbor                    |                +                |
| JFrog Artifactory         |                +                |
| Yandex container registry |                +                |
| Nexus                     |                +                |
| Quay                      |                -                |
