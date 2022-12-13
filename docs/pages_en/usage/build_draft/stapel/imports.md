---
title: Importing from images and artifacts
permalink: usage/build_draft/stapel/imports.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: import
---

The size of the image can be increased several times due to the assembly tools and source files, while the user does not need them.
To solve such problems, Docker community suggests doing the installation of tools, the assembly, and removal in one step.

```
RUN “download-source && cmd && cmd2 && remove-source”
```

> To obtain a similar effect when using werf, it is sufficient describing the instructions in one _user stage_. For example, _shell assembly instructions_ for the _install stage_ (similarly for _ansible_):
```yaml
shell:
  install:
  - "download-source"
  - "cmd"
  - "cmd2"
  - "remove-source"
```

However, with this method, it is not possible using a cache, so toolkit installation runs each time.

Another solution is using multi-stage builds, which are supported starting with Docker 17.05.

```
FROM node:latest AS storefront
WORKDIR /app
COPY react-app .
RUN npm install
RUN npm run build

FROM maven:latest AS appserver
WORKDIR /app
COPY . .
RUN mvn package

FROM java:8-jdk-alpine
COPY --from=storefront /app/react-app/build/ /static
COPY --from=appserver /app/target/AtSea-0.0.1-SNAPSHOT.jar /app/AtSea.jar
```

The meaning of such an approach is as follows, describe several auxiliary images and selectively copy artifacts from one image to another leaving behind everything you do not want in the result image.

We suggest the same, but using [_images_]({{ "reference/werf_yaml.html#image-section" | true_relative_url }}) and [_artifacts_]({{ "reference/werf_yaml.html#image-section" | true_relative_url }}).

> Why is werf not using multi-stage?
* Historically, _imports_ appeared much earlier than Docker multi-stage, and
* werf gives more flexibility working with auxiliary images

Importing _resources_ from _images_ and _artifacts_ should be described in `import` directive in _destination image_ config section ([_image_]({{ "reference/werf_yaml.html#image-section" | true_relative_url }}) or [_artifact_]({{ "reference/werf_yaml.html#image-section" | true_relative_url }}). `import` is an array of records. Each record should contain the following:

- `image: <image name>` or `artifact: <artifact name>`: _source image_, image name from which you want to copy files.
- `stage: <stage name>`: _source image stage_, particular stage of _source_image_ from which you want to copy files.
- `add: <absolute path>`: _source path_, absolute file or folder path in _source image_ for copying.
- `to: <absolute path>`: _destination path_, absolute path in _destination image_. In case of absence, _destination path_ equals _source path_ (from `add` directive).
- `before: <install || setup>` or `after: <install || setup>`: _destination image stage_, stage for importing files. At present, only _install_ and _setup_ stages are supported.

```yaml
import:
- artifact: application-assets
  add: /app/public/assets
  to: /var/www/site/assets
  after: install
- image: frontend
  add: /app/assets
  after: setup
```

As in the case of adding _git mappings_, masks are supported for including, `include_paths: []`, and excluding files, `exclude_paths: []`, from the specified path.
You can also define the rights for the imported resources, `owner: <owner>` and `group: <group>`.
Read more about these in the [git directive article]({{ "usage/build_draft/stapel/git.html" | true_relative_url }}).

> Import paths and _git mappings_ must not overlap with each other

Information about _using artifacts_ available in [separate article]({{ "usage/build_draft/stapel/imports.html#what-is-an-artifact" | true_relative_url }}).

## What is an artifact?

***Artifact*** is a special image that is used by _images_ and _artifacts_ to isolate the build process and build tools resources (environments, software, data).

Using artifacts, you can independently assemble an unlimited number of components, and also solving the following problems:

- The application can consist of a set of components, and each has its dependencies. With a standard assembly, you should rebuild all every time, but you want to assemble each one on-demand.
- Components need to be assembled in other environments.

Importing _resources_ from _artifacts_ are described in [import directive]({{ "usage/build/building_images_with_stapel/import_directive.html" | true_relative_url }}).

### Configuration

The configuration of the _artifact_ is not much different from the configuration of _image_. Each _artifact_ should be described in a separate artifact config section.

The instructions associated with the _from stage_, namely the [_base image_]({{ "usage/build_draft/stapel/base.html" | true_relative_url }}) and [mounts]({{ "usage/build/building_images_with_stapel/mount_directive.html" | true_relative_url }}), and also [imports]({{ "usage/build/building_images_with_stapel/import_directive.html" | true_relative_url }}) remain unchanged.

The _docker_instructions stage_ and the [corresponding instructions]({{ "usage/build/building_images_with_stapel/docker_directive.html" | true_relative_url }}) are not supported for the _artifact_. An _artifact_ is an assembly tool and only the data stored in it is required.

The remaining _stages_ and instructions are considered further separately.

#### Naming

<div class="summary" markdown="1">
```yaml
artifact: string
```
</div>

_Artifact images_ are declared with `artifact` directive: `artifact: string`. Unlike the [name of the _image_]({{ "reference/werf_yaml.html#image-section" | true_relative_url }}), the artifact has no limitations associated with docker naming convention, as used only internal.

```yaml
artifact: "application assets"
```

#### Adding source code from git repositories

<div class="summary">

<a class="google-drawings" href="{{ "images/configuration/stapel_artifact2.png" | true_relative_url }}" data-featherlight="image">
  <img src="{{ "images/configuration/stapel_artifact2_preview.png" | true_relative_url }}">
</a>

</div>

Unlike with _image_, _artifact stage conveyor_ has no _gitCache_ and _gitLatestPatch_ stages.

> werf implements optional dependence on changes in git repositories for _artifacts_. Thus, by default werf ignores them and _artifact image_ is cached after the first assembly, but you can specify any dependencies for assembly instructions

Read about working with _git repositories_ in the corresponding [article]({{ "usage/build_draft/stapel/git.html" | true_relative_url }}).

#### Running assembly instructions

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

Read about working with _assembly instructions_ in the corresponding [article]({{ "usage/build/building_images_with_stapel/assembly_instructions.html" | true_relative_url }}).

## Using artifacts

Unlike [*stapel image*]({{ "usage/build/building_images_with_stapel/assembly_instructions.html" | true_relative_url }}), *stapel artifact* does not have a git latest patch stage.

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

**NOTE** User should employ multiple separate `git.add` directive invocations in every [*stapel image*]({{ "usage/build/building_images_with_stapel/assembly_instructions.html" | true_relative_url }}) and *stapel artifact* that needs git files — it is an optimal way to add git files into any image. Adding git files to artifact and then importing it into image using `import` directive is not recommended.
