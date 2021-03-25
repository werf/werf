---
title: Giterminism
permalink: advanced/giterminism.html
change_canonical: true
---

werf introduces the so-called _giterminism_ mode. The word is constructed from `git` and `determinism`, which means "determined by the git".

To provide a strong guarantee of reproducibility, werf reads the configuration and build's context files from the current git commit, and eliminates external dependencies. In this way, werf strives for easily reproducible configurations and allows developers in Windows, macOS, and Linux to reuse images just built in the CI system simply by switching to the desired commit.

werf does not allow working with uncommitted and untracked files. If the files are not required, they should be ignored with `.gitignore` (including global) or `.helmignore` file.

> We strongly recommend following this approach, but if necessary, you can loosen giterminism restrictions explicitly and enable the features that require careful use with [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }})

During development or debugging, changing the project files might be annoying due to the necessity of creating redundant commits. We are working on the development mode to simplify this process, while keeping the whole logic unchanged. 
Currently, the development mode (activated by the `--dev` option) has two different modes (set by the `--dev-mode` option, _simple_ by default):
- _simple_: for working with the worktree state of the git repository;
- _strict_: for working with [the index state](http://shafiul.github.io/gitbook/1_the_git_index.html) of the git repository.

## Configuration

The configuration of an application may include the following project files:

- The werf configuration (`werf.yaml` by default) that describes all images, deployment, and cleanup settings.
- The werf configuration templates (`.werf/**/*.tmpl`).
- The files that are used with Go-template functions [.Files.Get]({{ "reference/werf_yaml_template_engine.html#filesget" | true_relative_url }}) and [.Files.Glob]({{ "reference/werf_yaml_template_engine.html#filesglob" | true_relative_url }}).
- The helm chart files (`.helm` by default).

> All configuration files must be in the project directory. A symbolic link is supported, but the link must point to a file in the project git repository

By default, werf prohibits using a particular set of directives and Go-template functions that might lead to external dependencies in the werf configuration.

### werf configuration

#### Go-template functions

##### env 

The use of the function [env]({{ "reference/werf_yaml_template_engine.html#env" | true_relative_url }}) complicates the sharing and reproducibility of the configuration in CI jobs and among developers because the value of the environment variable affects the final digest of built images and must be identical at all steps of the pipeline and during local development.

To activate the `env` function it is necessary to use [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }}), but we recommend thinking again about the possible consequences.

#### dockerfile image

##### contextAddFiles

The use of the directive [contextAddFiles]({{ "reference/werf_yaml.html#contextaddfiles" | true_relative_url }}) complicates the sharing and reproducibility of the configuration in CI jobs and among developers because the file data affects the final digest of built images and must be identical at all steps of the pipeline and during local development.

To activate the `contextAddFiles` directive it is necessary to use [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }}), but we recommend thinking again about the possible consequences.

#### stapel image

##### fromLatest

If [fromLatest]({{ "advanced/building_images_with_stapel/base_image.html#from-fromlatest" | true_relative_url }}) is true, then werf starts using the actual _base image_ digest in the stage digest. Thus, using this directive may break the reproducibility of previous builds. The changing of the base image in the registry makes all previously built images unusable.

 * Previous pipeline jobs (e.g., converge) cannot be retried without the image rebuilding after changing a registry base image.
 * If the base image is modified unexpectedly, it may lead to an inexplicably failed pipeline. For instance, the modification occurs after a successful build, and the following jobs will be failed due to changing stages digests alongside base image digest.

As an alternative, we recommend using an unchangeable tag or periodically change [fromCacheVersion]({{ "advanced/building_images_with_stapel/base_image.html#fromcacheversion" | true_relative_url }}) value to guarantee the application's controllable and predictable life cycle.

To activate the `fromLatest` directive it is necessary to use [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }}), but we recommend thinking again about the possible consequences.

##### git

###### branch

[Remote git mapping]({{ "advanced/building_images_with_stapel/git_directive.html#working-with-remote-repositories" | true_relative_url }}) with a branch (master branch by default) may break the previous builds' reproducibility. werf uses the history of a git repository to calculate the stage digest. Thus, the new commit in the branch makes all previously built images unusable.

 * The existing pipeline jobs (e.g., converge) would not run and would require rebuilding an image if a remote git branch has been changed.
 * Unplanned commits to a remote git branch might lead to the pipeline failing seemingly for no apparent reasons. For instance, changes may occur after the build process is completed successfully. In this case, the related pipeline jobs will fail due to changes in stage digests along with the branch HEAD.

As an alternative, we recommend using an unchangeable reference, tag, or commit to guarantee the application's controllable and predictable life cycle.

To activate the `branch` directive it is necessary to use [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }}), but we recommend thinking again about the possible consequences.

##### mount

###### build_dir

The use of the [build_dir mount]({{ "advanced/building_images_with_stapel/mount_directive.html" | true_relative_url }}) may lead to unpredictable behavior when used in parallel and potentially affect reproducibility and reliability.

To activate the `build_dir` mount it is necessary to use [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }}), but we recommend thinking again about the possible consequences.

###### fromPath

The use of the [fromPath mount]({{ "advanced/building_images_with_stapel/mount_directive.html" | true_relative_url }}) may lead to unpredictable behavior when used in parallel and potentially affect reproducibility and reliability. The data in the mounted directory has no effect on the final image digest, which can lead to invalid images and hard-to-trace issues.

To activate the `fromPath` mount it is necessary to use [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }}), but we recommend thinking again about the possible consequences.

## Build's context files

Dockerfile image build context is context (read more about [context]({{ "reference/werf_yaml.html" | true_relative_url }}) directive) files from the current project git repository commit.

Stapel image build context is all files that are added with [git]({{ "advanced/building_images_with_stapel/git_directive.html" | true_relative_url }}) directive from the current project git repository commit.
