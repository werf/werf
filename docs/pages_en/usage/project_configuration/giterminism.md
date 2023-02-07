---
title: Giterminism
permalink: usage/project_configuration/giterminism.html
---

## About giterminism

To ensure consistency and guarantee reproducibility, werf introduces the so-called mechanism of _giterminism_. Its name comes from a combination of the words `git` and `determinism` and refers to "deterministic Git" mode. In this mode, werf reads the configuration and build context from the current project repository commit and, by default, does not permit the use of external dependencies.

Any user drift from giterminism must be documented in a dedicated `werf-giterminism.yaml` file so that the configuration management process remains meaningful and the usage is transparent to all participants of the delivery cycle.

## Excluding unused files

werf does not permit working with uncommitted and untracked files in Git. If files are not required, they should be explicitly excluded using the `.gitignore` and `.helmignore` files.

## Using uncommitted and untracked files in debugging and development

When debugging and developing, making changes to project files can be inconvenient due to the need to create intermediate commits. We are implementing a development mode to streamline this process and at the same time keep the general logic of the project unchanged.

In current versions, development mode (activated by the `--dev` option) allows you to work with the worktree state of the project's Git repository as well as with tracked and untracked files. werf ignores changes based on the rules defined in `.gitignore` as well as the rules set by the user using the `--dev-ignore=<glob>` option (can be used multiple times).

##  Enabling nondeterministic functionality on a case-by-case basis

### werf.yaml templating

#### Functions of the Go templating engine

##### env

Using [env]({{"reference/werf_yaml_template_engine.html#env-1" | true_relative_url }}) makes it difficult to share and reproduce the configuration in CI jobs and among developers, because the value of the environment variable affects the final digest of the images being built, so the value must be the same at all stages of the CI pipeline and in local development during reproduction.

The `env` function can be activated using [werf-giterminism.yaml]({{"reference/werf_giterminism_yaml.html" | true_relative_url }}), but we strongly recommend that you carefully consider the possible implications of this.

### Assembling

#### Dockerfile image

##### contextAddFiles

Using the `contextAddFiles` directive complicates configuration sharing and reproducibility in CI jobs and among developers, since file data affects the final digest of the built images and must be identical at all stages of the CI pipeline and in local development during reproduction.

The `contextAddFiles` directive can be activated using [werf-giterminism.yaml]({{"reference/werf_giterminism_yaml.html" | true_relative_url }}), but we strongly recommend that you carefully consider the possible implications of this.

#### Stapel image

##### fromLatest

When using [fromLatest]({{"usage/build/stapel/base.html#from-fromlatest" | true_relative_url }}), werf factors in the real _base image_ digest when calculating the build image digest. Thus, using this directive can break the reproducibility of earlier builds. Changing the base image in the container registry would render all previously built images unusable.

* Any preceding jobs of the CI pipeline (e.g., converge) will not run without first rebuilding the image, once the base image in the container registry has been changed.
* Changing the base image in the container registry will result in unexpected CI pipeline failures. Suppose the change occurs after a successful build. In this case, subsequent jobs will fail because the final image digest has been changed along with the base image digest.

As an alternative, we recommend using an immutable tag or periodically changing the value of [fromCacheVersion]({{"usage/build/stapel/base.html#fromcacheversion" | true_relative_url }}) to ensure a manageable and predictable application lifecycle.

The `fromLatest` directive can be activated using [werf-giterminism.yaml]({{"reference/werf_giterminism_yaml.html" | true_relative_url }}), but we strongly recommend that you carefully consider the possible implications of this.

##### git

###### branch

Using the branch directive (the default is `master` branch) when working with [arbitrary git repositories]({{"usage/build/stapel/git.html#working-with-remote-repositories" | true_relative_url }}) can break reproducibility of previous builds. werf uses the history of a git repository to calculate the digest of the image being built. Thus, a new commit in a branch will render all previously built images unusable.

* The existing jobs of the CI pipeline (e.g., converge) will not run and will require rebuilding the image if the branch's HEAD has been changed.
* Out-of-order commits to a branch can cause the CI pipeline to fail for no apparent reason. Suppose, a change has been made after the build process has been successfully completed. In this case, all the related CI pipeline jobs will fail due to changes in digests of the images being built, along with the HEAD of the corresponding branch.

As an alternative, we recommend using an immutable link, tag, or commit to guarantee a manageable and predictable application lifecycle.

The `branch` directive can be activated using [werf-giterminism.yaml]({{"reference/werf_giterminism_yaml.html" | true_relative_url }}), but we strongly recommend that you carefully consider the possible implications of this.

##### mount

###### build_dir

Using the [build_dir]({{"usage/build/stapel/mounts.html" | true_relative_url }}) mounting directive may lead to unpredictable behavior if used in parallel and potentially affect reproducibility and reliability.

The `build_dir` directive can be activated using [werf-giterminism.yaml]({{"reference/werf_giterminism_yaml.html" | true_relative_url }}), but we strongly recommend that you carefully consider the possible implications of this.

###### fromPath

The use of the [fromPath]({{"usage/build/stapel/mounts.html" | true_relative_url }}) mounting directive may result in unpredictable behavior if used in parallel and potentially affect reproducibility and reliability. The data in the mount directory does not affect the final digest of the built image, which may result in invalid images as well as hard to track problems.

The `fromPath` directive can be activated using [werf-giterminism.yaml]({{"reference/werf_giterminism_yaml.html" | true_relative_url }}), but we strongly recommend that you carefully consider the possible implications of this.

### Deploying

#### The --use-custom-tag option

The use of tag aliases with immutable values (e.g., `%image%-master`) makes previous deploys unreproducible and requires setting the `imagePullPolicy: Always` policy for each image when configuring application containers in the Helm chart.

The `--use-custom-tag` oprion can be activated using [werf-giterminism.yaml]({{"reference/werf_giterminism_yaml.html" | true_relative_url }}), but we strongly recommend that you carefully consider the possible implications of this.
