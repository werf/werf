---
title: Overview
sidebar: documentation
permalink: documentation/reference/plugging_into_cicd/overview.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

werf represents the new generation of CI/CD tools that integrates low-level building, cleaning, and deploying tools into a single instrument, making it easily pluggable into any existing CI/CD system. All this was made possible by werf following established concepts of all such systems.

werf plugs into a CI/CD system via the so-called *ci-env command*. Ci-env command gathers all required information from the CI/CD system and defines corresponding werf parameters with environment variables that will be referred to as *ci-env*.

![ci-env]({% asset plugging_into_cicd.svg @path %})

In the following chapters we will see:
 * [what is ci-env](#what-is-ci-env) — what information werf gathers from CI/CD system and why;
 * [ci-env tagging modes](#ci-env-tagging-modes) — about different [images]({{ site.baseurl }}/documentation/reference/stages_and_images.html#images) tagging modes available;
 * [how ci-env works](#how-ci-env-works) — how *ci-env command* should be used and how it passes info to other werf commands;
 * [complete list of ci-env params and customizing](#other-ci-env-params-and-customizing) — complete list of all ci-env params passed to other werf commands and how to customize common params.

## What is ci-env?

werf gathers data from CI/CD systems with the *ci-env command* and sets modes of operation to accomplish the following tasks:
 * Docker registry integration;
 * Git integration;
 * CI/CD pipelines integration;
 * CI/CD configuration integration;
 * Configure modes of operation in CI/CD systems.

### Docker registry integration

Typically, a CI/CD system can provide each job with:
 1. Docker registry address.
 2. Credentials to access the Docker registry.

werf ci-env command should provide authorization for all subsequent commands, allowing them to log in to the detected Docker registry using the discovered credentials. Learn more [about the docker login below](#docker-registry-login). Also, the [`DOCKER_CONFIG=PATH_TO_TMP_CONFIG`](#docker_config) variable will be set as a result.

The address of the Docker registry will also be used as a value for the `--images-repo` parameter. Thus, the [`WERF_IMAGES_REPO=DOCKER_REGISTRY_REPO`](#werf_images_repo) variable will be set.

### Git integration

Typically, a CI/CD system that uses git runs each job in the detached commit state of the git worktree. And the current git-commit, git-tag, or git-branch is passed to the job using environment variables.

The werf ci-env command detects the current git-commit, git-tag, or git-branch and uses them to tag [images]({{ site.baseurl }}/documentation/reference/stages_and_images.html#images) described in the `werf.yaml` configuration file. The particular use of the git information depends on the chosen tagging scheme, [see more info below](#ci-env-tagging-modes).

In this case, the [`WERF_TAG_GIT_TAG=GIT_TAG`](#werf_tag_git_tag) or [`WERF_TAG_GIT_BRANCH=GIT_BRANCH`](#werf_tag_git_branch) variable will be set.

### CI/CD pipelines integration

werf can embed any information into annotations and labels of the deployed Kubernetes resources. Usually, a CI/CD system exposes various information such as a link to a CI/CD web page of the project, link to the job itself, job and pipeline IDs, and so forth, to users.

The `werf ci-env` command automatically finds the URL of the CI/CD web page of the project and embeds it into annotations of every resource deployed with werf.

The annotation name depends on the selected CI/CD system and is composed as follows: `"project.werf.io/CI_CD_SYSTEM_NAME-url": URL`.

The [`WERF_ADD_ANNOTATION_PROJECT_GIT="project.werf.io/git": URL`](#werf_add_annotation_project_git) environment variable will be set as a result of running the `werf ci-env` command.

Also, werf can automatically pass/set various other *annotations* by analyzing the CI/CD system as well as *custom annotations and labels*; see the [article for more details]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#annotate-and-label-chart-resources).

### CI/CD configuration integration

There is a concept of the *environment* in CI/CD systems. The environment defines used host nodes, access parameters, information about the Kubernetes cluster connection, job parameters (e.g., used environment variables), and other data. Typical environments are: *development*, *staging*, *testing*, *production*, and *review environments* with support for dynamical names.

werf also uses the concept of an *environment name* in the [deploying process]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#environment).

The `werf ci-env` command identifies the current environment name of the CI/CD system and passes it to all subsequent werf commands automatically. werf favors the slugged name of the environment if the CI/CD system exports such a name.

Also, the [`WERF_ENV=ENV`](#werf_env) environment variable will be set.

### Configure modes of operation in CI/CD systems

The ci-env command configures [cleanup policies]({{ site.baseurl }}/documentation/reference/cleaning_process.html#cleanup-policies) in the following manner:
 * keep no more than 10 images built for git-tags. This behaviour is determined by setting the [`WERF_GIT_TAG_STRATEGY_LIMIT=10`](#werf_git_tag_strategy_limit) environmental variable;
 * keep images built for git-tags for no more than 30 days. This behaviour is determined by setting the [`WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS=30`](#werf_git_tag_strategy_expiry_days) environmental variable.

If the CI/CD system has support for the outout of text of different colors, thwn the `werf ci-env` will set the [`WERF_LOG_COLOR_MODE=on`](#werf_log_color_mode) environment variable.

Logging the project directory where werf runs will be forced if the [`WERF_LOG_PROJECT_DIR=1`](#werf_log_project_dir) environment variable is set (by default, werf does not print the directory). Such logging helps to debug problems within the CI/CD system by providing a convenient standard output.

The `werf ci-env` command enables the so-called process exterminator. Some CI/CD systems kill job processes with a `SIGKILL` Linux signal when the user hits the `Cancel` button in the user interface. In this case, child processes continue to run until termination. If the process exterminator is enabled, werf constantly checks for its parent processes' PIDs in the background. If one of them has died, then werf terminates on its own. The [`WERF_ENABLE_PROCESS_EXTERMINATOR=1`](#werf_enable_process_exterminator) environment variable enables this mode in the CI/CD system (it is disabled by default). 

The `werf ci-env` command sets the logging output width to 100 symbols since it is an experimentally proven universal width that fits most modern screens. In this case, the [`WERF_LOG_TERMINAL_WIDTH=100`](#werf_log_terminal_width) variable will be set.

## Ci-env tagging modes

The tagging mode determines how [images]({{ site.baseurl }}/documentation/reference/stages_and_images.html#images) that are described in the `werf.yaml` and built by werf will be named during the [publishing process]({{ site.baseurl }}/documentation/reference/publish_process.html).

Currently, the only available mode is a *tag-or-branch* mode.

The support for *content-based tagging* and a corresponding ci-env tagging mode are [coming soon](https://github.com/flant/werf/issues/1184).

### tag-or-branch

The current git-tag or git-branch is used to tag [images]({{ site.baseurl }}/documentation/reference/stages_and_images.html#images) that are described in the `werf.yaml` and built by werf.

An image, associated with a corresponding git-tag or git-branch, will be kept in the Docker registry in accordance with [cleanup policies]({{ site.baseurl }}/documentation/reference/cleaning_process.html#cleanup-policies).

This mode uses [werf publishing parameters]({{ site.baseurl }}/documentation/reference/publish_process.html#naming-images) such as `--tag-git-tag` or `--tag-git-branch` and automatically selects the appropriate one. These parameters are also used in the [werf deploy command]({{ site.baseurl }}/documentation/cli/main/deploy.html).

The above tagging mode is selected by setting the `--tagging-strategy=tag-or-branch` option of the [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html).

> If the value of tag or branch used does not match a regular expression `^[\w][\w.-]*$` or consists of more than 128 characters, then werf would slugify this tag (read more in the [slug reference]({{ site.baseurl }}/documentation/reference/toolbox/slug.html)).
  <br />
  <br />
  For example:
  - the branch `developer-feature` is valid, and the tag will not change;
  - the branch `developer/feature` is not valid, and the resulting tag will be corrected to `developer-feature-6e0628fc`.

## How ci-env works

Ci-env command passes all parameters to werf using environment variables, see [pass cli params as environment variables below](#pass-cli-params-as-environment-variables).

[`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) should be called in the begin of any CI/CD job prior running any other werf commands.

**NOTE** werf ci-env command prints bash script which exports [werf params using environment variables](#pass-cli-params-as-environment-variables). So to actually use ci-env command user must `source` command output using bash. For example:

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

As noted in the [Docker registry integration](#docker-registry-integration) werf performs autologin into detected docker registry.

Ci-env command always creates a new temporal [docker config](https://docs.docker.com/engine/reference/commandline/cli/#configuration-files).

Temporal docker config needed so that parallel running jobs cannot interfere with each other. Also temporal docker config is a more secure way, than logging into Docker registry using system-wide default docker config (`~/.docker`).

If `DOCKER_CONFIG` variable is defined or there is `~/.docker`, then werf will *copy* this already existing config into a new temporal one. So any *docker logins* already made before running `werf ci-env` command will still be active and available using this new temporal config.

After new tmp config was created werf performs additional login into detected Docker registry.

As a result of [Docker registry integration procedure](#docker-registry-integration) `werf ci-env` will export `DOCKER_CONFIG` variable with new temporal docker config.

This config then will be used by any following werf command and also can be used by `docker login` command to perform any additional custom logins on your needs.

### Complete list of ci-env params and customizing

As an output of ci-env command werf exports following list of variables. To customize these variables user may predefine any variable with the name prefix `WERF_` prior running `werf ci-env` command (ci-env command will detect already defined environment variable and will use this variable as is). This can be accomplished using project environment variables or simply by exporting variables in shell.

Following variables are defined by [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html).

#### DOCKER_CONFIG

Path to new temporal docker config generated by [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html).

#### WERF_IMAGES_REPO

Within [Docker registry integration](#docker-registry-integration) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will detect docker registry and define `--images-repo` param using `WERF_IMAGES_REPO` environment variable.

#### WERF_TAG_GIT_TAG

Within [git integration](#git-integration) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will detect if this job is running for git-tag and optionally define `--tag-git-tag` param using `WERF_TAG_GIT_TAG` environment variable.

#### WERF_TAG_GIT_BRANCH

Within [git integration](#git-integration) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will detect if this job is running for git-branch and optionally define `--tag-git-branch` param using `WERF_TAG_GIT_BRANCH` environment variable.

#### WERF_ENV

Within [CI/CD configuration integration](#ci-cd-configuration-integration) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will detect environment name and define `--env` param using `WERF_ENV` environment variable.

#### WERF_ADD_ANNOTATION_PROJECT_GIT

Within [CI/CD pipelines integration](#ci-cd-pipelines-integration) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will detect web page project url and set `--add-annotation` param using `WERF_ADD_ANNOTATION_PROJECT_GIT` environment variable.

#### WERF_ADD_ANNOTATION_CI_COMMIT

Within [CI/CD pipelines integration](#ci-cd-pipelines-integration) procedure [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) will detect current commit and set `--add-annotation` param using `WERF_ADD_ANNOTATION_CI_COMMIT` environment variable.

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

 * [How GitLab CI integration works]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/gitlab_ci.html).
 * [How to use werf with unsupported CI/CD system]({{ site.baseurl }}/documentation/guides/unsupported_ci_cd_integration.html).
