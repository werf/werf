---
title: Аннотации и лейблы для ресурсов чарта
permalink: usage/deploy/deploy_process/annotating_and_labeling.html
published: false  # TODO: Статья не нужна в деплое
---

## Автоматические аннотации

werf автоматически выставляет следующие встроенные аннотации всем ресурсам чарта в процессе деплоя:

 * `"werf.io/version": FULL_WERF_VERSION` — версия werf, использованная в процессе запуска команды `werf converge`;
 * `"project.werf.io/name": PROJECT_NAME` — имя проекта, указанное в файле конфигурации `werf.yaml`;
 * `"project.werf.io/env": ENV` — имя окружения, указанное с помощью параметра `--env` или переменной окружения `WERF_ENV` (не обязательно, аннотация не устанавливается, если окружение не было указано при запуске).

При использовании команды `werf ci-env` перед выполнением команды `werf converge`, werf также автоматически устанавливает аннотации содержащие информацию из используемой системы CI/CD (например, GitLab CI).

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
