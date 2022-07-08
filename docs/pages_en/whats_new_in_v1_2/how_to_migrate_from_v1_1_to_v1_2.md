---
title: How to migrate from v1.1 to v1.2
permalink: whats_new_in_v1_2/how_to_migrate_from_v1_1_to_v1_2.html
description: How to migrate your application from v1.1 to v1.2
---

Simply follow this guide to migrate your project from v1.1 to v1.2.

**IMPORTANT** There are v1.1 configurations that heavily use **`mount from build_dir`** and **`mount fromPath`** directives for such cases like caching of external dependencies (`Gemfile`, or `package.json`, or `go.mod`, etc.). While such `mount` directives still supported by the werf v1.2, it is not recommended for usage and is disabled by default by the [giterminism mode]({{ "/whats_new_in_v1_2/changelog.html#giterminism" | true_relative_url }}) (except `mount from tmp_dir` — this has not been forbidden) because it may lead to harmful, unreliable, host-dependent and nondeterministic builds. For such configurations there will be added a new feature to store such cache in the docker images, this feature is planned to be added into the werf v1.2, but it is [**not implemented yet**](https://github.com/werf/werf/issues/3318). So if your configuration uses such mounts, then there are several options:
 1. Keep using v1.1 version until this feature is [added into the werf v1.2](https://github.com/werf/werf/issues/3318).
 2. Completely remove `mounts` usage and wait until this feature is [added into the werf v1.2](https://github.com/werf/werf/issues/3318).
 3. Describe and enable your `mounts` in werf v1.2 using the [`werf-giterminism.yaml`](https://werf.io/documentation/reference/werf_giterminism_yaml.html) (**not recommended**).

## 1. Use content based tagging

[Content based tagging strategy]({{ "/whats_new_in_v1_2/changelog.html#tagging-strategy" | true_relative_url }}) is the default strategy and only choice to be used internally by the werf deploy process.

 - Remove `--tagging-strategy ...` param of `werf ci-env` command.
 - Remove `--tag-custom`, `--tag-git-tag`, `--tag-git-branch`, `--tag-by-stages-signature` params.

 In the case when you need a certain docker tag for a built image to exist in the container registry to be used outside the werf, then use `--report-path` and `--report-format` options like follows:
 - `werf build/converge --report-path=images-report.json --repo REPO`;
 - `docker pull $(cat images-report.json | jq -r .Images.IMAGE_NAME_FROM_WERF_YAML.DockerImageName)`;
 - `docker tag $(cat images-report.json | jq -r .Images.IMAGE_NAME_FROM_WERF_YAML.DockerImageName) REPO:mytag`;
 - `docker push REPO:mytag`.

## 2. Use `converge` command instead of `build-and-publish` command

Use `werf converge` command instead of `werf deploy` to build and publish images, and deploy into kubernetes. All params are the same as for `werf deploy`.

Commands changes (for reference):
 - `werf build-and-publish` command has been removed, there is only `werf build` command, usage of which is optional:
 - `werf converge` will build and publish needed images by itself, if not built already.
 - You may use `werf build` command in `prebuild` CI/CD pipeline stage instead of `werf build-and-publish` command.

## 3. Use the `--repo` parameter instead of the `--stages-storage` and `--images-repo` parameters

The `werf converge` command stores both the collected images and their stages in the container registry. To specify the storage to be used for storing images and stages, you now need to specify one parameter `--repo` instead of the two previously used in v1.1 `--stages-storage` and `--images-repo`. You can read more about this parameter [in changelog](https://werf.io/documentation/v1.2/whats_new_in_v1_2/changelog.html#always-store-image-stages-in-the-container-registry).

Similarly with the environment variables used earlier: instead of `WERF_STAGES_STORAGE` and `WERF_IMAGES_REPO`, one variable `WERF_REPO` is now used.

## 4. Change Helm templates

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
 - Use `"werf.io/replicas-on-creation": "NUM"` annotation instead of `"werf.io/set-replicas-only-on-creation": "true"`.
   > **IMPORTANT!** Specify `"NUM"` as **string** instead of number `NUM`, annotation [will be ignored otherwise]({{ "/reference/deploy_annotations.html#replicas-on-creation" | true_relative_url }}).

   > **IMPORTANT!** Remove `spec.replicas` field, more info [in the changelog]({{ "/whats_new_in_v1_2/changelog.html#configuration" | true_relative_url }}).

## 5. Use `.helm/Chart.lock` for subcharts

 Due to [giterminism]({{ "/whats_new_in_v1_2/changelog.html#giterminism" | true_relative_url  }}) werf does not allow uncommitted `.helm/charts` dir. To use subcharts specify dependencies in the `.helm/Chart.yaml` like that:

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

 werf will automatically download subcharts into the cache and load subchart files in `werf converge` command (and other toplevel commands which require helm chart). More info [in the docs]({{ "advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}).

## 6. Cleanup by Git history

Remove `--git-history-based-cleanup-v1.2` option for a cleanup. werf always uses git-history cleanup in the v1.2. More info [in the changelog]({{ "/whats_new_in_v1_2/changelog.html#cleanup" | true_relative_url }}) and [in the cleanup article]({{ "/advanced/cleanup.html" | true_relative_url }}).

## 7. Define environment variables in `werf-giterminism.yaml`

> **NOTE!** It is not recommended to use environment variables in the `werf.yaml`, more info [in the article]({{ "/advanced/giterminism.html" | true_relative_url }}).

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

## 8. Adjust `werf.yaml` configuration

 1. Change relative include paths from {% raw %}{{ include ".werf/templates/1.tmpl" . }}{% endraw %} to {% raw %}{{ include "templates/1.tmpl" . }}{% endraw %}.
 1. Rename `fromImageArtifact` to `fromArtifact`.
    > **NOTICE!** It is not required, but better to change `artifact` and `fromArtifact` to `image` and `fromImage` in such case. See [deprecation note in the changelog]({{ "/whats_new_in_v1_2/changelog.html#werfyaml" | true_relative_url }}).

## 9. Define custom `.helm` chart dir in the `werf.yaml`

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

## 10. Migrate to helm 3

 werf v1.2 [performs migration of existing Helm 2 release to helm 3 automatically]({{ "/whats_new_in_v1_2/changelog.html#helm-3" | true_relative_url }}).
 > Helm 2 release should have the same name as newly used helm 3 release.

 Before migrating werf will try to render and validate current `.helm/templates` and continue migration only when render succeeded.

 > **IMPORTANT!** Once project has been migrated to helm 3 there is no legal way back to the werf v1.1 and helm 2.
