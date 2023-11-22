---
title: Bundles and charts
permalink: usage/distribute/bundles.html
---

## About bundles and charts

A *bundle* is a way to distribute a chart and its related images as a single entity.

The `werf bundle publish` command allows you to publish the chart and its related images to be deployed later with werf. This way, access to the application's Git repository is no longer needed during deployment.

The same command can publish a chart. A chart published to the OCI repository can be a main or dependent one when used by werf, Helm, Argo CD, Flux, and similar solutions.

werf automatically adds the following data to the chart when compiling it:

* the names of the images to be built and their dynamic tags in the chart Values;
* values passed via command line parameters or environment variables to the chart Values;
* global user and service annotations and labels to be added to the chart resources when deploying using the `werf bundle apply` command.

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

2. Put the files in the main chart directory (by default, `.helm` in the Git repository root). Note that *only* the following files and directories will be included in the chart when you publish it:

   ```
   .helm/
     charts/
     templates/
     crds/
     files/
     Chart.yaml
     values.yaml
     values.schema.json
     LICENSE
     README.md
   ```

   To include additional files/directories in the bundle, set the environment variable `WERF_BUNDLE_SCHEMA_NONSTRICT=1`. In this case, *all* files and directories in the main chart directory will be published, not just the ones mentioned above.

3. The next step is to publish the bundle. Build and publish the images defined in `werf.yaml` (if any), and then publish the contents of `.helm` as an `example.org/bundles/mybundle:latest` OCI image:

   ```shell
   werf bundle publish --repo example.org/bundles/mybundle
   ```

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

## Excluding files or directories from the chart being published

The `.helmignore` file in the chart root can include filename filters that prevent files or directories from being added to the chart when it is published. The rules format is the same as [in .gitignore](https://git-scm.com/docs/gitignore) except for the following:

- `**` is not supported;

- `!` at the beginning of a line is not supported;

- `.helmignore` does not exclude itself by default.

Also, the `--disable-default-values` flag for the `werf bundle publish` command excludes the `values.yaml` file from the chart being published.

## Specifying the chart version when publishing

By default, the chart is tagged as `latest` when published. You can specify a different tag, e.g., a semantic version for the chart being published, using the `--tag` option:

```shell
werf bundle publish --repo example.org/bundles/mybundle --tag v1.0.0
```

Running the above command will result in the `example.org/bundles/mybundle:v1.0.0` chart being published.

If the OCI repository finds that a chart with this tag already exists, the chart in the repository will be overwritten.

## Changing the version of the published chart

To change the tag of a published chart, copy the bundle and add the new tag to it using the `werf bundle copy` command, for example:

```shell
werf bundle copy --from example.org/bundles/mybundle:v1.0.0 --to example.org/bundles/renamedbundle:v2.0.0
```

## Copying the bundle to a different repository

The `werf bundle copy` command provides a convenient way to copy the bundle to another repository. Besides copying the chart and its related images, this command will also update the Values in the chart with the image paths.

Example:

```shell
werf bundle copy --from example.org/bundles/mybundle:v1.0.0 --to other.example.org/bundles/mybundle:v1.0.0
```

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
