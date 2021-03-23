---
title: Artifacts
permalink: advanced/building_images_with_stapel/artifacts.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: artifact
---

## What is an artifact?

***Artifact*** is a special image that is used by _images_ and _artifacts_ to isolate the build process and build tools resources (environments, software, data).

Using artifacts, you can independently assemble an unlimited number of components, and also solving the following problems:

- The application can consist of a set of components, and each has its dependencies. With a standard assembly, you should rebuild all every time, but you want to assemble each one on-demand.
- Components need to be assembled in other environments.

Importing _resources_ from _artifacts_ are described in [import directive]({{ "advanced/building_images_with_stapel/import_directive.html" | true_relative_url }}).

## Configuration

The configuration of the _artifact_ is not much different from the configuration of _image_. Each _artifact_ should be described in a separate artifact config section.

The instructions associated with the _from stage_, namely the [_base image_]({{ "advanced/building_images_with_stapel/base_image.html" | true_relative_url }}) and [mounts]({{ "advanced/building_images_with_stapel/mount_directive.html" | true_relative_url }}), and also [imports]({{ "advanced/building_images_with_stapel/import_directive.html" | true_relative_url }}) remain unchanged.

The _docker_instructions stage_ and the [corresponding instructions]({{ "advanced/building_images_with_stapel/docker_directive.html" | true_relative_url }}) are not supported for the _artifact_. An _artifact_ is an assembly tool and only the data stored in it is required.

The remaining _stages_ and instructions are considered further separately.

### Naming

<div class="summary" markdown="1">
```yaml
artifact: string
```
</div>

_Artifact images_ are declared with `artifact` directive: `artifact: string`. Unlike the [name of the _image_]({{ "reference/werf_yaml.html#image-section" | true_relative_url }}), the artifact has no limitations associated with docker naming convention, as used only internal.

```yaml
artifact: "application assets"
```

### Adding source code from git repositories

<div class="summary">

<a class="google-drawings" href="{{ "images/configuration/stapel_artifact2.png" | true_relative_url }}" data-featherlight="image">
  <img src="{{ "images/configuration/stapel_artifact2_preview.png" | true_relative_url }}">
</a>

</div>

Unlike with _image_, _artifact stage conveyor_ has no _gitCache_ and _gitLatestPatch_ stages.

> werf implements optional dependence on changes in git repositories for _artifacts_. Thus, by default werf ignores them and _artifact image_ is cached after the first assembly, but you can specify any dependencies for assembly instructions

Read about working with _git repositories_ in the corresponding [article]({{ "advanced/building_images_with_stapel/git_directive.html" | true_relative_url }}).

### Running assembly instructions

<div class="summary">

<a class="google-drawings" href="{{ "images/configuration/stapel_artifact3.png" | true_relative_url }}" data-featherlight="image">
  <img src="{{ "images/configuration/stapel_artifact3_preview.png" | true_relative_url }}">
</a>

</div>

Directives and _user stages_ remain unchanged: _beforeInstall_, _install_, _beforeSetup_ and _setup_.

If there are no dependencies on files specified in git `stageDependencies` directive for _user stages_, the image is cached after the first build and will no longer be reassembled while the related _stages_ exist in _storage_.

> If the artifact should be rebuilt on any change in the related git repository, you should specify the _stageDependency_ `**/*` for any _user stage_, e.g., for _install stage_:
```yaml
git:
- to: /
  stageDependencies:
    install: "**/*"
```

Read about working with _assembly instructions_ in the corresponding [article]({{ "advanced/building_images_with_stapel/assembly_instructions.html" | true_relative_url }}).

## Using artifacts

Unlike [*stapel image*]({{ "advanced/building_images_with_stapel/assembly_instructions.html" | true_relative_url }}), *stapel artifact* does not have a git latest patch stage.

The git latest patch stage is supposed to be updated on every commit, which brings new changes to files. *Stapel artifact* though is recommended to be used as a deeply cached image, which will be updated in rare cases, when some special files changed.

For example: import git into *stapel artifact* and rebuild assets in this artifact only when dependent assets files in git has changes. For every other change in git where non-dependent files has been changed assets will not be rebuilt.

However in the case when there is a need to bring changes of any git files into *stapel artifact* (to build Go application for example) user should define `git.stageDependencies` of some stage that needs these files explicitly as `*` pattern:

```
git:
- add: /
  to: /app
  stageDependencies:
    setup:
    - "*"
```

In this case every change in git files will result in artifact rebuild, all *stapel images* that import this artifact will also be rebuilt.

**NOTE** User should employ multiple separate `git.add` directive invocations in every [*stapel image*]({{ "advanced/building_images_with_stapel/assembly_instructions.html" | true_relative_url }}) and *stapel artifact* that needs git files — it is an optimal way to add git files into any image. Adding git files to artifact and then importing it into image using `import` directive is not recommended.
