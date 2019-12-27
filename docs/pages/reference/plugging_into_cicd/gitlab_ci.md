---
title: GitLab CI
sidebar: documentation
permalink: documentation/reference/plugging_into_cicd/gitlab_ci.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

According to [ci-env description]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#what-is-ci-env) werf should define a set of `WERF_*` variables and perform some actions to integrate with CI/CD system.

werf uses following values for werf environment variables:

### WERF_IMAGES_REPO

[`WERF_IMAGES_REPO`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_images_repo) value is taken from [`CI_REGISTRY_IMAGE`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable.

### WERF_TAG_GIT_TAG

[`WERF_TAG_GIT_TAG`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_tag_git_tag) value is taken from [`CI_COMMIT_TAG`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable.

### WERF_TAG_GIT_BRANCH

[`WERF_TAG_GIT_BRANCH`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_tag_git_branch) value is taken from [`CI_COMMIT_REF_NAME`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable.

### WERF_ADD_ANNOTATION_PROJECT_GIT

[`WERF_ADD_ANNOTATION_PROJECT_GIT`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_add_annotation_project_git) value is taken from [`CI_PROJECT_URL`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable and constructed as:

```
project.werf.io/git=$CI_PROJECT_URL
```

### WERF_ADD_ANNOTATION_CI_COMMIT

[`WERF_ADD_ANNOTATION_CI_COMMIT`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_add_annotation_ci_commit) value is taken from [`CI_COMMIT_SHA`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable and constructed as:

```
ci.werf.io/commit=$CI_COMMIT_SHA
```

### WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL

`WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL` value is taken from [`CI_PIPELINE_ID`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable and constructed as:

```
gitlab.ci.werf.io/pipeline-url=$CI_PROJECT_URL/pipelines/$CI_PIPELINE_ID
```

### WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL

`WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL` value is taken from [`CI_JOB_ID`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable and constructed as:

```
gitlab.ci.werf.io/job-url=$CI_PROJECT_URL/-/jobs/$CI_JOB_ID
```

### WERF_ENV

GitLab has [environments support](https://docs.gitlab.com/ce/ci/environments.html). werf will detect current environment for the pipeline in gitlab and use it as environment parameter.

[`WERF_ENV`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_env) value is taken from [`CI_ENVIRONMENT_SLUG`](https://docs.gitlab.com/ee/ci/variables/) gitlab environment variable.

### Other variables

Other variables are configured in the common way described in the [overview article]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html):
 * [`DOCKER_CONFIG`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#docker_config);
 * [`WERF_GIT_TAG_STRATEGY_LIMIT`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_git_tag_strategy_limit);
 * [`WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_git_tag_strategy_expiry_days);
 * [`WERF_LOG_COLOR_MODE`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_log_color_mode);
 * [`WERF_LOG_PROJECT_DIR`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_log_project_dir);
 * [`WERF_ENABLE_PROCESS_EXTERMINATOR`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_enable_process_exterminator);
 * [`WERF_LOG_TERMINAL_WIDTH`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_log_terminal_width).

## How to use

GitLab CI is turned on in [`werf ci-env` command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) by required positional argument:

```shell
werf ci-env gitlab --tagging-strategy ...
```
