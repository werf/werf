---
title: Интеграция с CI/CD-системами
permalink: usage/integration_with_ci_cd_systems.html
---

## Обзор

При использовании CI/CD-системы пользователь должен учитывать её особенности и интегрироваться с существующими примитивами и компонентами.

werf предлагает готовую интеграцию для GitLab CI/CD и GitHub Actions. Используя служебные переменные окружения CI-заданий, интеграция выполняет следующие действия:

- Создание временной Docker-конфигурации на основе текущей пользовательской и авторизация в container registry CI.
- Простановка значений по умолчанию для команд werf:
  - Использование container registry CI (`WERF_REPO`).
  - Определение текущего окружения (`WERF_ENV`).
  - Аннотирование выкатываемых ресурсов чарта, связывание с CI-системой (`WERF_ADD_ANNOTATION_*`). Во все выкатываемые ресурсы добавляются аннотации, которые позволяют пользователю перейти в связанный пайплайн, задание и коммит при необходимости.
  - Настройка логирования werf (`WERF_LOG_*`).
  - Включение автоматической очистки процессов werf для отменённых CI-заданий (`WERF_ENABLE_PROCESS_EXTERMINATOR=1`). Процедура требуется только в тех CI-системах, которые не умеют посылать сигнал о завершении порожденным процессам (например, в GitLab CI/CD).

## GitLab CI/CD

Вся интеграция сводится к вызову команды [ci-env]({{ "reference/cli/werf_ci_env.html" | true_relative_url }}) и последующим выполнением инструкций, которые команда выводит в stdout.

```shell
. $(werf ci-env gitlab --as-file)
```

Далее в рамках shell-сессии все команды werf будут использовать проставленные значения по умолчанию и работать с container registry CI.

К примеру, стадия пайплайна для выката на production может выглядеть следующим образом:

```yaml
Deploy to Production:
  stage: deploy
  script:
    - . $(werf ci-env gitlab --as-file)
    - werf converge
  environment:
    name: production
```

> Полный `.gitlab-ci.yml` для готовых рабочих процессов, а также особенности использования werf c Shell, Docker и Kubernetes executors можно найти в соответствующих статьях руководства:
>
> - [Готовые рабочие процессы](/guides/nodejs/400_ci_cd_workflow/030_gitlab_ci_cd/010_workflows.html);
> - [Docker executor](/guides/nodejs/400_ci_cd_workflow/030_gitlab_ci_cd/020_docker_executor.html);
> - [Kubernetes executor](/guides/nodejs/400_ci_cd_workflow/030_gitlab_ci_cd/030_kubernetes_executor.html).

## GitHub Actions

По аналогии с GitLab CI/CD интеграция сводится к вызову команды [ci-env]({{ "reference/cli/werf_ci_env.html" | true_relative_url }}) и последующим выполнением инструкций, которые команда выводит в stdout.

```shell
. $(werf ci-env github --as-file)
```

Далее в рамках шага все команды werf будут использовать проставленные значения по умолчанию и работать с container registry CI.

К примеру, задание для выката на production может выглядеть следующим образом:

{% raw %}
```yaml
converge:
  name: Converge
  runs-on: ubuntu-latest
  steps:

    - name: Checkout code
      uses: actions/checkout@v3
      with:
        fetch-depth: 0

    - name: Install werf
      uses: werf/actions/install@v1.2

    - name: Run script
      run: |
        . $(werf ci-env github --as-file)
        werf converge
      env:
        WERF_KUBECONFIG_BASE64: ${{ secrets.KUBE_CONFIG_BASE64_DATA }}
        WERF_ENV: production
```
{% endraw %}

> Полный набор конфигураций (`.github/workflows/*.yml`) для готовых рабочих процессов можно найти [в соответствующей статье руководства](/guides/nodejs/400_ci_cd_workflow/040_github_actions.html).
