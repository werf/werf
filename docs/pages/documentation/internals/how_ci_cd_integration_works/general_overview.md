---
title: General overview
sidebar: documentation
permalink: documentation/internals/how_ci_cd_integration_works/general_overview.html
---

werf represents the new generation of CI/CD tools that integrates low-level building, cleaning, and deploying tools into a single instrument, making it easily pluggable into any existing CI/CD system. All this was made possible by werf following established concepts of all such systems.

werf plugs into a CI/CD system via the so-called *ci-env command*. Ci-env command gathers all required information from the CI/CD system and defines corresponding werf parameters with environment variables that will be referred to as *ci-env*.

![ci-env]({% asset plugging_into_cicd.svg @path %})

In the sections below you will learn:
 * [what is ci-env](#what-is-ci-env) — what information werf gathers from the CI/CD system and why;
 * [how ci-env works](#how-ci-env-works) — how *ci-env command* should be used and how it passes information to other werf commands;
 * [complete list of ci-env params and customizing](#a-complete-list-of-ci-env-parameters) — the complete list of all ci-env params passed to other werf commands and how to customize common parameters.

## What is ci-env?

werf gathers data from CI/CD systems with the *ci-env command* and sets modes of operation to accomplish the following tasks:
 * Docker registry integration;
 * CI/CD pipelines integration;
 * CI/CD configuration integration;
 * Configuring modes of operation in CI/CD systems.

### Docker registry integration

Typically, a CI/CD system can provide each job with:
 1. Docker registry address.
 2. Credentials to access the Docker registry.

werf ci-env command should provide authorization for all subsequent commands, allowing them to log in to the detected Docker registry using discovered credentials. Learn more [about the docker login below](#docker-registry-login). Also, the [`DOCKER_CONFIG=PATH_TO_TMP_CONFIG`](#docker_config) variable will be set as a result.

The address of the Docker registry will also be used as the basis for the `--repo` parameter. Thus, the [`WERF_REPO=DOCKER_REGISTRY_REPO`](#werf_repo) variable will be set.

Almost all built-in Docker registry implementations in CI/CD systems have their own peculiarities and therefore werf must know which implementation it works with (more about the features of the supported Docker registry implementations [here]({{ "documentation/advanced/supported_registry_implementations.html" | true_relative_url }})). If the registry address unambiguously determines the type of implementation, then nothing is additionally required from the user. Otherwise, the type of implementation should be specified by the `--repo-implementation` option (`$WERF_REPO_IMPLEMENTATION`). The `werf ci-env` command sets a value for the built-in Docker registry, if necessary.

### CI/CD pipelines integration

werf can embed any information into annotations and labels of deployed Kubernetes resources. Usually, a CI/CD system exposes various information such as a link to a CI/CD web page of the project, link to the job itself, job and pipeline IDs, and so forth, to users.

The `werf ci-env` command automatically finds the URL of the CI/CD web page of the project and embeds it into annotations of every resource deployed with werf.

The annotation name depends on the selected CI/CD system and is composed as follows: `"project.werf.io/CI_CD_SYSTEM_NAME-url": URL`.

The [`WERF_ADD_ANNOTATION_PROJECT_GIT="project.werf.io/git": URL`](#werf_add_annotation_project_git) environment variable will be set as a result of running the `werf ci-env` command.

Also, werf can automatically pass/set various other *annotations* by analyzing the CI/CD system as well as *custom annotations and labels*; see the [article for more details]({{ "documentation/advanced/helm/basics.html#annotating-and-labeling-of-chart-resources" | true_relative_url }}).

### CI/CD configuration integration

There is a concept of the *environment* in CI/CD systems. The environment defines used host nodes, access parameters, information about the Kubernetes cluster connection, job parameters (e.g., environment variables used), and other data. Typical environments are *development*, *staging*, *testing*, *production*, and *review environments* with support for dynamical names.

werf also uses the concept of an *environment name* in the [deploying process]({{ "documentation/advanced/helm/basics.html#environment" | true_relative_url }}).

The `werf ci-env` command identifies the current environment name of the CI/CD system and passes it to all subsequent werf commands automatically. werf prefers the slugged name of the environment if the CI/CD system exports such a name.

Also, the [`WERF_ENV=ENV`](#werf_env) environment variable will be set.

### Configure modes of operation in CI/CD systems

If the CI/CD system has support for the output of text in different colors, the `werf ci-env` will then set the [`WERF_LOG_COLOR_MODE=on`](#werf_log_color_mode) environment variable.

Logging the project directory where werf runs will be forced if the [`WERF_LOG_PROJECT_DIR=1`](#werf_log_project_dir) environment variable is set (by default, werf does not print the contents of the directory). Such logging helps to debug problems within the CI/CD system by providing a convenient standard output.

The `werf ci-env` command enables the so-called process exterminator. Some CI/CD systems kill job processes with a `SIGKILL` Linux signal when the user hits the `Cancel` button in the user interface. In this case, child processes continue to run until termination. If the process exterminator is enabled, werf constantly checks for its parent processes' PIDs in the background. If one of them has died, werf terminates on its own. The [`WERF_ENABLE_PROCESS_EXTERMINATOR=1`](#werf_enable_process_exterminator) environment variable enables this mode in the CI/CD system (it is disabled by default).

The `werf ci-env` command sets the logging output width to 100 symbols since it is an experimentally proven universal width that fits most modern screens. In this case, the [`WERF_LOG_TERMINAL_WIDTH=100`](#werf_log_terminal_width) variable will be set.

## How ci-env works

The ci-env command passes all parameters to werf via environment variables, see the [pass cli params as environment variables](#pass-cli-parameters-as-environment-variables) section below.

The [`werf ci-env` command]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url }}) must be called at the beginning of any CI/CD job before running any other werf command.

**NOTE:** The `werf ci-env` command returns a bash script that exports [werf params by using environment variables](#pass-cli-parameters-as-environment-variables). So to take advantage of the ci-env command, the user must use a `source` shell builtin for the command output. For example:

```shell
source $(werf ci-env gitlab --verbose --as-file)
werf converge
```

Sourcing ci-env command output will also print all exported variables with its values to the screen.

Here is an example output of the `werf ci-env` command's script without sourcing:

```shell
### DOCKER CONFIG
echo '### DOCKER CONFIG'
export DOCKER_CONFIG="/tmp/werf-docker-config-204033515"
echo 'export DOCKER_CONFIG="/tmp/werf-docker-config-204033515"'

### REPO
echo
echo '### REPO'
export WERF_REPO="registry.domain.com/project/werf"
echo 'export WERF_REPO="registry.domain.com/project/werf"'

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
 * `--env=staging` may be replaced by `WERF_ENV=staging`;
 * `--repo=myregistry.myhost.com/project/werf` is the same as `WERF_REPO=myregistry.myhost.com/project/werf`; and so on.

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

The [`werf ci-env` command]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url }}) defines the following variables:

#### DOCKER_CONFIG

A path to the new temporary docker config generated by the [`werf ci-env` command]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url }}).

#### WERF_REPO

As part of the [Docker registry integration](#docker-registry-integration) procedure, [`werf ci-env` command]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url }}) identifies the Docker registry and defines the `--repo` parameter using `WERF_REPO` environment variable.

#### WERF_REPO_IMPLEMENTATION

As part of the [Docker registry integration](#docker-registry-integration) procedure, [`werf ci-env` command]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url }}) defines the `--repo-implementation` parameter using `WERF_REPO_IMPLEMENTATION` in addition to [`WERF_REPO`](#werf_repo).

#### WERF_ENV

As part of the [CI/CD configuration integration](#cicd-configuration-integration) procedure, [`werf ci-env` command]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url }}) identifies an environment name and defines the `--env` parameter based on the `WERF_ENV` environment variable.

#### WERF_ADD_ANNOTATION_PROJECT_GIT

Within the [CI/CD pipelines integration](#cicd-pipelines-integration) procedure, [`werf ci-env` command]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url }}) detects the URL of a project's web page and sets an `--add-annotation` parameter using `WERF_ADD_ANNOTATION_PROJECT_GIT` environment variable.

#### WERF_ADD_ANNOTATION_CI_COMMIT

Under the [CI/CD pipelines integration](#cicd-pipelines-integration) procedure, [`werf ci-env` command]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url }}) detects the current commit and sets an`--add-annotation` parameter using the `WERF_ADD_ANNOTATION_CI_COMMIT` environment variable.

#### WERF_LOG_COLOR_MODE

In the [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-cicd-systems) procedure, [`werf ci-env` command]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url }}) sets a `--log-color-mode` parameter using the `WERF_LOG_COLOR_MODE` environment variable.

#### WERF_LOG_PROJECT_DIR

In the [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-cicd-systems) procedure, [`werf ci-env` command]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url }}) sets a `--log-project-dir` parameter using the `WERF_LOG_PROJECT_DIR` environment variable.

#### WERF_ENABLE_PROCESS_EXTERMINATOR

As part of the [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-cicd-systems) procedure, [`werf ci-env` command]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url }}) sets an `--enable-process-exterminator` parameter based on the `WERF_ENABLE_PROCESS_EXTERMINATOR` environment variable.

#### WERF_LOG_TERMINAL_WIDTH

Within the [configure modes of operation in CI/CD systems](#configure-modes-of-operation-in-cicd-systems) procedure, [`werf ci-env` command]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url }}) sets a `--log-terminal-width` parameter using the `WERF_LOG_TERMINAL_WIDTH` environment variable.

## Further reading

 * [How does an integration with GitLab CI work?]({{ "documentation/internals/how_ci_cd_integration_works/gitlab_ci_cd.html" | true_relative_url }}).
 * [How does an integration with GitHub Actions work?]({{ "documentation/internals/how_ci_cd_integration_works/github_actions.html" | true_relative_url }}).
 * [Generalized guide on using werf with any CI/CD system]({{ "documentation/advanced/ci_cd/generic_ci_cd_integration.html" | true_relative_url }}).
