---
title: Overview
sidebar: documentation
permalink: reference/plugging_into_cicd/overview.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

werf represents the new generation of CI/CD tools that integrates low-level building, cleaning, and deploying tools into a single instrument, making it easily pluggable into any existing CI/CD system. All this was made possible by werf following established concepts of all such systems.

werf plugs into a CI/CD system via the so-called *ci-env command*. Ci-env command gathers all required information from the CI/CD system and defines corresponding werf parameters with environment variables that will be referred to as *ci-env*.

![ci-env]({% asset plugging_into_cicd.svg @path %})

In the sections below you will learn:
 * [what is ci-env](#what-is-ci-env) — what information werf gathers from the CI/CD system and why;
 * [ci-env tagging modes](#ci-env-tagging-modes) — what [images]({{ site.baseurl }}/reference/stages_and_images.html#images) tagging modes are available;
 * [how ci-env works](#how-ci-env-works) — how *ci-env command* should be used and how it passes information to other werf commands;
 * [complete list of ci-env params and customizing](#a-complete-list-of-ci-env-parameters) — the complete list of all ci-env params passed to other werf commands and how to customize common parameters.

## What is ci-env?

werf gathers data from CI/CD systems with the *ci-env command* and sets modes of operation to accomplish the following tasks:
 * Docker registry integration;
 * Git integration;
 * CI/CD pipelines integration;
 * CI/CD configuration integration;
 * Configuring modes of operation in CI/CD systems.

### Docker registry integration

Typically, a CI/CD system can provide each job with:
 1. Docker registry address.
 2. Credentials to access the Docker registry.

werf ci-env command should provide authorization for all subsequent commands, allowing them to log in to the detected Docker registry using discovered credentials. Learn more [about the docker login below](#docker-registry-login). Also, the [`DOCKER_CONFIG=PATH_TO_TMP_CONFIG`](#docker_config) variable will be set as a result.

The address of the Docker registry will also be used as the basis for the `--images-repo` parameter. Thus, the [`WERF_IMAGES_REPO=DOCKER_REGISTRY_REPO`](#werf_images_repo) variable will be set.

### Git integration

Typically, a CI/CD system that uses git runs each job in the detached commit state of the git worktree. And the current git-commit, git-tag, or git-branch is passed to the job via environment variables.

The werf ci-env command detects the current git-commit, git-tag, or git-branch and uses them to tag [images]({{ site.baseurl }}/reference/stages_and_images.html#images) described in the `werf.yaml` configuration file. The particular use of the information provided by git depends on the chosen tagging scheme, [see more info below](#ci-env-tagging-modes).

In this case, the [`WERF_TAG_GIT_TAG=GIT_TAG`](#werf_tag_git_tag) or [`WERF_TAG_GIT_BRANCH=GIT_BRANCH`](#werf_tag_git_branch) variable will be set.

### CI/CD pipelines integration

werf can embed any information into annotations and labels of deployed Kubernetes resources. Usually, a CI/CD system exposes various information such as a link to a CI/CD web page of the project, link to the job itself, job and pipeline IDs, and so forth, to users.

The `werf ci-env` command automatically finds the URL of the CI/CD web page of the project and embeds it into annotations of every resource deployed with werf.

The annotation name depends on the selected CI/CD system and is composed as follows: `"project.werf.io/CI_CD_SYSTEM_NAME-url": URL`.

The [`WERF_ADD_ANNOTATION_PROJECT_GIT="project.werf.io/git": URL`](#werf_add_annotation_project_git) environment variable will be set as a result of running the `werf ci-env` command.

Also, werf can automatically pass/set various other *annotations* by analyzing the CI/CD system as well as *custom annotations and labels*; see the [article for more details]({{ site.baseurl }}/reference/deploy_process/deploy_into_kubernetes.html#annotating-and-labeling-chart-resources).

### CI/CD configuration integration

There is a concept of the *environment* in CI/CD systems. The environment defines used host nodes, access parameters, information about the Kubernetes cluster connection, job parameters (e.g., environment variables used), and other data. Typical environments are *development*, *staging*, *testing*, *production*, and *review environments* with support for dynamical names.

werf also uses the concept of an *environment name* in the [deploying process]({{ site.baseurl }}/reference/deploy_process/deploy_into_kubernetes.html#environment).

The `werf ci-env` command identifies the current environment name of the CI/CD system and passes it to all subsequent werf commands automatically. werf prefers the slugged name of the environment if the CI/CD system exports such a name.

Also, the [`WERF_ENV=ENV`](#werf_env) environment variable will be set.

### Configure modes of operation in CI/CD systems

The ci-env command configures [cleanup policies]({{ site.baseurl }}/reference/cleaning_process.html#cleanup-policies) in the following manner:
 * keep no more than 10 images built for git-tags. You can control this behaviour by setting the [`WERF_GIT_TAG_STRATEGY_LIMIT=10`](#werf_git_tag_strategy_limit) environmental variable;
 * keep images built for git-tags for no more than 30 days. You can control this behaviour by setting the [`WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS=30`](#werf_git_tag_strategy_expiry_days) environmental variable.

If the CI/CD system has support for the output of text in different colors, the `werf ci-env` will then set the [`WERF_LOG_COLOR_MODE=on`](#werf_log_color_mode) environment variable.

Logging the project directory where werf runs will be forced if the [`WERF_LOG_PROJECT_DIR=1`](#werf_log_project_dir) environment variable is set (by default, werf does not print the contents of the directory). Such logging helps to debug problems within the CI/CD system by providing a convenient standard output.

The `werf ci-env` command enables the so-called process exterminator. Some CI/CD systems kill job processes with a `SIGKILL` Linux signal when the user hits the `Cancel` button in the user interface. In this case, child processes continue to run until termination. If the process exterminator is enabled, werf constantly checks for its parent processes' PIDs in the background. If one of them has died, werf terminates on its own. The [`WERF_ENABLE_PROCESS_EXTERMINATOR=1`](#werf_enable_process_exterminator) environment variable enables this mode in the CI/CD system (it is disabled by default).

The `werf ci-env` command sets the logging output width to 100 symbols since it is an experimentally proven universal width that fits most modern screens. In this case, the [`WERF_LOG_TERMINAL_WIDTH=100`](#werf_log_terminal_width) variable will be set.

## Ci-env tagging modes

The tagging mode determines how [images]({{ site.baseurl }}/reference/stages_and_images.html#images) that are described in the `werf.yaml` and built by werf will be named during the [publishing process]({{ site.baseurl }}/reference/publish_process.html).

### stages-signature

Werf uses image _stages signature_ to tag result images. Each image defined in the `werf.yaml` config will have an own _stages signature_ which depends on the content of the image and git history which lead to this content.

Learn more about stages signature tagging in the [publish process article]({{ site.baseurl }}/reference/publish_process.html#content-based-tagging). With this tagging strategy werf automatically enables `--tag-by-stages-signature=true` option of `werf publish` command.

This is default and recommended tagging strategy. By omitting `--tagging-strategy`  option of the [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) werf will use `stages-signature` strategy, or user may explicitly specify an option `--tagging-strategy=stages-signature`.

### tag-or-branch

The current git-tag or git-branch is used to tag [images]({{ site.baseurl }}/reference/stages_and_images.html#images) described in the `werf.yaml` and built by werf.

An image, associated with a corresponding git-tag or git-branch, will be kept in the Docker registry in accordance with [cleanup policies]({{ site.baseurl }}/reference/cleaning_process.html#cleanup-policies).

This mode uses [werf publishing parameters]({{ site.baseurl }}/reference/publish_process.html#naming-images) such as `--tag-git-tag` or `--tag-git-branch` and automatically selects the appropriate one. These parameters are also used in the [werf deploy command]({{ site.baseurl }}/cli/main/deploy.html).

The above tagging mode is selected by setting the `--tagging-strategy=tag-or-branch` option of the [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html).

> If the value of tag or branch used does not match a regular expression `^[\w][\w.-]*$` or consists of more than 128 characters, then werf would slugify this tag (read more in the [slug reference]({{ site.baseurl }}/reference/toolbox/slug.html)).
  <br />
  <br />
  For example:
  - the branch `developer-feature` is valid, and the tag would remain unchanged;
  - the branch `developer/feature` is not valid, and the resulting tag will be corrected to `developer-feature-6e0628fc`.

## How ci-env works

The ci-env command passes all parameters to werf via environment variables, see the [pass cli params as environment variables](#pass-cli-parameters-as-environment-variables) section below.

The [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) must be called at the beginning of any CI/CD job before running any other werf command.

**NOTE:** The `werf ci-env` command returns a bash script that exports [werf params by using environment variables](#pass-cli-parameters-as-environment-variables). So to take advantage of the ci-env command, the user must use a `source` shell builtin for the command output. For example:

```shell
source $(werf ci-env gitlab --verbose --as-file)
werf build-and-publish
```

Sourcing ci-env command output will also print all exported variables with its values to the screen.

Here is an example output of the `werf ci-env` command's script without sourcing:

```shell
### DOCKER CONFIG
echo '### DOCKER CONFIG'
export DOCKER_CONFIG="/tmp/werf-docker-config-204033515"
echo 'export DOCKER_CONFIG="/tmp/werf-docker-config-204033515"'

### IMAGES REPO
echo
echo '### IMAGES REPO'
export WERF_IMAGES_REPO="registry.domain.com/project/x"
echo 'export WERF_IMAGES_REPO="registry.domain.com/project/x"'

### TAGGING
echo
echo '### TAGGING'
export WERF_TAG_GIT_TAG="v1.0.3"
echo 'export WERF_TAG_GIT_TAG="v1.0.3"'

### DEPLOY
echo
echo '### DEPLOY'
export WERF_ENV="dev"
echo 'export WERF_ENV="dev"'
export WERF_ADD_ANNOTATION_PROJECT_GIT="project.werf.io/git=https://gitlab.domain.com/project/x"
echo 'export WERF_ADD_ANNOTATION_PROJECT_GIT="project.werf.io/git=https://gitlab.domain.com/project/x"'
export WERF_ADD_ANNOTATION_CI_COMMIT="ci.werf.io/commit=b9a1ddd366aa6a20a0fd43fb6612f349d33465ff"
echo 'export WERF_ADD_ANNOTATION_CI_COMMIT="ci.werf.io/commit=b9a1ddd366aa6a20a0fd43fb6612f349d33465ff"'
export WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL="gitlab.ci.werf.io/pipeline-url=https://gitlab.domain.com/project/x/pipelines/43107"
echo 'export WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL="gitlab.ci.werf.io/pipeline-url=https://gitlab.domain.com/project/x/pipelines/43107"'
export WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL="gitlab.ci.werf.io/job-url=https://gitlab.domain.com/project/x/-/jobs/110681"
echo 'export WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL="gitlab.ci.werf.io/job-url=https://gitlab.domain.com/project/x/-/jobs/110681"'

### IMAGE CLEANUP POLICIES
echo
echo '### IMAGE CLEANUP POLICIES'
export WERF_GIT_TAG_STRATEGY_LIMIT="10"
echo 'export WERF_GIT_TAG_STRATEGY_LIMIT="10"'
export WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS="30"
echo 'export WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS="30"'

### OTHER
echo
echo '### OTHER'
export WERF_LOG_COLOR_MODE="on"
echo 'export WERF_LOG_COLOR_MODE="on"'
export WERF_LOG_PROJECT_DIR="1"
echo 'export WERF_LOG_PROJECT_DIR="1"'
export WERF_ENABLE_PROCESS_EXTERMINATOR="1"
echo 'export WERF_ENABLE_PROCESS_EXTERMINATOR="1"'
export WERF_LOG_TERMINAL_WIDTH="95"
echo 'export WERF_LOG_TERMINAL_WIDTH="95"'
```

### Pass CLI parameters as environment variables

Any parameter of any werf command (except ci-env command itself) that can be passed via CLI can also be passed using environment variables.

There is a common rule of conversion of a CLI parameter into the name of an environment variable: an environment variable gets `WERF_` prefix, consists of capital letters, and dashes `-` are replaced with underscores `_`.

Examples:
 * `--tag-git-tag=mytag` is the same as `WERF_TAG_GIT_TAG=mytag`;
 * `--env=staging` may be replaced by `WERF_ENV=staging`;
 * `--images-repo=myregistry.myhost.com/project/x` is the same as `WERF_IMAGES_REPO=myregistry.myhost.com/project/x`; and so on.

The `--add-label` and `--add-annotation` parameters are an exception to this rule. They can be specified multiple times. Use the following template to set these parameters by environment variables: `WERF_ADD_ANNOTATION_<ARBITRARY_VARIABLE_NAME_SUFFIX>="annoName1=annoValue1"`. For example:

```shell
export WERF_ADD_ANNOTATION_MYANNOTATION_1="annoName1=annoValue1"
export WERF_ADD_ANNOTATION_MYANNOTATION_2="annoName2=annoValue2"
export WERF_ADD_LABEL_MYLABEL_1="labelName1=labelValue1"
export WERF_ADD_LABEL_MYLABEL_2="labelName2=labelValue2"
```

Note that `_MYANNOTATION_1`, `_MYANNOTATION_2`, `_MYLABEL_1`, `_MYLABEL_2` suffixes do not affect names or values of annotations and labels, and are used solely to differentiate between multiple environment variables.

### Docker registry login

As noted in the [Docker registry integration](#docker-registry-integration), werf performs autologin into the detected docker registry.

The ci-env command creates a new temporary [docker config](https://docs.docker.com/engine/reference/commandline/cli/#configuration-files) in all cases.

The temporary docker config prevents the mutual interference of jobs running concurrently. The temporary docker config-based approach is also more secure than logging into the Docker registry using the system-wide default docker config (`~/.docker`).

If a `DOCKER_CONFIG` variable is defined or there is a `~/.docker` file, then werf would *copy* this existing config into a new, temporary one. So any *docker logins* performed before running the `werf ci-env` command will still be active and available using this new temporary config.

After the new tmp config has been created, werf performs additional login into the detected Docker registry.

As a result of a [Docker registry integration procedure](#docker-registry-integration), `werf ci-env` would export a `DOCKER_CONFIG` variable that contains a path to the new temporary docker config.

This config then will be utilized by all subsequent werf commands. Also, it can be used by a `docker login` command to perform any additional custom logins depending on your needs.

### A complete list of ci-env parameters

As an output of the ci-env command, werf exports the following list of variables. To customize these variables, the user may predefine any variable with the prefix `WERF_` before running the `werf ci-env` command (werf will detect the already defined environment variable and will use it as-is). You can achieve this by using the environment variables of the project or by exporting variables in the shell.

The [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) defines the following variables:

#### DOCKER_CONFIG

A path to the new temporary docker config generated by the [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html).

#### WERF_STAGES_STORAGE

As part of the [Docker registry integration](#docker-registry-integration) procedure, [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) identifies the Docker registry and defines the `--stages-storage` parameter using `WERF_STAGES_STORAGE` environment variable.

#### WERF_IMAGES_REPO

As part of the [Docker registry integration](#docker-registry-integration) procedure, [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) identifies the Docker registry and defines the `--images-repo` parameter using `WERF_IMAGES_REPO` environment variable.

#### WERF_TAG_BY_STAGES_SIGNATURE

When `--tagging-strategy=stages-signature` is set explicitly or omitted werf uses `stages-signature` tagging strategy. Werf sets `WERF_TAG_BY_STAGES_SIGNATURE=true` in this case.

#### WERF_TAG_GIT_TAG

As part of the [git integration](#git-integration) procedure, [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) detects if this job is running for a git-tag and optionally defines the `--tag-git-tag` parameter using the `WERF_TAG_GIT_TAG` environment variable.

#### WERF_TAG_GIT_BRANCH

Within the [git integration](#git-integration) procedure, [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) checks if this job is running for a git-branch and optionally defines the `--tag-git-branch` parameter using the `WERF_TAG_GIT_BRANCH` environment variable.

#### WERF_ENV

As part of the [CI/CD configuration integration](#cicd-configuration-integration) procedure, [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) identifies an environment name and defines the `--env` parameter based on the `WERF_ENV` environment variable.

#### WERF_ADD_ANNOTATION_PROJECT_GIT

Within the [CI/CD pipelines integration](#cicd-pipelines-integration) procedure, [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) detects the URL of a project's web page and sets an `--add-annotation` parameter using `WERF_ADD_ANNOTATION_PROJECT_GIT` environment variable.

#### WERF_ADD_ANNOTATION_CI_COMMIT

Under the [CI/CD pipelines integration](#cicd-pipelines-integration) procedure, [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) detects the current commit and sets an`--add-annotation` parameter using the `WERF_ADD_ANNOTATION_CI_COMMIT` environment variable.

#### WERF_GIT_TAG_STRATEGY_LIMIT

As part of the [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-cicd-systems) procedure, [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) sets a `--git-tag-strategy-limit` parameter using the `WERF_GIT_TAG_STRATEGY_LIMIT` environment variable.

#### WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS

As part of the [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-cicd-systems) procedure, [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) sets an `--git-tag-strategy-expiry-days` parameter using the `WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS` environment variable.

#### WERF_LOG_COLOR_MODE

In the [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-cicd-systems) procedure, [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) sets a `--log-color-mode` parameter using the `WERF_LOG_COLOR_MODE` environment variable.

#### WERF_LOG_PROJECT_DIR

In the [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-cicd-systems) procedure, [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) sets a `--log-project-dir` parameter using the `WERF_LOG_PROJECT_DIR` environment variable.

#### WERF_ENABLE_PROCESS_EXTERMINATOR

As part of the [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-cicd-systems) procedure, [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) sets an `--enable-process-exterminator` parameter based on the `WERF_ENABLE_PROCESS_EXTERMINATOR` environment variable.

#### WERF_LOG_TERMINAL_WIDTH

Within the [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-cicd-systems) procedure, [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) sets a `--log-terminal-width` parameter using the `WERF_LOG_TERMINAL_WIDTH` environment variable.

## Further reading

 * [How does an integration with GitLab CI work?]({{ site.baseurl }}/reference/plugging_into_cicd/gitlab_ci.html).
 * [Generalized guide on using werf with any CI/CD system]({{ site.baseurl }}/guides/generic_ci_cd_integration.html).
