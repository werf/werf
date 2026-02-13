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

### Изолированная Docker-конфигурация

По умолчанию `werf ci-env` выбирает Docker-конфигурацию в следующем порядке приоритета: значение из `--docker-config`, затем `WERF_DOCKER_CONFIG`, затем `DOCKER_CONFIG`, и только если ни одно из них не задано — директория `~/.docker` на хосте.
Выбранная конфигурация затем копируется во временную директорию для CI-задания. На shared CI-раннерах это может приводить к конфликтам между заданиями или утечке credentials.

Используйте флаг `--init-tmp-docker-config`, чтобы werf создал изолированную пустую Docker-конфигурацию вместо копирования любой существующей:

```shell
. $(werf ci-env gitlab --as-file --init-tmp-docker-config)
```

Этот флаг:
- Создаёт новую временную директорию с пустым `config.json`
- Игнорирует любую существующую Docker-конфигурацию (из `--docker-config`, `WERF_DOCKER_CONFIG` или `DOCKER_CONFIG`)
- Работает совместно с автоматическим логином в CI registry (включён по умолчанию через `--login-to-registry`) при наличии CI-переменных окружения

Также можно включить через переменную окружения: `WERF_INIT_TMP_DOCKER_CONFIG=1`.

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

> Полный `.gitlab-ci.yml` для готовых рабочих процессов, а также особенности использования werf c Shell, Docker и Kubernetes executors можно найти в конфигураторе «[Быстрый старт](https://werf.io/getting_started/?usage=ci&ci=gitlabCiCd)»,
> указав в нём _CI/CD_ как сценарий использования и _GitLab CI/CD_ как CI/CD-систему, а затем выбрав интересующий вас тип CI-раннера.

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
      uses: werf/actions/install@v2

    - name: Run script
      run: |
        . $(werf ci-env github --as-file)
        werf converge
      env:
        WERF_KUBECONFIG_BASE64: ${{ secrets.KUBE_CONFIG_BASE64_DATA }}
        WERF_ENV: production
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```
{% endraw %}

> Полный набор конфигураций (`.github/workflows/*.yml`) для готовых рабочих процессов можно найти в конфигураторе «[Быстрый старт](https://werf.io/getting_started/?usage=ci&ci=githubActions)»,
> выбрав в нём _CI/CD_ как сценарий использования и _GitHub Actions_ — как CI/CD-систему.
