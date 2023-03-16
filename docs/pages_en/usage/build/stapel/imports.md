---
title: Importing from images and artifacts
permalink: usage/build/stapel/imports.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: import
---

The size of the final image can grow dramatically due to the assembly tools and source files eating up space. These files are generally not needed in the final image.
To avoid this, the Docker community suggests installing tools, building, and removing irrelevant files in one step:

```Dockerfile
RUN “download-source && cmd && cmd2 && remove-source”
```

> You can do the same in werf — just specify the relevant instructions for some _user stage_. Below is an example of specifying the _shell assembly instructions_ for the _install stage_ (you can do so for the _ansible_ builder as well):
```yaml
shell:
  install:
  - "download-source"
  - "cmd"
  - "cmd2"
  - "remove-source"
```

However, this method does not support caching. Thus, a build toolkit will be installed all over again.

Another way is to use a multi-stage build, a feature supported in Docker since version 17.05:

```Dockerfile
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

The point of this approach is to describe several auxiliary images and selectively copy artifacts from one image to another, leaving all the unnecessary data outside the final image.

werf offers the same approach, but using _images_ and _artifacts_.

> Why doesn't werf use multi-stage assembly?
* Historically, _imports_ came much earlier than the Docker multi-stage mechanism, and
* werf allows for greater flexibility when working with auxiliary images.

Importing _resources_ from the _images_ and _artifacts_ must be described in the `import` directive in the _destination image_ in the _image_ or _artifact_ config section. `import` is an array of records, where each record must contain the following:

- `image: <image name>` or `artifact: <artifact name>`: _source image_; the name of the image copy files from.
- `stage: <stage name>`: _source image stage_; the stage of the _source_image_ to copy files from.
- `add: <absolute path>`: _source path_; the absolute path to the file or directory in the _source image_ to copy from.
- `to: <absolute path>`: _destination path_; the absolute path in the _destination image_. If absent, the _destination path_ defaults to the  _source path_ (as specified by the `add` directive).
- `before: <install || setup>` or `after: <install || setup>`: _destination image stage_; the stage to import files. Currently, only _install_ and _setup_ stages are supported.

An example of the `import` directive:

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

As with the _git mappings_ configuration, include and exclude file and directory masks are supported (`include_paths: []` and `exclude_paths: []`, respectively). Masks must be specified relative to the source path (as in the `add` parameter).
You can also specify an owner and a group for the imported resources, `owner: <owner>` and `group: <group>`.
This behavior is similar to the one used when adding code from Git repositories, and you can read more about it in the [git directive section]({{ "usage/build/stapel/git.html" | true_relative_url }}).

> Note that the path of imported resources and the path specified in _git mappings_ must not overlap.

Refer to [this section]({{ "usage/build/stapel/imports.html#what-is-an-artifact" | true_relative_url }}) to learn more about _using artifacts_.

## What is an artifact?

An ***artifact*** is a special image used by other _images_ and _artifacts_ to isolate the build process and build tools resources (environments, software, data) from the application image build process. Examples of such resources include software or data used to build the image but not needed to run the application, etc.

Using artifacts, you can build an unlimited number of components. They also help you to solve the following problems:

- If an application consists of a set of components, each with their own dependencies, you usually have to rebuild all the components every time. Artifacts allow you to assemble only the components that you really need.
- Assembling components in different environments.

Refer to [import directive]({{ "usage/build/stapel/imports.html" | true_relative_url }}) to learn more about importing _resources_ from the _artifacts_.

### Configuration

The configuration of an _artifact_ is similar to that of an _image_. Each _artifact_ must be described in irs own artifact config section.

The instructions related to the _from stage_, namely, the instruction to set the [_base image_]({{ "usage/build/stapel/base.html" | true_relative_url }}), [mounts]({{ "usage/build/stapel/mounts.html" | true_relative_url }}), and [imports]({{ "usage/build/stapel/imports.html" | true_relative_url }}) are exactly the same as those for images.

The _docker_ _ _instructions stage_ and the [related instructions]({{ "usage/build/stapel/dockerfile.html" | true_relative_url }}) are not supported for _artifacts_. An _artifact_ is an assembly tool, and and the only thing it provides is data.

The remaining _stages_ and instructions for defining artifacts are discussed in detail below.

#### Naming

<div class="summary" markdown="1">
```yaml
artifact: string
```
</div>

The _artifact images_ is declared with the `artifact` directive: `artifact: string`. In contrast to naming _images_, when naming artifacts, you do not have to observe the Docker naming convention, since artifacts used purely internally.

```yaml
artifact: "application assets"
```

#### Adding source code from Git repositories

<div class="summary">

<a class="google-drawings" href="{{ "images/configuration/stapel_artifact2.svg" | true_relative_url }}" data-featherlight="image">
  <img src="{{ "images/configuration/stapel_artifact2.svg" | true_relative_url }}">
</a>

</div>

Unlike regular _images_, the _artifact stage conveyor_ has no _gitCache_ and _gitLatestPatch_ stages.

> In werf, there is an optional dependency on changes to Git repositories for _artifacts_. So, by default, werf ignores any changes to the Git repositories by caching the _artifact image_ after the first build. But you can define file and directory dependencies that will trigger an artifact image rebuild when changed.

Refer to the [documentation]({{ "usage/build/stapel/git.html" | true_relative_url }}) to learn more about working with _git repositories_.

#### Running assembly instructions

<div class="summary">

<a class="google-drawings" href="{{ "images/configuration/stapel_artifact3.svg" | true_relative_url }}" data-featherlight="image">
  <img src="{{ "images/configuration/stapel_artifact3.svg" | true_relative_url }}">
</a>

</div>

Artifacts share the same set of directives and _user stages_ as regular images: _beforeInstall_, _install_, _beforeSetup_, and _setup_.

If you don't specify any file dependencies in the `stageDependencies` directive in the Git section for a _user stage_, the image will be cached after the first build and will not be reassembled as long as the corresponding _stage_ exists in the _stages storage_.

> To rebuild the artifact whenever there are any changes in the related Git repository, specify _stageDependency_ `**/*` for any _user stage_. Below is an example of setting it for the _install stage_:

```yaml
git:
- to: /
  stageDependencies:
    install: "**/*"
```

Refer to the [documentation]({{ "usage/build/stapel/instructions.html" | true_relative_url }}) to learn more about working with _assembly instructions_.

## Using artifacts

Unlike the [*stapel image*]({{ "usage/build/stapel/instructions.html" | true_relative_url }}), the *stapel artifact* has no _git latest patch_ stage.

This is intentional, because the _git latest patch_ stage usually runs on every commit, applying the changes made to the files. However, it makes sense to use the *Stapel artifact* as a highly cacheable image that is updated infrequently (e.g., when some special files are changed).

Suppose you want to import data from Git into a *Stapel artifact*, and only rebuild the assets when the files that affect the assembly of the assets change. This means assets should not be rebuilt if other files in Git are changed.

Of course, sometimes you want to include changes made to any files in the Git repository in your *stapel artifact* (for example, if a Go app is being built in the artifact). In that case, you have to specify a dependency relative to the stage to build if there are changes in Git using `git.stageDependencies` and `*` as a pattern:

```
git:
- add: /
  to: /app
  stageDependencies:
    setup:
    - "*"
```

In this case, any changes to the files in the Git repository will result in a rebuild of the artifact image and all *stapel images* into which that artifact is imported.

**NOTE:** If you use any files in both the *stapel artifact* and the [*stapel image*]({{"usage/build/stapel/instructions.html" | true_relative_url }}) builds, the proper way is to use the `git.add` directive in each image, that is, multiple times, when adding Git files. We do not recommend adding files to the artifact and then importing them into the image using the `import` directive.
