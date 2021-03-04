---
title: Аннотации и лейблы для ресурсов чарта
permalink: documentation/advanced/helm/deploy_process/annotating_and_labeling.html
---

## Автоматические аннотации

werf автоматически выставляет следующие встроенные аннотации всем ресурсам чарта в процессе деплоя:

 * `"werf.io/version": FULL_WERF_VERSION` — версия werf, использованная в процессе запуска команды `werf converge`;
 * `"project.werf.io/name": PROJECT_NAME` — имя проекта, указанное в файле конфигурации `werf.yaml`;
 * `"project.werf.io/env": ENV` — имя окружения, указанное с помощью параметра `--env` или переменной окружения `WERF_ENV` (не обязательно, аннотация не устанавливается, если окружение не было указано при запуске).

При использовании команды `werf ci-env` перед выполнением команды `werf converge`, werf также автоматически устанавливает аннотации содержащие информацию из используемой системы CI/CD (например, GitLab CI).
Например, [`project.werf.io/git`]({{ "documentation/internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_project_git" | true_relative_url }}), [`ci.werf.io/commit`]({{ "documentation/internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_ci_commit" | true_relative_url }}), [`gitlab.ci.werf.io/pipeline-url`]({{ "documentation/internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_gitlab_ci_pipeline_url" | true_relative_url }}) и [`gitlab.ci.werf.io/job-url`]({{ "documentation/internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_gitlab_ci_job_url" | true_relative_url }}).

Для более подробной информации об интеграции werf с системами CI/CD читайте статьи по темам:

 * [Общие сведения по работе с CI/CD системами]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html" | true_relative_url }});
 * [Работа GitLab CI]({{ "documentation/internals/how_ci_cd_integration_works/gitlab_ci_cd.html" | true_relative_url }}).

## Пользовательские аннотации и лейблы

Пользователь может устанавливать произвольные аннотации и лейблы используя CLI-параметры при деплое `--add-annotation annoName=annoValue` (может быть указан несколько раз) и `--add-label labelName=labelValue` (может быть указан несколько раз).

Например, для установки аннотаций и лейблов `commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57`, `gitlab-user-email=vasya@myproject.com` всем ресурсам Kubernetes в чарте, можно использовать следующий вызов команды деплоя:

```shell
werf converge \
  --add-annotation "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-label "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-annotation "gitlab-user-email=vasya@myproject.com" \
  --add-label "gitlab-user-email=vasya@myproject.com" \
  --env dev \
  --repo REPO
```
