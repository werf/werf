---
title: Annotating and labeling of chart resources
permalink: advanced/helm/deploy_process/annotating_and_labeling.html
---

## Auto annotations

werf automatically sets the following built-in annotations to all deployed chart resources:

* `"werf.io/version": FULL_WERF_VERSION` — version of werf used when running the `werf converge` command;
* `"project.werf.io/name": PROJECT_NAME` — project name specified in the `werf.yaml`;
* `"project.werf.io/env": ENV` — environment name specified via the `--env` param or `WERF_ENV` variable; optional, will not be set if env is not used.

werf also sets auto annotations containing information from the CI/CD system used (for example, GitLab CI)  when running the `werf ci-env` command prior to the `werf converge` command. For example, [`project.werf.io/git`]({{ "internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_project_git" | true_relative_url }}), [`ci.werf.io/commit`]({{ "internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_ci_commit" | true_relative_url }}), [`gitlab.ci.werf.io/pipeline-url`]({{ "internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_gitlab_ci_pipeline_url" | true_relative_url }}) and [`gitlab.ci.werf.io/job-url`]({{ "internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_gitlab_ci_job_url" | true_relative_url }}).

For more information about the CI/CD integration, please refer to the following pages:

* [plugging into CI/CD overview]({{ "/internals/how_ci_cd_integration_works/general_overview.html" | true_relative_url }});
* [plugging into GitLab CI]({{ "/internals/how_ci_cd_integration_works/gitlab_ci_cd.html" | true_relative_url }}).

## Custom annotations and labels

The user can pass arbitrary additional annotations and labels using `--add-annotation annoName=annoValue` (can be used repeatedly) and `--add-label labelName=labelValue` (can be used repeatedly) CLI options when invoking werf converge.

For example, you can use the following werf converge invocation to set `commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57`, `gitlab-user-email=vasya@myproject.com`  annotations/labels to all Kubernetes resources in a chart:

```shell
werf converge \
  --add-annotation "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-label "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-annotation "gitlab-user-email=vasya@mydomain.com" \
  --add-label "gitlab-user-email=vasya@mydomain.com" \
  --repo REPO
```
