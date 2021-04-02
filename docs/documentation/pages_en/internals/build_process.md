---
title: Build process
permalink: internals/build_process.html
---

The build process in werf implies a [sequential assembly of stages]({{ "/internals/stages_and_storage.html#stage-conveyor" | true_relative_url }}) for images defined in [werf.yaml]({{ "/reference/werf_yaml.html" | true_relative_url }}).

While the [stage pipelines]({{ "/internals/stages_and_storage.html#stage-conveyor" | true_relative_url }}) for the Dockerfile image, Stapel image, and Stapel artifact are unique, each stage follows the same rules for [selecting](#selecting-stages), [saving](#saving-stages-to-the-storage), [caching & locking]({{ "/advanced/synchronization.html" | true_relative_url }}) in parallel runs.

## Building a stage of the Dockerfile image

werf creates a single [stage]({{ "/internals/stages_and_storage.html#stage-conveyor" | true_relative_url }}) called `dockerfile` to build a Dockerfile image.

Currently, werf uses the standard commands of the built-in Docker client (in a way similar to invoking the `docker build` command) as well as arguments defined in `werf.yaml` to build a stage. The cache created during the build is used as if it is created via a regular docker client (werf isn't involved in the process).

The distinctive feature of werf is that it uses the git repository (instead of the project directory) as a source of files for the build context. All files in the directory specified by the `context` directive  (by default, this is the project directory) are coming from the current commit in the project repository (you can learn more about giterminism in a [separate article]({{ "/advanced/giterminism.html#dockerfile-image" | true_relative_url }})).

Here is an example of a command to build a `dockerfile` stage:

```shell
docker build --file=Dockerfile - < ~/.werf/service/tmp/context/4b9d6bc2-a549-42f9-86b8-4032c146f888
```

Learn more about the `werf.yaml` build configuration file in the [corresponding section]({{ "reference/werf_yaml.html#dockerfile-builder" | true_relative_url }}).

## Building a stage of the Stapel image and Stapel artifact

During the build, the stage instructions are assumed to be run in a container based on the previously built stage or the [base image]({{ "/advanced/building_images_with_stapel/base_image.html#from-fromlatest" | true_relative_url }}). Hereinafter, such a container will be referred to as a **build container**.

Before starting the _build container_, werf prepares a set of instructions that depends on the stage type. It contains both werf service commands and [user commands]({{ "/advanced/building_images_with_stapel/assembly_instructions.html" | true_relative_url }}) specified in the `werf.yaml` configuration file. For example, service commands may include adding files, applying patches, running ansible tasks, and so on.

The Stapel builder uses its own set of tools and libraries and runs independently of the base image. When starting the _build container_, werf mounts a special service image `flant/werf-stapel` that contains all the necessary tools and libraries. You may find more info about the stapel image in the [article]({{ "/internals/development/stapel_image.html" | true_relative_url }}).

The [socket of the host's SSH agent]({{ "/internals/integration_with_ssh_agent.html" | true_relative_url }}) is forwarded to the _build container_. Also, werf can use [custom mounts]({{ "/advanced/building_images_with_stapel/mount_directive.html" | true_relative_url }}).

Note that during the building, werf ignores some parameters of the base image manifest, overwriting them with the following values:

- `--user=0:0`;
- `--workdir=/`;
- `--entrypoint=/.werf/stapel/embedded/bin/bash`.

The final sequence of parameters for initiating a build of any stage might look as follows:

```shell
docker run \
  --volume=/tmp/ssh-ln8yCMlFLZob/agent.17554:/.werf/tmp/ssh-auth-sock \
  --volumes-from=stapel_0.6.1 \
  --env=SSH_AUTH_SOCK=/.werf/tmp/ssh-auth-sock \
  --user=0:0 \
  --workdir=/ \
  --entrypoint=/.werf/stapel/embedded/bin/bash \
  sha256:d6e46aa2470df1d32034c6707c8041158b652f38d2a9ae3d7ad7e7532d22ebe0 \
  -ec eval $(echo c2V0IC14 | /.werf/stapel/embedded/bin/base64 --decode)
```

Learn more about the `werf.yaml` build configuration file in the [corresponding section]({{ "/reference/werf_yaml.html#stapel-builder" | true_relative_url }}).

### How the Stapel builder processes CMD and ENTRYPOINT

To build a stage, werf starts container with the `CMD` and `ENTRYPOINT` service parameters and then substitutes them with the values of the [base image]({{ "/advanced/building_images_with_stapel/base_image.html" | true_relative_url }}). If these values are not set in the base image, werf resets them as follows:

- `[]` for `CMD`;
- `[""]` for `ENTRYPOINT`.

Also, werf resets (uses special empty values) `ENTRYPOINT` of the base image if the `CMD` parameter is specified in the configuration (`docker.CMD`).

Otherwise, the behavior of werf is similar to that of [Docker](https://docs.docker.com/engine/reference/builder/#understand-how-cmd-and-entrypoint-interact).

## Selecting stages

The stage selection algorithm in werf includes the following steps:

1. The [stage digest]({{ "/internals/stages_and_storage.html#stage-digest" | true_relative_url }}) is calculated.
2. All the suitable stages are selected (multiple stages in the [storage]({{ "/internals/stages_and_storage.html#storage" | true_relative_url }}) can be associated with a single digest).
3. The stage with the oldest `TIMESTAMP_MILLISEC` is selected (refer to [this article]({{ "/internals/stages_and_storage.html#stage-naming" | true_relative_url }}) to learn more about the naming of stages).

### Additional steps for Stapel images and artifacts

The check for the ancestry of git commits is added to the basic algorithm. After the step **2**, additional filtering based on the git history is performed. Only stages associated with commits that are the ancestors of the current commit are selected if the current stage is associated with git (the git-archive stage, the custom stage with git patches, or the git-latest-patch stage). Thus, the commits in neighboring branches are discarded.

There can be a situation when several built images have the same digest. Moreover, stages related to different git branches can have the same digest. However, werf is guaranteed to prevent cache reuse between unrelated branches. The cache in different branches can only be reused if it is associated with a commit that serves as a basis for both branches.

## Saving stages to the storage

Multiple werf processes (on a single or several hosts) can initiate a parallel build of the same stage at the same time because that stage is not yet in the repository.

werf uses an optimistic locking while saving a newly built image to the storage. When the build of the new image is complete, werf blocks the storage for any operations with the target digest:

- If no suitable image emerges in the repository while the building process continues, werf saves the newly built image using the `TIMESTAMP_MILLISEC` ID that is guaranteed to be unique.
- If a suitable image is saved to the repository while the building process continues, werf discards the newly built image and uses the image in the repository instead.

In other words, the first process that finishes the build (the fastest one) is granted a chance to save the newly built stage to the storage. Thus, the slower build processes don't block the faster ones in a parallel and distributed environment.

When selecting and saving new stages to the storage, werf uses a [lock manager]({{ "/advanced/synchronization.html" | true_relative_url }}) to coordinate the work of several werf processes.

## Parallel build

The parallel assembly in werf is managed by `--parallel` (`-p`) and `--parallel-tasks-limit` parameters. By default, it is enabled and limited to build five images in parallel.

After building a tree of image dependencies, werf splits the assembly process into stages. Each stage contains a set of independent images that can be built in parallel.

```shell
┌ Concurrent builds plan (no more than 5 images at the same time)
│ Set #0:
│ - ⛵ image common-base
│ - 🛸 artifact jq
│ - 🛸 artifact libjq
│ - 🛸 artifact kcov
│
│ Set #1:
│ - ⛵ image base-for-go
│
│ Set #2:
│ - 🛸 artifact terraform-provider-vsphere
│ - 🛸 artifact terraform-provider-gcp
│ - 🛸 artifact candictl
│ - ⛵ image candictl-tests
│ - 🛸 artifact helm
│ - 🛸 artifact controller
│
│ Set #3:
│ - ⛵ image base
│
│ Set #4:
│ - ⛵ image tests
│ - ⛵ image app
└ Concurrent builds plan (no more than 5 images at the same time)
```
