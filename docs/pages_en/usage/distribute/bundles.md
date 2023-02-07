---
title: Bundles
permalink: usage/distribute/bundles.html
change_canonical: true
published: false
---

## Bundles and charts

The *charts* in werf are Helm charts with some extra capabilities. *Bundles* in werf are essentially charts in the OCI repository which are used for the same purposes as regular charts, but provide a range of additional features:

* saving the names of the images being built and their dynamic tags in the bundle Values;

* copying both the bundle and its related assembled images to another repository and automatically updating the paths to those images in the bundle Values using a single `werf bundle copy` command;

* saving global user and service annotations and labels for resources in the bundle;

* saving the Values passed via command line parameters or environment variables in the bundle.

*Any* chart published to the OCI repository is a bundle. *Any* bundle is also a chart, although when used as a regular chart, some of its advanced features, such as saved global service annotations, will not be available.

The `werf bundle publish` command is used to publish both charts and bundles. It publishes the bundle specifically, but the bundle can be used both as a bundle and as a chart since a bundle is an enhanced version of a chart.

## Publishing a bundle

You can publish the bundle to the OCI repository as follows:

1. Create `werf.yaml` if it does not exist:

   ```yaml
   # werf.yaml:
   project: mybundle
   configVersion: 1
   ```

2. Put the bundle files in the main chart directory (by default, `.helm` in the Git repository root). Note that *only* the following files and directories will be included in the bundle when you publish it:

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

3. The next step is to build and publish the images defined in `werf.yaml` (if any), and then publish the contents of `.helm` as a `example.org/bundles/mybundle:latest` bundle in the form of an OCI image:

   ```shell
   werf bundle publish --repo example.org/bundles/mybundle
   ```

## Publishing multiple bundles from a single Git repository

Place the `.helm` file with the bundle contents and its corresponding `werf.yaml` in a separate directory for each bundle:

```
bundle1/
  .helm/
    templates/
    ...
  werf.yaml
bundle2/
  .helm/
    templates/
    ...
  werf.yaml
```

You can now publish each bundle individually:

```shell
cd bundle1
werf bundle publish --repo example.org/bundles/bundle1

cd ../bundle2
werf bundle publish --repo example.org/bundles/bundle2
```

## Excluding files or directories from the bundle

The `.helmignore` file at the bundle root can include filename filters that prevent files or directories from being added to the bundle when it is published. The rules format is the same as [in .gitignore](https://git-scm.com/docs/gitignore) except for the following:

- `**` is not supported;

- `!` at the beginning of a line is not supported;

- `.helmignore` does not exclude itself by default.

Also, the `--disable-default-values` flag for the `werf bundle publish` command excludes the `values.yaml` file from the bundle being published.

## Bundle versioning

By default, the bundle is tagged as `latest` when published. You can specify a different tag, e.g., a semantic version for the published package, using the `--tag` option:

```shell
werf bundle publish --repo example.org/bundles/mybundle --tag v1.0.0
```

Running the command above will result in the `example.org/bundles/mybundle:v1.0.0` bundle being published.

If the OCI repository finds that a bundle with this tag already exists, the bundle in the repository will be overwritten.

## Copying the bundle and its images to a different repository

The `werf bundle copy` command provides a convenient way to copy the bundle and its related built images to another repository. Besides copying the bundle and images, this command will also update the Values stored in the bundle, which contain the path to the images.

Example:

```shell
werf bundle copy --from example.org/bundles/mybundle:v1.0.0 --to other.example.org/bundles/mybundle:v1.0.0
```

## Changing the name or tag of a published bundle

To change the name or tag of a published bundle, copy it under the new name/tag using the `werf bundle copy` command, for example:

```shell
werf bundle copy --from example.org/bundles/mybundle:v1.0.0 --to example.org/bundles/renamedbundle:v2.0.0
```

## Exporting the bundle and its images from the repository to an archive

After publication, the bundle and its related images can be exported from the repository to a local archive for distribution by other means using the `werf bundle copy` command, for example:

```shell
werf bundle copy --from example.org/bundles/mybundle:v1.0.0 --to archive:archive.tar.gz
```

## Importing the bundle and its images from the archive to the repository

The bundle and its related images can be imported back into the same or another OCI repository using the `werf bundle copy` command, for example:

```shell
werf bundle copy --from archive:archive.tar.gz --to other.example.org/bundles/mybundle:v1.0.0
```

Then the newly published bundle and its images can be used as usual again.

## Container registries that support the publication of bundles

Publishing bundles requires a container registry to support the OCI ([Open Container Initiative](https://github.com/opencontainers/image-spec)) specification. Below is a list of the most popular container registries that have been tested and found to be compatible:

| Container Registry        | Supports bundle publishing      |
| ------------------------- |:-------------------------------:|
| AWS ECR                   | +                               |
| Azure CR                  | +                               |
| Docker Hub                | +                               |
| GCR                       | +                               |
| GitHub Packages           | +                               |
| GitLab Registry           | +                               |
| Harbor                    | +                               |
| JFrog Artifactory         | +                               |
| Yandex Container Registry | +                               |
| Nexus                     | +                               |
| Quay                      | -                               |