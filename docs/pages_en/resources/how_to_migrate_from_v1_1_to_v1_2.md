---
title: How to migrate from v1.1 to v1.2
permalink: resources/how_to_migrate_from_v1_1_to_v1_2.html
description: How to migrate your application from v1.1 to v1.2
---

Simply follow this guide to migrate your project from v1.1 to v1.2.

**IMPORTANT** There are v1.1 configurations that heavily use **`mount from build_dir`** and **`mount fromPath`** directives for such cases like caching of external dependencies (`Gemfile`, or `package.json`, or `go.mod`, etc.). While such `mount` directives still supported by the werf v1.2, it is not recommended for usage and is disabled by default by the [giterminism mode]({{ "/guides/changelog.html#giterminism" | true_relative_url }}) (except `mount from tmp_dir` — this has not been forbidden) because it may lead to harmful, unreliable, host-dependent and nondeterministic builds. For such configurations there will be added a new feature to store such cache in the docker images, this feature is planned to be added into the werf v1.2, but it is [**not implemented yet**](https://github.com/werf/werf/issues/3318). So if your configuration uses such mounts, then there are several options:
 1. Keep using v1.1 version until this feature is [added into the werf v1.2](https://github.com/werf/werf/issues/3318).
 2. Completely remove `mounts` usage and wait until this feature is [added into the werf v1.2](https://github.com/werf/werf/issues/3318).
 3. Describe and enable your `mounts` in werf v1.2 using the [`werf-giterminism.yaml`](https://werf.io/documentation/reference/werf_giterminism_yaml.html) (**not recommended**).

## Key moments

### 1. Use content based tagging

[Content based tagging strategy]({{ "/guides/changelog.html#tagging-strategy" | true_relative_url }}) is the default strategy and only choice to be used internally by the werf deploy process.

 - Remove `--tagging-strategy ...` param of `werf ci-env` command.
 - Remove `--tag-custom`, `--tag-git-tag`, `--tag-git-branch`, `--tag-by-stages-signature` params.

 In the case when you need a certain docker tag for a built image to exist in the container registry to be used outside the werf, then use `--save-build-report` (and optionally `--build-report-path`) option like follows:
 - `werf build/converge --save-build-report --repo REPO`;
 - `docker pull $(cat .werf-build-report.json | jq -r .Images.IMAGE_NAME_FROM_WERF_YAML.DockerImageName)`;
 - `docker tag $(cat .werf-build-report.json | jq -r .Images.IMAGE_NAME_FROM_WERF_YAML.DockerImageName) REPO:mytag`;
 - `docker push REPO:mytag`.

### 2. Use `converge` command instead of `build-and-publish` command

Use `werf converge` command instead of `werf deploy` to build and publish images, and deploy into kubernetes. All params are the same as for `werf deploy`.

Commands changes (for reference):
 - `werf build-and-publish` command has been removed, there is only `werf build` command, usage of which is optional:
 - `werf converge` will build and publish needed images by itself, if not built already.
 - You may use `werf build` command in `prebuild` CI/CD pipeline stage instead of `werf build-and-publish` command.

### 3. Use the `--repo` parameter instead of the `--stages-storage` and `--images-repo` parameters

The `werf converge` command stores both the collected images and their stages in the container registry. To specify the storage to be used for storing images and stages, you now need to specify one parameter `--repo` instead of the two previously used in v1.1 `--stages-storage` and `--images-repo`. You can read more about this parameter [in changelog](#always-store-image-stages-in-the-container-registry).

Similarly with the environment variables used earlier: instead of `WERF_STAGES_STORAGE` and `WERF_IMAGES_REPO`, one variable `WERF_REPO` is now used.

### 4. Change Helm templates

 Use `.Values.werf.image.IMAGE_NAME` instead of `werf_container_image` as follows:

    {% raw %}
    ```
    spec:
      template:
        spec:
          containers:
            - name: main
              image: {{ .Values.werf.image.myimage }}
    ```
    {% endraw %}

  > **IMPORTANT!** Notice usage of `apiVersion: v2` to allow description of dependencies directly in the `.helm/Chart.yaml` instead of `.helm/dependencies.yaml`.

Helm templates changes (for reference):
 - Remove `werf_container_env` template usage completely.
 - Use `.Values.werf.env` instead of `.Values.global.env`.
 - Use `.Values.werf.namespace` instead of `.Values.global.namespace`.
 - If the image name contains a hyphen (`-`), then the entry should look like this: {% raw %}`image: '{{ index .Values.werf.image "IMAGE-NAME" }}'`{% endraw %}.
 - Use `"werf.io/replicas-on-creation": "NUM"` annotation instead of `"werf.io/set-replicas-only-on-creation": "true"`.
   > **IMPORTANT!** Specify `"NUM"` as **string** instead of number `NUM`, annotation [will be ignored otherwise]({{ "/reference/deploy_annotations.html#replicas-on-creation" | true_relative_url }}).

   > **IMPORTANT!** Remove `spec.replicas` field, more info [in the changelog](#configuration).

### 5. Use `.helm/Chart.lock` for subcharts

 Due to [giterminism](#giterminism) werf does not allow uncommitted `.helm/charts` dir. To use subcharts specify dependencies in the `.helm/Chart.yaml` like that:

```yaml
# .helm/Chart.yaml
apiVersion: v2
dependencies:
- name: redis
  version: "12.7.4"
  repository: "https://charts.bitnami.com/bitnami"
```

> **IMPORTANT!** Notice usage of `apiVersion: v2` to allow description of dependencies directly in the `.helm/Chart.yaml` instead of `.helm/dependencies.yaml`.

Next, follow these steps:

 1. Add `.helm/charts` into the `.gitignore`.
 1. Run `werf helm dependency update` command, which will create `.helm/Chart.lock` file and `.helm/charts` dir.
 1. Commit `.helm/Chart.lock` file into the project git repo.

werf will automatically download subcharts into the cache and load subchart files in `werf converge` command (and other toplevel commands which require helm chart).

### 6. Cleanup by Git history

Remove `--git-history-based-cleanup-v1.2` option for a cleanup. werf always uses git-history cleanup in the v1.2. More info [in the changelog]({{ "/guides/changelog.html#cleanup" | true_relative_url }}) and [in the cleanup article]({{ "/usage/cleanup/cr_cleanup.html" | true_relative_url }}).

### 7. Define environment variables in `werf-giterminism.yaml`

> **NOTE!** It is not recommended to use environment variables in the `werf.yaml`, more info [in the article]({{ "/usage/project_configuration/giterminism.html" | true_relative_url }}).

If you use environment variables in your `werf.yaml`, then use the following `werf-giterminism.yaml` snippet (for example to enable `ENV_VAR1` and `ENV_VAR2` variables):

```yaml
# werf-giterminism.yaml
giterminismConfigVersion: 1
config:
  goTemplateRendering:
    allowEnvVariables:
      - ENV_VAR1
      - ENV_VAR2
```

### 8. Adjust `werf.yaml` configuration

 1. Change relative include paths from {% raw %}{{ include ".werf/templates/1.tmpl" . }}{% endraw %} to {% raw %}{{ include "templates/1.tmpl" . }}{% endraw %}.
 1. Rename `fromImageArtifact` to `fromArtifact`.
    > **NOTICE!** It is not required, but better to change `artifact` and `fromArtifact` to `image` and `fromImage` in such case. See [deprecation note in the changelog](#werfyaml).

### 9. Define custom `.helm` chart dir in the `werf.yaml`

`--helm-chart-dir` has been removed, use `deploy.helmChartDir` directive in `werf.yaml` like follows:

{% raw %}
```yaml
configVersion: 1
deploy:
  helmChartDir: .helm2
```
{% endraw %}

Consider using different layout for your project: werf v1.2 supports multiple `werf.yaml` applications in the single Git repo.
  - Instead of defining a different `deploy.helmChartDir` in your `werf.yaml`, create multiple `werf.yaml` in subfolders of your project.
  - Each subfolder will contain own `.helm` directory.
  - Run werf from the subfolder in such case.
  - All relative paths specified in the `werf.yaml` should be adjusted to the subfolder where `werf.yaml` stored.
  - Absolute paths specified with `git.add` directive should use absolute paths from the root of the git repo (these paths settings are compatible with 1.1).

### 10. Migrate to helm 3

 werf v1.2 [performs migration of existing Helm 2 release to helm 3 automatically](#helm-3).
 > Helm 2 release should have the same name as newly used helm 3 release.

 Before migrating werf will try to render and validate current `.helm/templates` and continue migration only when render succeeded.

 > **IMPORTANT!** Once project has been migrated to helm 3 there is no legal way back to the werf v1.1 and helm 2.

## Changelog

This article contains full descriptive list of changes since v1.1.

### Giterminism

werf introduces the so-called giterminism mode. The word is constructed from git and determinism, which means “determined by the git”.

All werf configuration files, helm configuration and application files werf reads from the current commit of the git project directory. There is also so called [dev mode](#follow-and-dev) to simplify local development of werf configuration and local development of an application with werf.

More info [at the page]({{ "/usage/project_configuration/giterminism.html" | true_relative_url }}).

Introduced new giterminism configuration file [`werf-giterminism.yaml`]({{ "/reference/werf_giterminism_yaml.html" | true_relative_url }}).

### Local development

#### Follow and dev

All toplevel commands: `werf converge`, `werf run`, `werf bundle publish`, `werf render` and `werf build` — has two main flags `--follow` and `--dev` aimed for local development.

By default, each of these commands reads all needed files from the current commit of the git repo. With the `--dev` flag command will read files from the project git worktree including untracked files. werf ignores changes in compliance with the rules described in `.gitignore` as well as rules that the user sets with the `--dev-ignore=<glob>` option (can be used multiple times).

Command with the `--follow` flag will work in a loop, where either:
- a new commit to the project git work tree will cause command rerun — by default;
- any changes in the git repository worktree including untracked files will cause command rerun —  when `--follow` flag is combined with the `--dev` flag.

Internally werf will commit changes to dev branch `werf-dev-<commit>`. Dev-cache is linked only to those temporal commits and will not interfere with the main cache (when no `--dev` has been specified).

#### Composer support

werf supports main composer commands. When running these commands werf will set special environment variables with full docker images names of images described in the `werf.yaml`:

```
WERF_<FORMATTED_WERF_IMAGE_NAME>_DOCKER_IMAGE_NAME=application:45f03bdd90c844eb2e61e7e01dae491588d2bdadbd195881b25be9b0-1613371915351
```

For example, given the following `werf.yaml`:

```yaml
# werf.yaml
project: myproj
configVersion: 1
---
image: myimage
from: alpine
```

`myimage` full docker image name could be used in the `docker-compose.yml` like that:

```yaml
version: '3'
services:
  web:
    image: "${WERF_MYIMAGE_DOCKER_IMAGE_NAME}"
```

werf provides following compose commands:
- [`werf compose config`]({{ "/reference/cli/werf_compose_config.html" | true_relative_url }}) — run docker-compose config command with forwarded image names;
- [`werf compose down`]({{ "/reference/cli/werf_compose_down.html" | true_relative_url }}) — run docker-compose down command with forwarded image names;
- [`werf compose up`]({{ "/reference/cli/werf_compose_up.html" | true_relative_url }}) — run docker-compose up command with forwarded image names.

### New bundles functionality

- Bundles allow splitting the process of creation of a new release of the source code for an application and the process of deploying of this release into the kubernetes.
- Bundles are stored in the container registry.
- Working with bundles implies two steps: 1) publishing a bundle for your application version from the application git directory into the container registry; 2) deploy published bundle from the container registry into the kubernetes.

### Change main commands interface and behaviour

#### Interface rework

- `werf deploy` is removed. Use `werf converge` — a single command to build and publish needed images into the container registry and deploy application into the kubernetes.
  - Use `werf converge --require-built-images` to emulate `werf deploy` behaviour: werf will exit with error in case needed images was not found in the container registry.
- Commands `werf stages *`, `werf images *` and `werf host project *` are removed.
- There is no more `werf publish` command, because `werf build` command with `--repo` param will publish images with all it's stages into the container registry automatically.
- There is no more `--tag-by-stages-signature`, `--tag-git-branch`, `--tag-git-commit`, `--tag-git-tag`, `--tag-custom` options, `--tag-by-stages-signature` behaviour is default.
  - Forcing of custom tags is [not yet available](https://github.com/werf/werf/issues/2869) in the v1.2
  - `werf ci-env` command only accepts hidden `--tagging-strategy` flag for legacy command call: `werf ci-env --tagging-strategy` without `--as-file` flag.
    - This flag usages should be completely removed from the `werf ci-env` invocations.
    - Specified param for this flag does not make a difference, `--tag-by-stages-signature` behaviour will be used anyway.

#### werf-build and werf-converge commands behaviour

- By default `werf build` command does not require any arguments and will build cache in the local docker server images storage.
- When running `werf build` with `--repo registry.mydomain.org/project` param werf will lookup for already built images in the local docker server images storage and upload them if found any (without unnecessary rebuild).
- `werf converge` command requires `--repo` to work, but as `werf build` command will automatically upload into the repo any locally existing images.

#### Cleaning commands behaviour

- The `werf cleanup/purge` commands only for cleaning the project container registry.
- The `werf host purge --project=<project-name>` for removing the project images in the local Docker (previously, it was possible to use the `werf purge` command).

#### Automatical images building in some toplevel commands

- `werf converge`, `werf run`, `werf bundle publish` and `werf render` commands will automatically build needed images, which does not exist in the repo.
  - `werf render` will build images only when `--repo` param has been specified.

#### Storing of images in CI/CD systems

Use project's container registry repository as `--repo` param for werf commands.

- `werf ci-env` command for GitLab CI/CD for example will generate `WERF_REPO=$CI_REGISTRY_IMAGE`.
- Specified `--repo` repository will be used as storage both for built [images stages and final images](#always-store-image-stages-in-the-container-registry).

### Always store image stages in the container registry

- `werf converge` command always stores images and all its stages in the container registry.
- Single `--repo` param to specify storage of images with all its stages (there was separate `--stages-storage` and `--images-repo` params in the v1.1).
- Due to the usage of content-based tagging final built image is the same as the last stage of this image.
  - And as image consists of multiple stages, all these stages are stored in the container registry anyway.
    - So storing of intermediate stages of the image does not create overhead for the container registry.
    - This makes werf build process host independent, because werf will reuse intermediate stages of an image from the container registry as a build cache.

### Signature changes

- **Signature renamed to digest**.
- All signatures/digests of built images has been changed.
- Dockerfile images and stapel images will be rebuilt.
  - In other words: build cache from the v1.1 is not valid anymore.

### Stapel builder

- Remove the ability to cache each instruction separately with `asLayers` directive.

### Optional env param

- For `werf converge`, `werf render` and `werf bundle publish` commands there is `--env` param.
- `--env` param [affects helm release name and kubernetes namespace]({{ "/usage/deploy/environments.html" | true_relative_url }}) as in v1.1.
  - When `--env` has been specified [helm release]({{ "/usage/deploy/releases.html" | true_relative_url }}) name will be generated by the template `[[ project ]]-[[ env ]]` and kubernetes namespace will be generated by the same template `[[ project ]]-[[ env ]]` — this is the same as in v1.1.
  - When no `--env` param has been specified, then werf will use `[[ project ]]` template for the helm release and the same `[[ project ]]` template for the kubernetes namespace generation.

### New werf-render command

- [`werf render`]({{ "/reference/cli/werf_render.html" | true_relative_url }}) will build and publish needed images into the container registry automatically.
- This command allows reproducing exact templates in the same way as it will be rendered in [`werf converge`]({{ "/reference/cli/werf_converge.html" | true_relative_url }}) command.
- [`werf render`]({{ "/reference/cli/werf_render.html" | true_relative_url }}) work in [giterminism mode](#giterminism).

### Helm

#### Helm 3

- Helm 3 is default and the only choice when deploying with werf v1.2.
- An already existing helm 2 release will be migrated to helm 3 automatically in the `werf converge` command given that helm 2 release has the same name as newly deployed helm 3 release.
  - **CAUTION** There is no legal way back once helm 2 release has been migrated to helm 3.
  - werf-converge command will check that project helm charts are correctly rendered before running the process of migration.

#### Configuration

- `.Values.werf.image.IMAGE_NAME` instead of `werf_image` template.
- Removed `werf_container_env` helm template function.
- [Giterministic loading](#giterminism) of all chart files (including subcharts).
  - This is only true for top-level commands like `werf converge` or `werf render`.
  - Low-level helm commands `werf helm *` still load chart files from the local filesystem.
- Environment passed with `--env` param is available at the `.Values.werf.env`.
- Target namespace is available at the `.Values.werf.namespace`.
- Fix usage of the `.helm/Chart.yaml`. This makes `.helm/Chart.yaml` optional, but werf will use it when it exists.
  - Take `.helm/Chart.yaml` from the repository if exists.
  - Override `metadata.name` field with werf project name from the `werf.yaml`.
  - Set `metadata.version = 1.0.0` if not set.
- Added `.Values.werf.version` service value with werf cli util version.
- Support setting initial number of replicas when HPA is active for Deployment and other resources kinds. Set `"werf.io/replicas-on-creation": NUM` annotation, do not set `spec.replicas` field in templates in such case explicitly.
  - `spec.replicas` will override `werf.io/replicas-on-creation`.
  - This annotation is especially useful when HPA is active, [here is the reason]({{ "/reference/deploy_annotations.html#replicas-on-creation" | true_relative_url }}).

#### Working with subcharts

- Lock file with dependencies should be committed `.helm/Chart.lock`.
- werf will automatically download dependencies charts using this lock file.
- Typically `.helm/charts` directory should be added to `.gitignore` file.
- werf allows storing chart dependencies directly in the `.helm/charts` though.
  - These charts will override charts automatically downloaded by the `.helm/Chart.lock` file

#### Deleted werf-helm-deploy-chart command

Use `werf helm install/upgrade` native helm commands instead of `werf helm deploy-chart`.

#### Better integration with werf-helm-* commands

- `werf helm template` command could be used to render chart templates even if custom werf templates function are used (like `werf_secret_file`).
- Secret files are fully supported by `werf helm *` commands.
- Custom annotations and labels are fully supported by `werf helm *` commands (`--add-annotation` and `--add-label` options).

#### Changed service values format

Removed `.Values.global.werf.image` section, use `.Values.werf.image` instead.

#### Deploy process

- Wait until all release resources terminated before exiting in `werf dismiss` and `werf helm uninstall` commands.

### werf.yaml

- Rename `fromImageArtifact` to `fromArtifact`
  - **CAUTION** `fromImageArtifact/fromArtifact` has also been deprecated and will be removed in the v1.3, it is recommended to use `image` and `fromImage` in such case.
- Remove the `herebyIAdmitThatFromLatestMightBreakReproducibility` and `herebyIAdmitThatBranchMightBreakReproducibility` directives. Now when using the directives `fromLatest` and `git.branch` you must loosen the [giterminism]({{ "/usage/project_configuration/giterminism.html" | true_relative_url }}) rules with the corresponding directives in [werf-giterminism.yaml]({{ "/reference/werf_giterminism_yaml.html" | true_relative_url }}).
- Remove `--helm-chart-dir` option, define helm chart dir in the `werf.yaml`:

   ```yaml
   configVersion: 1
   deploy:
     helmChartDir: .helm
   ```

- Move `--allow-git-shallow-clone`, `--git-unshallow` and `--git-history-synchronization` options into the `werf.yaml`, added new `gitWorktree` meta section]({{ "/reference/werf_yaml.html#meta-section-gitWorktree" | true_relative_url }}). Disabled generation of old options in the `werf ci-env` command. Always unshallow git clone by default.
- Use sprig v3 instead of v2: [http://masterminds.github.io/sprig/](http://masterminds.github.io/sprig/).
- New documentation page describing [`werf.yaml` template engine]({{ "/reference/werf_yaml_template_engine.html" | true_relative_url }}).
- Fix naming for the user template. The template name is the path relative to the templates directory. {% raw %}{{ include ".werf/templates/1.tmpl" . }} => {{ include "templates/1.tmpl" . }}{% endraw %}.
- Nameless image declaration `image: ~` is deprecated, it is better to always name your images. Will be removed in v1.3.
- `env "ENVIRONMENT_VARIABLE_NAME"` function requires `ENVIRONMENT_VARIABLE_NAME` variable to exists (it may contain empty value, but should be explicitly defined).
- The new `required` function gives developers the ability to declare a value entry as required for config rendering. If the value is empty, the config will not render and will return an error message supplied by the developer.

  {% raw %}
   ```
   {{ required "A valid <anything> value required!" <anything> }}
   ```
  {% endraw %}

### Tagging strategy

- Only content-based.
- Export of built images into arbitrary container registry coming soon.
- Forcing custom built tags coming soon.

### Cleanup

- Git history based by default.
  - Previously this mode was enabled with `--git-history-based-cleanup-v1.2` option.
    - Option has been removed.
- Removed tag strategy based algorithms completely.
- `--keep-stages-built-within-last-n-hours` option to keep images that were built within period (2 hours by default).
- Git-history based cleanup improvement:
  - The keep policy `imagesPerReference.last` counts several stages that were built on the same commit as one.
  - Thus, the policy may cover more tags than expected, and this behavior is correct.

### Caching of stapel artifact/image imports by files checksum

- Suppose we have an artifact which builds some files.
- This artifact has a stage-dependency in the `werf.yaml` to rebuild these files when some source files has been changed.
- When source files has been changed werf rebuilds this artifact.
  - Because stage-dependency has changed, last artifact stage will also have a different digest.
  - But after artifact rebuild factual checksum of built files not changed — the same files.
    - In the v1.1 werf would still perform import of these files into the target image.
      - As a consequence target image will be unnecessarily rebuilt and redeployed.
    - In the v1.2 werf would check checksum of imported files and perform reimport only when checksum changed.

### Primary and secondary images storage support

- Automatically upload locally built images into the specified `--repo`.
- Use stages from read-only secondary images storage specified by `--secondary-repo` options (can be specified multiple times).
  - Suitable stages from the `--secondary-repo` will be copied into the primary `--repo`.

### Built images report format changes

`werf converge`, `werf build`, `werf run`, `werf bundle publish` and `werf render` commands has `--save-build-report` and `--build-report-path` options. `--save-build-report` option enables generation of built images report in the following format:

```shell
$ werf build --save-build-report

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

```shell
$ werf build --save-build-report --build-report-path=.werf-build-report.env

WERF_RESULT_DOCKER_IMAGE_NAME=quickstart-application:32e88a6a19a425c9254374ee2899b365876de31ac7d6857b523696a1-1613371915843
WERF_WORKER_DOCKER_IMAGE_NAME=quickstart-application:1b16118e7d5c67aa3c61fc0f8d49b3eccf8f72810f01c33a40290418-1613371916044
WERF_VOTE_DOCKER_IMAGE_NAME=quickstart-application:45f03bdd90c844eb2e61e7e01dae491588d2bdadbd195881b25be9b0-1613371915351
```

### Support multiple werf.yaml in the single git repo

User may put werf.yaml (and .helm) into any subdirectory of the project git repo.

- All relative paths specified in the werf.yaml will be calculated relatively to the werf process cwd or `--dir` param.
- Typically, user will run werf from the subdirectory where werf.yaml reside.
- There is also `--config` option to pass custom werf.yaml config, all relative paths will also be calculated relatively to the werf process cwd or `--dir` param.
- Added `--git-work-tree` param (or `WERF_GIT_WORK_TREE` variable) to specify a directory that contains `.git` in the case when autodetect failed, or we want to use specific work tree.
  - For example when running werf from the submodule of the project we may want to use root repo worktree instead of submodule's work tree.
    - Built stages will be linked to the commits of the root repo in such case.

### Other and internals

- Added git-archives and git-patches cache in the `~/.werf/local_cache/` along with git-worktrees. Do not create patches and archives in `/tmp` (or `--tmp-dir`).
-  Remove `stages_and_images` read-only lock during build process
  - `stages_and_images` read only lock is preventive measure against running cleanup and build commands at the same time;
  - This lock creates unnecessary load on the synchronization server, because this lock working all the time the build process is active for each build process.
  - It is mostly safe to just omit this lock completely.
- Also removed unused `image` lock from the lock manager (legacy from v1.1).
- Also renamed kubernetes-related stage and stage-cache locks to include project name (just for more correctness, incompatible with v1.1 change).

### Documentation rework

- Split all documentation into sections by the level of the understanding:
  - landing page;
  - introduction;
  - quickstart;
  - guides;
  - reference;
  - advanced;
  - internals.
- New guides section.
