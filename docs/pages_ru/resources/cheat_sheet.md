---
title: Шпаргалка
permalink: resources/cheat_sheet.html
toc: false
---

## Сборка и развертывание одной командой

Собрать образы и развернуть приложение в production:

```shell
werf converge --repo ghcr.io/group/project --env production
```

Собрать образы с кастомными тегами и развернуть приложение в окружение по умолчанию:

```shell
werf converge --repo ghcr.io/group/project --use-custom-tag "%image%-$CI_JOB_ID"
```

## Компоненты пайплайна

Приведенные ниже команды позволяют организовать пайплайн, адаптированный под ваши нужды.

При выполнении большинства команд будет проверено наличие собранных образов в указанном репозитории container registry. При отсутствии нужных образов, будут запущены инструкции сборки. В некоторых сценариях (например, запуск тестов в CI системе) правильнее вначале собирать образы с помощью команды `werf build`, а затем использовать те же самые образы на следующих шагах пайплайна, указывая флаг `--require-built-images`. В этом случае команда завершится с ошибкой, если нужные образы отсутствуют.

### Интеграция с CI-системой (в настоящее время поддерживаются GitLab и GitHub Workflows)

Задать значения по умолчанию для команд werf и выполнить авторизацию в container registry, используя переменные окружения GitLab:

```shell
. $(werf ci-env gitlab --as-file)
```

### Сборка, тегирование и публикация образов

Собрать образы с использованием container registry:

```shell
werf build --repo ghcr.io/group/project
```

Собрать образы и прикрепить к ним кастомные теги в дополнение к тегам на основе содержимого:

```shell
werf build --repo ghcr.io/group/project --add-custom-tag latest --add-custom-tag 1.2.1
```

Собрать и сохранить конечные образы в отдельный registry, развернутый в кластере Kubernetes:

```shell
werf build --repo ghcr.io/group/project --final-repo fast-in-cluster-registry.cluster/group/project
```

Собрать образы с использованием реестра контейнеров (можно использовать локальное хранилище) и экспортировать их в другой реестр контейнеров:

```shell
werf export --repo ghcr.io/group/project --tag ghcr.io/group/otherproject/%image%:latest
```

### Выполнение разовых задач (юнит-тесты, линтеры, разовые job'ы)

Запустить тесты с использованием предварительно созданного образа `frontend_image` в Pod'е Kubernetes:

```shell
werf kube-run frontend_image --repo ghcr.io/group/project -- npm test
```

Запустить тесты в Pod'е, но перед выполнением команды скопировать файл с секретными ENV-переменными в контейнер:

```shell
werf kube-run frontend_image --repo ghcr.io/group/project --copy-to ".env:/app/.env" -- npm run e2e-tests
```

Запустить тесты в Pod'е Kubernetes и скачать отчет о покрытии тестов из контейнера после завершения:

```shell
werf kube-run frontend_image --repo ghcr.io/group/project --copy-from "/app/report:." -- go test -coverprofile report ./...
```

Запустить команду по умолчанию собранного образа в Pod'е Kubernetes с заданными CPU requests:

```shell
werf kube-run frontend_image --repo ghcr.io/group/project --overrides='{"spec":{"containers":[{"name": "%container_name%", "resources":{"requests":{"cpu":"100m"}}}]}}'
```

### Запуск интеграционных тестов

Как правило, для запуска интеграционных тестов (e2e, acceptance, security и т.д.) необходимо production-окружение (его можно подготовить с помощью `converge` или `bundle`) и контейнер с соответствующей командой.

Запустить интеграционные тесты с помощью `converge`:

```shell
werf converge --repo ghcr.io/group/project --env integration
```

Запустить интеграционные тесты, подготовив окружение с помощью `converge`, и выполнить разовую задачу с помощью `kube-run`:

```shell
werf converge --repo ghcr.io/group/project --env integration_infra
werf kube-run --repo ghcr.io/group/project --env integration -- npm run acceptance-tests
```

### Подготовка артефактов релиза (по желанию)

Бандлы werf позволяют подготовить артефакты релиза, которые могут быть протестированы или развернуты позже (с помощью werf, Argo CD или Helm), и сохранить их в реестре контейнеров с указанным тегом.

Использовать тег semver, совместимый с OCI-чартом Helm:

```shell
werf bundle publish --repo ghcr.io/group/project --tag 1.0.0
```

Использовать произвольный символьный тег:

```shell
werf bundle publish --repo ghcr.io/group/project --tag latest
```

### Развертывание приложения

Собрать и развернуть приложение в production:

```shell
werf converge --repo ghcr.io/group/project --env production
```

Развернуть приложение, собранное на предыдущем шаге, и использовать кастомный тег вместо тега по умолчанию на основе содержимого:

```shell
werf converge --require-built-images --repo ghcr.io/group/project --use-custom-tag "%image%-$CI_JOB_ID"
```

Развернуть ранее опубликованный бандл с тегом 1.0.0 в production:

```shell
werf bundle apply --repo ghcr.io/group/project --env production --tag 1.0.0
```

### Очистка реестра контейнеров

> Процедура должна запускаться по расписанию. Иначе количество образов и метаданных werf может значительно увеличить занимаемый размер реестра и время выполнения операций

Выполнить безопасную процедуру очистки неактуальных образов и метаданных werf из реестра контейнеров с учётом пользовательских политик очистки и запущенных в K8s-кластере образов:

```shell
werf cleanup --repo ghcr.io/group/project
```

## Локальная разработка

У большинства команд werf имеется флаг `--dev` для локальной разработки. Он позволяет выполнять команды, не добавляя (`git add`) их в Git. Флаг `--follow` перезапускает команду при изменении файлов в репозитории.

Отрендерить и показать манифесты:

```shell
werf render --dev
```

Собрать образ и запустить интерактивную оболочку в контейнере с неудавшейся стадией в случае ошибки:

```shell
werf build --dev [--follow] --introspect-error
```

Собрать образ, запустить его в Pod'е Kubernetes и выполнить в нем команду:

```shell
werf kube-run --dev [--follow] --repo ghcr.io/group/project frontend -- npm lint
```

Запустить командную оболочку в контейнере в Pod'e Kubernetes для указанного образа:

```shell
werf kube-run --dev --repo ghcr.io/group/project -it frontend -- bash
```

Собрать образ и развернуть его в dev-кластере (может быть локальным):

```shell
werf converge --dev [--follow] --repo ghcr.io/group/project
```

Собрать образ и развернуть его в dev-кластере; использовать стадии из вторичного read-only-реестра для ускорения сборки:

```shell
werf converge --dev [--follow] --repo ghcr.io/group/project --secondary-repo ghcr.io/group/otherproject
```

Выполнить команду `docker-compose up` с переданными именами образов:

```shell
werf compose up --dev [--follow]
```
