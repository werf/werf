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

## Конфигурация Docker

По умолчанию `werf ci-env` копирует текущую директорию конфигурации Docker (из `~/.docker` или пути, указанного через `--docker-config`/`WERF_DOCKER_CONFIG`) во временную директорию и экспортирует переменную окружения `DOCKER_CONFIG`, указывающую на неё. Это изолирует Docker-операции CI-задания от конфигурации хоста.

Если задана переменная окружения `DOCKER_AUTH_CONFIG`, `werf ci-env` автоматически использует её: вместо копирования существующей конфигурации Docker создаётся новая на основе её содержимого. Это полезно, когда:

- CI-раннер не имеет постоянной конфигурации Docker (например, эфемерные раннеры).
- Учётные данные реестра передаются через `DOCKER_AUTH_CONFIG` (распространено в GitLab CI/CD).
- Требуется чистая конфигурация Docker без унаследованных credential helpers или настроек.

Значение `DOCKER_AUTH_CONFIG` должно быть валидной JSON-строкой в формате Docker config, например:

```json
{"auths": {"registry.example.com": {"auth": "base64-encoded-user:password"}}}
```

Это поведение можно явно контролировать флагом `--use-docker-auth-config`/`--use-docker-auth-config=false` (или переменной `WERF_USE_DOCKER_AUTH_CONFIG`). Явное значение всегда имеет приоритет над автоопределением: например, `--use-docker-auth-config=false` заставит скопировать существующую конфигурацию Docker, даже если `DOCKER_AUTH_CONFIG` задан.

Если флаг `--use-docker-auth-config` явно включён, но `DOCKER_AUTH_CONFIG` не задан, `werf ci-env` завершится с ошибкой.

> После создания временной конфигурации, если включён флаг `--login-to-registry` (по умолчанию включен), `werf ci-env` дополнительно авторизуется в container registry CI, добавляя учётные данные к временной конфигурации.

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

Если CI-раннер предоставляет учётные данные реестра через переменную окружения `DOCKER_AUTH_CONFIG`, можно использовать `--use-docker-auth-config` для создания конфигурации Docker на её основе:

```yaml
Deploy to Production:
  stage: deploy
  script:
    - . $(werf ci-env gitlab --as-file --use-docker-auth-config)
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
