---
title: Annotating and labeling of chart resources
permalink: usage/deploy/deploy_process/annotating_and_labeling.html
published: false # TODO: this belongs to ci-env docs
---

## Auto annotations

werf automatically sets the following built-in annotations to all deployed chart resources:

* `"werf.io/version": FULL_WERF_VERSION` — version of werf used when running the `werf converge` command;
* `"project.werf.io/name": PROJECT_NAME` — project name specified in the `werf.yaml`;
* `"project.werf.io/env": ENV` — environment name specified via the `--env` param or `WERF_ENV` variable; optional, will not be set if env is not used.

werf also sets auto annotations containing information from the CI/CD system used (for example, GitLab CI) when running the `werf ci-env` command prior to the `werf converge` command.

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
