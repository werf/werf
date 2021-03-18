---
title: GitLab CI
sidebar: documentation
permalink: reference/plugging_into_cicd/gitlab_ci.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

According to the description of [ci-env command]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#what-is-ci-env), werf should define a set of `WERF_*` variables and perform some actions to integrate with the CI/CD system.

werf uses the following values for werf environment variables:

### WERF_IMAGES_REPO

The value of [`WERF_IMAGES_REPO`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_images_repo) is derived from the [`CI_REGISTRY_IMAGE`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable.

### WERF_TAG_BY_STAGES_SIGNATURE

The value of [`WERF_TAG_BY_STAGES_SIGNATURE`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_tag_by_stages_signature) is set to `true` when using [`stages-signature`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#stages-signature) tagging strategy (which is default).

### WERF_TAG_GIT_TAG

The value of [`WERF_TAG_GIT_TAG`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_tag_git_tag) is taken from the [`CI_COMMIT_TAG`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable.

### WERF_TAG_GIT_BRANCH

The value of [`WERF_TAG_GIT_BRANCH`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_tag_git_branch) is taken from the [`CI_COMMIT_REF_NAME`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable.

### WERF_ADD_ANNOTATION_PROJECT_GIT

The value of [`WERF_ADD_ANNOTATION_PROJECT_GIT`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_add_annotation_project_git) is based on the [`CI_PROJECT_URL`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable and composed as follows:

```
project.werf.io/git=$CI_PROJECT_URL
```

### WERF_ADD_ANNOTATION_CI_COMMIT

The value of [`WERF_ADD_ANNOTATION_CI_COMMIT`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_add_annotation_ci_commit) is extracted from the [`CI_COMMIT_SHA`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable and composed as follows:

```
ci.werf.io/commit=$CI_COMMIT_SHA
```

### WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL

The value of `WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL` is derived from the [`CI_PIPELINE_ID`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable and composed in the following way:

```
gitlab.ci.werf.io/pipeline-url=$CI_PROJECT_URL/pipelines/$CI_PIPELINE_ID
```

### WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL

The value of `WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL` is taken from the [`CI_JOB_ID`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable and composed as follows:

```
gitlab.ci.werf.io/job-url=$CI_PROJECT_URL/-/jobs/$CI_JOB_ID
```

### WERF_ENV

GitLab supports [environments](https://docs.gitlab.com/ce/ci/environments/index.html). werf will detect the current environment for the GitLab pipeline and use it as an environment parameter.

The value of [`WERF_ENV`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_env) is extracted from the [`CI_ENVIRONMENT_SLUG`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable.

### Other variables

Other variables are configured in the regular way described in the [overview article]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html):
 * [`DOCKER_CONFIG`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#docker_config);
 * [`WERF_GIT_TAG_STRATEGY_LIMIT`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_git_tag_strategy_limit);
 * [`WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_git_tag_strategy_expiry_days);
 * [`WERF_LOG_COLOR_MODE`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_log_color_mode);
 * [`WERF_LOG_PROJECT_DIR`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_log_project_dir);
 * [`WERF_ENABLE_PROCESS_EXTERMINATOR`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_enable_process_exterminator);
 * [`WERF_LOG_TERMINAL_WIDTH`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_log_terminal_width).

## How to use

You can turn on the integration with GitLab CI by invoking the [`werf ci-env` command]({{ site.baseurl }}/cli/toolbox/ci_env.html) with the required positional argument:

```shell
werf ci-env gitlab
```
