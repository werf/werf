---
title: Using artifacts
sidebar: reference
permalink: reference/build/artifact.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
summary: |
  <a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRD-K_z7KEoliEVT4GpTekCkeaFMbSPWZpZkyTDms4XLeJAWEnnj4EeAxsdwnU3OtSW_vuKxDaaFLgD/pub?w=1800&amp;h=850" data-featherlight="image">
  <img src="https://docs.google.com/drawings/d/e/2PACX-1vRD-K_z7KEoliEVT4GpTekCkeaFMbSPWZpZkyTDms4XLeJAWEnnj4EeAxsdwnU3OtSW_vuKxDaaFLgD/pub?w=640&amp;h=301">
  </a>
---

## What is an artifact?

***Artifact*** is a special image that is used by _images_ and _artifacts_ to isolate the build process and build tools resources (environments, software, data).

_Artifact_ cannot be [tagged like _image_]({{ site.baseurl }}/reference/registry/image_naming.html#werf-tag-procedure) and used as standalone application.

Using artifacts, you can independently assemble an unlimited number of components, and also solving the following problems:

- The application can consist of a set of components, and each has its dependencies. With a standard assembly, you should rebuild all every time, but you want to assemble each one on-demand.
- Components need to be assembled in other environments.

Importing _resources_ from _artifacts_ are described in [import directive]({{ site.baseurl }}/reference/build/import_directive.html) in _destination image_ config section ([_image_]({{ site.baseurl }}/reference/config.html#image-config-section) or [_artifact_]({{ site.baseurl }}/reference/config.html#artifact-config-section])).

## Configuration

The configuration of the _artifact_ is not much different from the configuration of _image_. Each _artifact_ should be described in a separate [artifact config section]({{ site.baseurl }}/reference/config.html#artifact-config-section).

The instructions associated with the _from stage_, namely the [_base image_]({{ site.baseurl }}/reference/build/base_image.html) and [mounts]({{ site.baseurl }}/reference/build/mount_directive.html), and also [imports]({{ site.baseurl }}/reference/build/import_directive.html) remain unchanged.

The _docker_instructions stage_ and the [corresponding instructions]({{ site.baseurl }}/reference/build/docker_directive.html) are not supported for the _artifact_. An _artifact_ is an assembly tool and only the data stored in it is required.

The remaining _stages_ and instructions are considered further separately.

### Naming

<div class="summary" markdown="1">
```yaml
artifact: <artifact name>
```
</div>

_Artifact images_ are declared with `artifact` directive: `artifact: <artifact name>`. Unlike the [naming of the _image_]({{ site.baseurl }}/reference/build/naming.html), the artifact has no limitations associated with docker naming convention, as used only internal.

```yaml
artifact: "application assets"
```

### Adding source code from git repositories

<div class="summary">

<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vQQiyUM9P3-_A6O5tLms_y1UCny9X6lxQSxtMtBalcyjcHhYV4hnPnISmTVY09c-ANOFqwHeOxYPs63/pub?w=2031&amp;h=144" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vQQiyUM9P3-_A6O5tLms_y1UCny9X6lxQSxtMtBalcyjcHhYV4hnPnISmTVY09c-ANOFqwHeOxYPs63/pub?w=1016&amp;h=72">
</a>

</div>

Unlike with _image_, _artifact stage conveyor_ has no _gitCache_ and _gitLatestPatch_ stages.

> Werf implements optional dependence on changes in git repositories for _artifacts_. Thus, by default werf ignores them and _artifact image_ is cached after the first assembly, but you can specify any dependencies for assembly instructions.

Read about working with _git repositories_ in the corresponding [article]({{ site.baseurl }}/reference/build/git_directive.html).

### Running assembly instructions

<div class="summary">

<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vTlpKbAr6wQCE4bSxVB5Kr6uxzbCGu_ncsviT2Ax6_qLL3zAVLWIsYElAi9_LMuVeFiDi1lo97HNvD2/pub?w=1428&h=649" data-featherlight="image">
      <img src="https://docs.google.com/drawings/d/e/2PACX-1vTlpKbAr6wQCE4bSxVB5Kr6uxzbCGu_ncsviT2Ax6_qLL3zAVLWIsYElAi9_LMuVeFiDi1lo97HNvD2/pub?w=426&h=216">
</a>

</div>

Directives and _user stages_ remain unchanged: _beforeInstall_, _install_, _beforeSetup_ and _setup_.

If there are no dependencies on files specified in git `stageDependencies` directive for _user stages_, the image is cached after the first build and will no longer be reassembled while the related _stages_ exist in _stages storage_.

> If the artifact should be rebuilt on any change in the related git repository, you should specify the _stageDependency_ `**/*` for any _user stage_, e.g., for _install stage_:
```yaml
git:
- to: /
  stageDependencies:
    install: "**/*"
```

Read about working with _assembly instructions_ in the corresponding [article]({{ site.baseurl }}/reference/build/assembly_instructions.html).

## All directives
```yaml
artifact: <artifact_name>
from: <image>
fromLatest: <bool>
fromCacheVersion: <version>
fromImage: <image_name>
fromImageArtifact: <artifact_name>
git:
# local git
- add: <absolute_path>
  to: <absolute_path>
  owner: <owner>
  group: <group>
  includePaths:
  - <relative_path or glob>
  excludePaths:
  - <relative_path or glob>
  stageDependencies:
    install:
    - <relative_path or glob>
    beforeSetup:
    - <relative_path or glob>
    setup:
    - <relative_path or glob>
# remote git
- url: <git_repo_url>
  branch: <branch_name>
  commit: <commit>
  tag: <tag>
  add: <absolute_path>
  to: <absolute_path>
  owner: <owner>
  group: <group>
  includePaths:
  - <relative_path or glob>
  excludePaths:
  - <relative_path or glob>
  stageDependencies:
    install:
    - <relative_path or glob>
    beforeSetup:
    - <relative_path or glob>
    setup:
    - <relative_path or glob>
shell:
  beforeInstall:
  - <cmd>
  install:
  - <cmd>
  beforeSetup:
  - <cmd>
  setup:
  - <cmd>
  cacheVersion: <version>
  beforeInstallCacheVersion: <version>
  installCacheVersion: <version>
  beforeSetupCacheVersion: <version>
  setupCacheVersion: <version>
ansible:
  beforeInstall:
  - <task>
  install:
  - <task>
  beforeSetup:
  - <task>
  setup:
  - <task>
  cacheVersion: <version>
  beforeInstallCacheVersion: <version>
  installCacheVersion: <version>
  beforeSetupCacheVersion: <version>
  setupCacheVersion: <version>
mount:
- from: build_dir
  to: <absolute_path>
- from: tmp_dir
  to: <absolute_path>
- fromPath: <absolute_path>
  to: <absolute_path>
import:
- artifact: <artifact name>
  image: <image name>
  before: <install || setup>
  after: <install || setup>
  add: <absolute path>
  to: <absolute path>
  owner: <owner>
  group: <group>
  includePaths:
  - <relative path or glob>
  excludePaths:
  - <relative path or glob>
asLayers: <bool>
```
