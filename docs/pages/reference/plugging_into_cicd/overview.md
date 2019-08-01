---
title: Overview
sidebar: documentation
permalink: documentation/reference/plugging_into_cicd/overview.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Werf represents the new breed of CI/CD tools that integrates lowlevel building, cleaning and deploying tools into a single tool, which can easily be plugged into any existing CI/CD system. This is possible because werf follows established concepts of all such systems.

Werf plugs into CI/CD system using so called *ci-env command*. Ci-env command is responsible to gather required info from CI/CD system and define corresponding werf params using environment variables which will be referred to as *ci-env*.

![ci-env]({{ site.baseurl }}/images/plugging_into_cicd.svg)

In the following chapters we will see:
 * [what is ci-env](#what-is-ci-env) — what information werf gathers from CI/CD system and why;
 * [ci-env tagging modes](#ci-env-tagging-modes) — about different [images]({{ site.baseurl }}/documentation/reference/stages_and_images.html#images) tagging modes available;
 * [how ci-env works](#how-ci-env-works) — how *ci-env command* should be used and how it passes info to other werf commands;
 * [complete list of ci-env params and customizing](#other-ci-env-params-and-customizing) — complete list of all ci-env params passed to other werf commands and how to customize common params.

## What is ci-env

With *ci-env command* werf gathers data from CI/CD systems and sets modes of operation to accomplish the following tasks:
 * Docker registry integration;
 * Git integration;
 * CI/CD pipelines integration;
 * CI/CD configuration integration;
 * Configure modes of operation in CI/CD systems.

### Docker registry integration

Typically CI/CD system can provide each job with:
 1. Docker registry address.
 2. Credentials to access docker registry.

Werf ci-env command should perform login into detected docker registry using detected credentials. See more [info about docker login below](#docker-registry-login). [`DOCKER_CONFIG=PATH_TO_TMP_CONFIG`](#docker_config) will be set.

Docker registry address will also be used as `--images-repo` parameter value. [`WERF_IMAGES_REPO=DOCKER_REGISTRY_REPO`](#werf_images_repo) will be set.

### Git integration

Typically CI/CD system that uses git runs each job in the detached commit state of git worktree. And current git-commit, git-tag or git-branch are passed to the job using environment variables.

Werf ci-env command detects current git-commit, git-tag or git-branch and uses this info to tag [images]({{ site.baseurl }}/documentation/reference/stages_and_images.html#images) from `werf.yaml` config. This usage of git info depends on the selected tagging scheme, [more info below](#ci-env-tagging-modes).

[`WERF_TAG_GIT_TAG=GIT_TAG`](#werf_tag_git_tag) or [`WERF_TAG_GIT_BRANCH=GIT_BRANCH`](#werf_tag_git_branch) will be set.

### CI/CD pipelines integration

Werf can embed any info into deployed kubernetes resources annotations and labels. Typically CI/CD system exposes for users such info as link to CI/CD web page of the project, link to job itself, job id and pipeline id and other info.

Werf ci-env command automatically detects CI/CD web page project url and embeds this url into annotations of every resource deployed with werf.

Annotation name depends on the selected CI/CD system and constructed as follows: `"project.werf.io/CI_CD_SYSTEM_NAME-url": URL`.

[`WERF_ADD_ANNOTATION_GIT_REPOSITORY_URL="project.werf.io/CI_CD_SYSTEM_NAME-url": URL`](#werf_add_annotation_git_repository_url) will be set.

There are another *auto annotations* set by werf using any CI/CD system, also *custom annotations and labels* can be passed, see [deploy article for details]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#annotate-and-label-chart-resources).

### CI/CD configuration integration

There is a concept used in CI/CD systems named *environment*. Environment can define used host nodes, access parameters, kubernetes cluster connection info, job parameters (using environment variables for example) and other info. Typical environments are: *development*, *staging*, *testing*, *production* and *review environments* with dynamical names.

Werf also uses concept of *environment name* in the [deploy process]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#environment).

Ci-env command detects current environment name of CI/CD system and passes it to all subsequent werf commands automatically. Slugged name of the environment will be preferred if CI/CD system exports such a name.

[`WERF_ENV=ENV`](#werf_env) will be set.

### Configure modes of operation in CI/CD systems

Ci-env command configures [cleanup policies]({{ site.baseurl }}/documentation/reference/cleaning_process.html#cleanup-policies) as follows:
 * keep no more than 10 images built for git-tags, [`WERF_GIT_TAG_STRATEGY_LIMIT=10`](#werf_git_tag_strategy_limit) will be set;
 * keep images built for git-tags no more than 30 days, [`WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS=30`](#werf_git_tag_strategy_expiry_days) will be set.

Colorized output of werf will be forced if CI/CD system has support for it. [`WERF_LOG_COLOR_MODE=on`](#werf_log_color_mode) will be set.

Logging of project directory where werf runs will be forced: convinient common output to debug problems within CI/CD system, by default when running werf it will not print the directory. [`WERF_LOG_PROJECT_DIR=1`](#werf_log_project_dir) will be set.

So called werf process exterminator will be enabled. Some CI/CD systems will kill job process with `SIGKILL` linux signal when user hits the `Cancel` button in the user interface. Child processes of this job will continue to run till termination in this case. Werf with enabled process exterminator will constantly check in the background for its parent processes pids and if one of them has died, then werf will be terminated by itself. This mode will be enabled in CI/CD system and is disabled by default. [`WERF_ENABLE_PROCESS_EXTERMINATOR=1`](#werf_enable_process_exterminator) will be set.

Logging output widht will be forced to 100 symbols, which is experimentally proven universal width to support most of the typical today screens. [`WERF_LOG_TERMINAL_WIDTH=100`](#werf_log_terminal_width) will be set.

## Ci-env tagging modes

Tagging mode determines how [images]({{ site.baseurl }}/documentation/reference/stages_and_images.html#images) from `werf.yaml` built by werf will be named in [publish process]({{ site.baseurl }}/documentation/reference/publish_process.html).

The only available mode of ci-env for now is *tag-or-branch* mode.

*Content based tagging* support and corresponding ci-env tagging mode is [coming soon](https://github.com/flant/werf/issues/1184).

### tag-or-branch

Current git-tag or git-branch will be used to tag [images]({{ site.baseurl }}/documentation/reference/stages_and_images.html#images) from `werf.yaml` built by werf.

Image related with git-tag and git-branch will be kept in the docker registry accourding to [cleanup policies]({{ site.baseurl }}/documentation/reference/cleaning_process.html#cleanup-policies).

This mode makes use of [werf publish params]({{ site.baseurl }}/documentation/reference/publish_process.html#images-naming): `--tag-git-tag` or `--tag-git-branch` — automatically choosing appropriate one. These params also used in the [werf deploy command]({{ site.baseurl }}/documentation/cli/main/deploy.html).

This tagging mode is selected by `--tagging-strategy=tag-or-branch` option of [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html).

## How ci-env works

Ci-env command passes all paramters to werf using environment variables, see [pass cli params as environment variables below](#pass-cli-params-as-environment-variables).

[`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) should be called in the begin of any CI/CD job prior running any other werf commands.

**NOTE** Werf ci-env command prints bash script which exports [werf params using environment variables](#pass-cli-params-as-environment-variables). So to actually use ci-env command user must `source` command output using bash. For example:

```
source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
werf build-and-publish --stages-storage :local
```

Sourcing ci-env command output will also result in printing all exported variables with values to the screen.

Example of `werf ci-env` command output script without sourcing:

```bash
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
export WERF_ADD_ANNOTATION_GIT_REPOSITORY_URL="project.werf.io/gitlab-url=https://gitlab.domain.com/project/x"
echo 'export WERF_ADD_ANNOTATION_GIT_REPOSITORY_URL="project.werf.io/gitlab-url=https://gitlab.domain.com/project/x"'

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
export WERF_LOG_TERMINAL_WIDTH="100"
echo 'export WERF_LOG_TERMINAL_WIDTH="100"'
```

### Pass CLI params as environment variables

Any parameter of any werf command (except ci-env command itself) that can be passed via CLI, can also be passed using environment variables.

There is a common rule of conversion of CLI param to environment variable name: environment variable have `WERF_` prefix, consists of capital letters, underscore symbol `_` instead of dash `-`.

Examples:
 * `--tag-git-tag=mytag` can be specified by `WERF_TAG_GIT_TAG=mytag`;
 * `--env=staging` can be specified by `WERF_ENV=staging`;
 * `--images-repo=myregistry.myhost.com/project/x` can be specified by `WERF_IMAGES_REPO=myregistry.myhost.com/project/x`; etc.

Exception to this rule is `--add-label` and `--add-annotation` params, which can be specified multiple times. To specify this params using environment variables use following pattern: `WERF_ADD_ANNOTATION_<ARBITRARY_VARIABLE_NAME_SUFFIX>="annoName1=annoValue1"`. For example:

```bash
export WERF_ADD_ANNOTATION_MYANNOTATION_1="annoName1=annoValue1"
export WERF_ADD_ANNOTATION_MYANNOTATION_2="annoName2=annoValue2"
export WERF_ADD_LABEL_MYLABEL_1="labelName1=labelValue1"
export WERF_ADD_LABEL_MYLABEL_2="labelName2=labelValue2"
```

Note that `_MYANNOTATION_1`, `_MYANNOTATION_2`, `_MYLABEL_1`, `_MYLABEL_2` suffixes does not affect annotations and labels names or values and used only to differentiate between multiple environment variables.

### Docker registry login

As noted in the [docker registry integration](#docker-registry-integration) werf performs autologin into detected docker registry.

Ci-env command always creates a new temporal [docker config](https://docs.docker.com/engine/reference/commandline/cli/#configuration-files).

Temporal docker config needed so that parallel running jobs cannot interfere with each other. Also temporal docker config is a more secure way, than logging into docker registry using system-wide default docker config (`~/.docker`).

If `DOCKER_CONFIG` variable is defined or there is `~/.docker`, then werf will *copy* this already existing config into a new temporal one. So any *docker logins* already made before running `werf ci-env` command will still be active and available using this new temporal config.

After new tmp config was created werf performs additional login into detected docker registry.

As a result of [docker registry integration procedure](#docker-registry-integration) `werf ci-env` will export `DOCKER_CONFIG` variable with new temporal docker config.

This config then will be used by any following werf command and also can be used by `docker login` command to perform any additional custom logins on your needs.

### Complete list of ci-env params and customizing

As an output of ci-env command werf exports following list of variables. To customize these variables user may predefine any variable with the name prefix `WERF_` prior running `werf ci-env` command (ci-env command will detect already defined environment variable and will use this variable as is). This can be accomplished using project environment variables or simply by exporting variables in shell.

Following variables are defined by [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html).

#### DOCKER_CONFIG

Path to new temporal docker config generated by [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html).

#### WERF_IMAGES_REPO

Within [docker registry integration](#docker-registry-integration) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will detect docker registry and define `--images-repo` param using `WERF_IMAGES_REPO` environment variable.

#### WERF_TAG_GIT_TAG

Within [git integration](#git-integration) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will detect if this job is running for git-tag and optionally define `--tag-git-tag` param using `WERF_TAG_GIT_TAG` environment variable.

#### WERF_TAG_GIT_BRANCH

Within [git integration](#git-integration) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will detect if this job is running for git-branch and optionally define `--tag-git-branch` param using `WERF_TAG_GIT_BRANCH` environment variable.

#### WERF_ENV

Within [CI/CD configuration integration](#ci-cd-configuration-integration) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will detect environment name and define `--env` param using `WERF_ENV` environment variable.

#### WERF_ADD_ANNOTATION_GIT_REPOSITORY_URL

Within [CI/CD pipelines integration](#ci-cd-pipelines-integration) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will detect web page project url and set `--add-annotation` param using `WERF_ADD_ANNOTATION_GIT_REPOSITORY_URL` environment variable.

#### WERF_GIT_TAG_STRATEGY_LIMIT

Within [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-ci-cd-systems) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will set `--git-tag-strategy-limit` param using `WERF_GIT_TAG_STRATEGY_LIMIT` environment variable.

#### WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS

Within [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-ci-cd-systems) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will set `--git-tag-strategy-expiry-days` param using `WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS` environment variable.

#### WERF_LOG_COLOR_MODE

Within [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-ci-cd-systems) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will set `--log-color-mode` param using `WERF_LOG_COLOR_MODE`.

#### WERF_LOG_PROJECT_DIR

Within [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-ci-cd-systems) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will set `--log-project-dir` param using `WERF_LOG_PROJECT_DIR`.

#### WERF_ENABLE_PROCESS_EXTERMINATOR

Within [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-ci-cd-systems) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will set `--enable-process-exterminator` param using `WERF_ENABLE_PROCESS_EXTERMINATOR`.

#### WERF_LOG_TERMINAL_WIDTH

Within [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-ci-cd-systems) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will set `--log-terminal-width` param using `WERF_LOG_TERMINAL_WIDTH`.

## Further reading

 * [How Gitlab CI integration works]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/gitlab_ci.html).
 * [How to use werf with unsupported CI/CD system]({{ site.baseurl }}/documentation/guides/unsupported_ci_cd_integration.html).

