---
title: Changelog
permalink: documentation/whats_new_in_v1_2/changelog.html
description: Description of key differences since v1.1
sidebar: documentation
---

This article contains full descriptive list of key changes since v1.1. If you need a fast guide about migration from v1.1 then read [how to migrate from v1.1 to v1.2 article]({{ "/documentation/whats_new_in_v1_2/how_to_migrate_from_v1_1_to_v1_2.html" | true_relative_url }}).

## Giterminism

Werf introduces the so-called giterminism mode. The word is constructed from git and determinism, which means “determined by the git”. All werf configuration files, helm configuration and application files werf reads from the current commit of the git project directory. There is also so called [dev mode](#follow-and-dev) to simplify local development of werf configuration and local development of an application with werf.

More info [at the page]({{ "/documentation/advanced/giterminism.html" | true_true_relative_url }}).

Introduced new giterminism configuration file [`werf-giterminism.yaml`]({{ "/documentation/reference/werf_giterminism_yaml.html" | true_true_relative_url }}).

## Local development

### Follow and dev

All toplevel commands: `werf converge`, `werf run`, `werf bundle publish`, `werf render` and `werf build` — has flags aimed for local development called `--follow` and `--dev`.

By default each of these toplevel commands read all needed files from the current commit of the git repo. With `--dev` flag command will read changes from the git index of the project git work tree (files added with `git add` command).

Command with the `--follow` flag will work in a loop, where either:
 - a new commit to the project git work tree will cause command rerun — by default;
 - changes to the git index of the project git work tree will cause command rerun — when `--follow` flag is combined with the `--dev` flag.

Internally werf will commit staged files to dev branch `werf-dev-<commit>`. Dev-mode cache is separated from the main cache when no `--dev` has been specified.

### Composer support

 - `werf compose config` — run docker-compose config command with forwarded image names;
 - `werf compose down` — run docker-compose down command with forwarded image names;
 - `werf compose up` — run docker-compose up command with forwarded image names.

## New bundles functionality

 - Allows splitting the process of creation of a new release for application and the process of deploying of this release into the kubernetes.
 - Two steps: 1) publish a bundle for your application version from the application git directory into container registry; 2) deploy published bundle from the container registry into the kubernetes.
 - More info [in the docs]({{ "/documentation/advanced/bundles.html" | true_relative_url }})

## Change main commands interface and behaviour

### Interface rework

 - single `werf converge` command to build and publish needed images and deploy application into the kubernetes
   - `werf converge --skip-build` emulates behaviour of old `werf deploy` command
 - removed `werf stages *` and `werf images *` commands
 - there is no more `werf publish` command, because `werf build` command with `--repo` param will publish images with all it's stages automatically
 - there is no more `--tag-by-stages-signature`, `--tag-git-branch`, `--tag-git-commit`, `--tag-git-tag`, `--tag-custom` options, `--tag-by-stages-signature` behaviour is default
   - forcing of custom tags is not yet available in the v1.2
   - werf ci-env command only accepts hidden `--tagging-strategy` flag for legacy command call: `werf ci-env --tagging-strategy` without `--as-file` flag.
     - this flag usages should be completely removed from the ci-env invocations
     - specified param for this flag does not make a difference, `--tag-by-stages-signature` behaviour will be used anyway

### werf-build and werf-converge commands behaviour

 - By default `werf build` command does not require any arguments and will build cache in the local docker server images storage.
 - When running `werf build` with `--repo registry.mydomain.org/project` param werf will lookup for already built images in the local docker server images storage and upload these if found any.
 - `werf converge` command requires `--repo` to work, but will automatically upload into the repo any locally existing images.

### Automatical images building in some toplevel commands

 - `werf converge`, `werf run`, `werf bundle publish` and `werf render` commands will automatically build needed images, which does not exists in the repo.
   - `werf render` will build images only when `--repo` param has been specified.

### Storing of images in CI/CD systems

Use project's container registry repository as `--repo` param for werf commands.

 - `werf ci-env` command for GitLab CI/CD for example will generate `WERF_REPO=$CI_REGISTRY_IMAGE`.
 - Specified `--repo` repository will be used as storage both for built images stages and final images (final images consists of stages anyway).

## Always store image stages in the container registry

 - werf-converge always stores images and all its stages in the container registry
 - single `--repo` param to specify storage of images with all its stages
 - due to the usage of content-based tagging final built image is the same as the last stage of this image
   - and as image consists of multiple stages, all these stages are stored in the container registry anyway
     - werf build process is host independent and will reuse intermediate stages of image from the container registry as a build cache
 - werf-build command could store built images locally or in container registry
   - but werf-converge command requires --repo to store images
   - werf-converge will copy locally built images if available into the specified repo instead of building these images (which was built with werf-build command)

## Signatures changes

 - **Signature renamed to digest**.
 - All signatures/digests of built images has been changed.
 - Dockerfile images and stapel images will be rebuilt.
   - Cache from the v1.1 is not valid anymore.

## Stapel builder

 - Remove the ability to cache each instruction separately with `asLayers` directive.

## `--env` param is optional

 - For `werf converge`, `werf render` and `werf bundle publish` commands there is `--env` param.
 - `--env` param affects used helm release name and kubernetes namespace as in v1.1.
   - When `--env` has been specified helm release will be generated by the template `[[ project ]]-[[ env ]]` and kubernetes namespace will be generated by the same template `[[ project ]]-[[ env ]]` — this is the same as in v1.1.
 - When no `--env` param has been specified, then werf will use `[[ project ]]` template for the helm release and the same `[[ project ]]` template for the kubernetes namespace generation.

## New werf-render command

 - `werf render` will build needed images automatically.
 - `werf render` work in giterminism mode.
 - this command allows reproducing exact templates in the same way as it will be rendered in werf-converge command.

## Helm

### Helm 3

 - Helm 3 is default and the only choice when deploying with werf v1.2.
 - Already existing helm 2 release will be migrated to helm 3 automatically in the werf-converge command given that helm 2 release has the same name as newly deployed helm 3 release.
   - **CAUTION** There is no way back once helm 2 release has been migrated to helm 3.
   - Werf-converge command will check that project helm charts are correctly rendered before running the process of migration.

### Configuration

 - `.Values.werf.image.IMAGE_NAME` instead of werf_image
 - Removed `werf_container_env` helm template function
 - Giterministic loading of all chart files (including subcharts)
 - Environment passed with `--env` param is available at the `.Values.werf.env`.
 - Fix auto generation of `.helm/Chart.yaml`. This makes Chart.yaml optional, but werf will use it when it exists.
   - Take `.helm/Chart.yaml` from the repository if exists.
   - Override `metadata.name` field with werf project name from the `werf.yaml`.
   - Set `metadata.version = 1.0.0` if not set.
 - Added `.Values.global.werf.version` service value with werf cli util version.
 - Support setting initial number of replicas when HPA is active for Deployment. Set `"werf.io/replicas-on-creation": NUM` annotation, do not set `spec.replicas` field in templates in such case explicitly.
     - `spec.replicas` will override `werf.io/replicas-on-creation`.
     - This annotation is especially useful when HPA is active, [here is the reason]({{ "/documentation/reference/deploy_annotations.html#replicas-on-creation" | true_relative_url }}).

### Working with subcharts

 - Lock file with dependencies should be committed .helm/Chart.lock.
 - Werf will automatically download dependencies charts using this lock file.
 - Typically .helm/charts directory should be added to .gitignore file
 - Werf allows storing chart dependencies directly in the .helm/charts though
   - These charts will override charts automatically downloaded by the .helm/Chart.lock file

### Deleted werf-helm-deploy-chart command

 - Use `werf helm install/upgrade` native helm commands instead.

### Better integration with werf-helm-* commands

 - `werf helm template` command could be used to render chart templates even if custom werf templates function are used (like `werf_secret_file`).
 - secret files are fully supported by werf-helm-* commands.
 - custom annotations and labels are fully supported by werf-helm-* commands (`--add-annotation` and `--add-label` options).

### Changed service values format

 - Do not use `.Values.global.werf.image` section, use `.Values.werf.image` instead.

### Deploy process

 - Wait until all release resources terminated in the `werf dismiss` command before exiting.

## werf.yaml

 - Rename `fromImageArtifact` to `fromArtifact`
 - Removed `--helm-chart-dir` option, define helm chart dir in the `werf.yaml`:

    ```
    configVersion: 1
    deploy:
      helmChartDir: .helm
    ```

 - Move `--allow-git-shallow-clone`, `--git-unshallow` and `--git-history-synchronization` options into the `werf.yaml`, added [new gitWorktree meta section]({{ "/documentation/reference/werf_yaml.html#git-worktree" | true_relative_url }}). Disabled ci-env generation of old options. Always unshallow git clone by default.
 - Use sprig v3 instead of v2: http://masterminds.github.io/sprig/.
 - New documentation page describing [werf.yaml template engine]({{ "/documentation/reference/werf_yaml_template_engine.html" | true_relative_url }}).
 - Fix naming for the user template. The template name is the path relative to the templates directory. {% raw %}{{ include ".werf/templates/1.tmpl" . }} => {{ include "templates/1.tmpl" . }}{% endraw %}.
 - Nameless image declaration `image: ~` is deprecated, it is better to always name your images. Will be removed in v1.3.
 - `env "ENVIRONMENT_VARIABLE_NAME"` function requires `ENVIRONMENT_VARIABLE_NAME` variable to exists (it may contain empty value, but should be explicitly defined).
 - The new `required` function gives developers the ability to declare a value entry as required for config rendering. If the value is empty, the config will not render and will return an error message supplied by the developer.

    {% raw %}
    ```
    {{ required "A valid <anything> value required!" <anything> }}
    ```
    {% endraw %}

## Tagging strategy

 - Only content-based.
 - Export of built images coming soon.
 - Forcing custom built tags coming soon.

## Cleanup

 - Git history based by default.
   - Previously this mode was enabled with `--git-history-based-cleanup-v1.2` option.
     - Option has been removed.
 - Removed tag strategy based algorithms completely.
 - `--keep-stages-built-within-last-n-hours` option to keep images that were built within period (2 hours by default).
 - Git-history based cleanup improvement:
   - The keep policy imagesPerReference.last counts several stages that were built on the same commit as one.
   - Thus, the policy may cover more tags than expected, and this behavior is correct.

## Caching of stapel artifact/image imports by files checksum

 - Suppose we have an artifact which builds some files.
 - This artifact has a stage-dependency in the werf.yaml to rebuild these files when some source files has been changed.
 - When source files has been changed werf rebuilds this artifact.
   - Because stage-dependency has changed, last artifact stage will also have a different digest.
   - But after artifact rebuild factual checksum of built files not changed — the same files.
     - In the v1.1 werf would still perform import of these files into the target image.
       - As a consequence target image will be unnecessarily rebuilt and redeployed.
     - In the v1.2 werf would check checksum of imported files and perform reimport only when checksum changed.

## Primary and secondary images storage support

 - Automatically upload locally built images into the specified `--repo`.
 - Use stages from read-only secondary images storage specified by `--secodary-repo` options (can be specified multiple times).
   - Suitable stages from the `--secondary-repo` will be copied into the primary `--repo`.

## Built images report format changes

`werf converge`, `werf build`, `werf run`, `werf bundle publish` and `werf render` commands has `--report-path` and `--report-format` options. `--report-path` option enables generation of built images report in the following format:

```
# werf build --report-path images.json --report-format json

{
  "Images": {
    "result": {
      "WerfImageName": "result",
      "DockerRepo": "quickstart-application",
      "DockerTag": "32e88a6a19a425c9254374ee2899b365876de31ac7d6857b523696a1-1613371915843",
      "DockerImageID": "sha256:fa6445196bc8ed44e4d2842eeb068aab4e627112d504334f2e56d235993ba4f0",
      "DockerImageName": "quickstart-application:32e88a6a19a425c9254374ee2899b365876de31ac7d6857b523696a1-1613371915843"
    },
    "vote": {
      "WerfImageName": "vote",
      "DockerRepo": "quickstart-application",
      "DockerTag": "45f03bdd90c844eb2e61e7e01dae491588d2bdadbd195881b25be9b0-1613371915351",
      "DockerImageID": "sha256:a6845bbc7912e45c601b0291170f9f503722efceb9e3cc98a5701ea4d26b017e",
      "DockerImageName": "quickstart-application:45f03bdd90c844eb2e61e7e01dae491588d2bdadbd195881b25be9b0-1613371915351"
    },
    "worker": {
      "WerfImageName": "worker",
      "DockerRepo": "quickstart-application",
      "DockerTag": "1b16118e7d5c67aa3c61fc0f8d49b3eccf8f72810f01c33a40290418-1613371916044",
      "DockerImageID": "sha256:546af94bd73dc20a7ab1f49562f42c547ee388fb75cf480eae9fde02b48ad6ad",
      "DockerImageName": "quickstart-application:1b16118e7d5c67aa3c61fc0f8d49b3eccf8f72810f01c33a40290418-1613371916044"
    }
  }
}
```

or

```
# werf build --report-path images.sh --report-format envfile

WERF_RESULT_DOCKER_IMAGE_NAME=quickstart-application:32e88a6a19a425c9254374ee2899b365876de31ac7d6857b523696a1-1613371915843
WERF_WORKER_DOCKER_IMAGE_NAME=quickstart-application:1b16118e7d5c67aa3c61fc0f8d49b3eccf8f72810f01c33a40290418-1613371916044
WERF_VOTE_DOCKER_IMAGE_NAME=quickstart-application:45f03bdd90c844eb2e61e7e01dae491588d2bdadbd195881b25be9b0-1613371915351
```

## Support multiple werf.yaml in the single git repo

User may put werf.yaml (and .helm) into any subdirectory of the project git repo.

 - All relative paths specified in the werf.yaml will be calculated relatively to the werf process cwd or `--dir` param.
 - Typically user will run werf from the subdirectory where werf.yaml reside.
 - There is also `--config` option to pass custom werf.yaml config, all relative paths will also be calculated relatively to the werf process cwd or `--dir` param.
 - Added `--git-work-tree` param (or `WERF_GIT_WORK_TREE` variable) to specify directory that contains `.git` in the case when autodetector failed or we want to use specific work tree.
     - For example when running werf from the submodule of the project we may want to use root repo worktree instead of submodule's work tree.

## Other and internals

 - Added git-archives and git-patches cache in the `~/.werf/local_cache/` along with git-worktrees. Do not create patches and archives in `/tmp` (or `--tmp-dir`).
 -  Remove stages_and_images read-only lock during build process
   - stages_and_images read only lock is preventive measure against running cleanup and build commands at the same time;
   - this lock creates unnecessary load on the synchronization server, because this lock working all the time the build process is active for each build process;
   - it is mostly safe to just omit this lock completely;
   - also removed unused "image" lock from the lock manager (legacy from v1.1).
   - also renamed kubernetes-related stage and stage-cache locks to include project name (just for more correctness, incompatible with v1.1 change).

## Documentation rework

 - Split all documentation into sections by the level of the understanding:
   - landing page;
   - introduction;
   - quickstart;
   - guides;
   - reference;
   - advanced;
   - internals.
 - New guides section.
