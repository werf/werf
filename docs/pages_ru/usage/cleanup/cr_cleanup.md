---
title: Очистка container registry
permalink: usage/cleanup/cr_cleanup.html
---

## Обзор

Количество образов может стремительно расти, занимая больше места в container registry и, соответственно, значительно увеличивая его стоимость. Для контроля и поддержания приемлемого роста места, werf предлагает свой подход к очистке, который позволяет учитывать используемые образы в Kubernetes, а также актуальность образов по истории в Git.

Команда [**werf cleanup**]({{ "reference/cli/werf_cleanup.html" | true_relative_url }}) рассчитана на периодический запуск. Удаление производится в соответствии с политиками очистки и является безопасной процедурой.

Вероятнее всего, политики очистки по умолчанию полностью покроют потребности проекта и дополнительная настройка не потребуется.

Важно отметить, что фактически размер занимаемого образами места в container registry не сокращается после очистки. werf удаляет теги неактуальных образов (манифесты), а для очистки связанных с ними данных необходимо периодически запускать сборщик мусора container registry.

> Проблематика очистки образов в container registry и наш подход к решению этой проблемы детально освещены в статье [Проблема «умной» очистки образов контейнеров и её решение в werf](https://habr.com/ru/company/flant/blog/522024/)

## Автоматизация очистки container registry

Для того, чтобы автоматизировать очистку неактуальных образов в container registry, необходимо выполнить следующие действия:

- Настроить периодический запуск [**werf cleanup**]({{ "reference/cli/werf_cleanup.html" | true_relative_url }}) для удаления неактуальных тегов из container registry.
- Настроить [периодический запуск сборщика мусора](#сборщик-мусора-container-registry) для непосредственного освобождения места в container registry.

## Игнорирование образов, используемых в Kubernetes

werf подключается **ко всем кластерам** Kubernetes, описанным **во всех контекстах** конфигурации kubectl, и собирает имена образов для следующих типов объектов: `pod`, `deployment`, `replicaset`, `statefulset`, `daemonset`, `job`, `cronjob`, `replicationcontroller`.

Пользователь может регулировать поведение следующими параметрами (и связанными переменными окружения):
- `--kube-config`, `--kube-config-base64` для определения конфигурации kubectl (по умолчанию используется пользовательская конфигурация `~/.kube/config`).
- `--kube-context` для выполнения сканирования только в определённом контексте.
- `--scan-context-namespace-only` для сканирования только связанного с контекстом namespace (по умолчанию все).

Сканирование Kubernetes отключается с помощью соответствующей директивы в `werf.yaml`:

```yaml
cleanup:
  disableKubernetesBasedPolicy: true
```

Пока в кластере Kubernetes существует объект использующий образ, он никогда не удалится из container registry. Другими словами, если что-то было запущено в вашем кластере Kubernetes, то используемые образы ни при каких условиях не будут удалены при очистке.

## Игнорирование свежесобранных образов

При удалении werf игнорирует образы, собранные в заданный период времени (по умолчанию за прошедшие 2 часа). При необходимости можно изменить период или совсем отключить политику соответствующими директивами в `werf.yaml`:

```yaml
cleanup:
  disableBuiltWithinLastNHoursPolicy: false
  keepImagesBuiltWithinLastNHours: 2
```

## Конфигурация политик очистки по истории Git

Конфигурация очистки состоит из набора политик, `keepPolicies`, по которым выполняется выборка значимых образов на основе истории git. Таким образом, в результате очистки __неудовлетворяющие политикам образы удаляются__.

Каждая политика состоит из двух частей:
- `references` определяет множество references, git-тегов или git-веток, которые будут использоваться при сканировании.
- `imagesPerReference` определяет лимит искомых образов для каждого reference из множества.

Любая политика должна быть связана с множеством Git-тегов (`tag`) либо Git-веток (`branch`). Можно указать определённое имя reference или задать специфичную группу, используя [синтаксис регулярных выражений golang](https://golang.org/pkg/regexp/syntax/#hdr-Syntax).

```yaml
tag: v1.1.1  # or /^v.*$/
branch: main # or /^(main|production)$/
```

> При сканировании описанный набор git-веток будет искаться среди origin remote references, но при написании конфигурации префикс `origin/` в названии веток опускается

Заданное множество references можно лимитировать, основываясь на времени создания git-тега или активности в git-ветке. Группа параметров `limit` позволяет писать гибкие и эффективные политики под различные workflow.

```yaml
- references:
    branch: /^features\/.*/
    limit:
      last: 10
      in: 168h
      operator: And
```

В примере описывается выборка из не более чем 10 последних веток с префиксом `features/` в имени, в которых была какая-либо активность за последнюю неделю.

- Параметр `last` позволяет выбирать последние `n` reference'ов из определённого в `branch`/`tag` множества.
- Параметр `in` (синтаксис доступен [в документации](https://golang.org/pkg/time/#ParseDuration)) позволяет выбирать Git-теги, которые были созданы в указанный период, или Git-ветки с активностью в рамках периода. Также для определённого множества `branch`/`tag`.
- Параметр `operator` определяет, какие referencе'ы будут результатом политики: удовлетворяющие оба условия или любое из них (`And` по умолчанию).

По умолчанию при сканировании reference количество искомых образов не ограничено, но поведение может настраиваться группой параметров `imagesPerReference`:

```yaml
imagesPerReference:
  last: int
  in: string
  operator: string
```

- Параметр `last` определяет количество искомых образов для каждого reference. По умолчанию количество не ограничено (`-1`).
- Параметр `in` (синтаксис доступен [в документации](https://golang.org/pkg/time/#ParseDuration)) определяет период, в рамках которого необходимо выполнять поиск образов.
- Параметр `operator` определяет, какие образы сохранятся после применения политики: удовлетворяющие оба условия или любое из них (`And` по умолчанию).

> Для Git-тегов проверяется только HEAD-коммит и значение `last` >1 не имеет никакого смысла, является невалидным

### Приоритет политик для конкретного reference

При описании группы политик необходимо идти от общего к частному. Другими словами, `imagesPerReference` для конкретного reference будет соответствовать последней политике, под которую он подпадает:

```yaml
- references:
    branch: /.*/
  imagesPerReference:
    last: 1
- references:
    branch: master
  imagesPerReference:
    last: 5
```

В данном случае, для reference _master_ справедливы обе политики и при сканировании ветки `last` будет равен 5.

### Политики по умолчанию

В случае, если в `werf.yaml` отсутствуют пользовательские политики очистки, используются политики по умолчанию, соответствующие следующей конфигурации:

```yaml
cleanup:
  keepPolicies:
  - references:
      tag: /.*/
      limit:
        last: 10
  - references:
      branch: /.*/
      limit:
        last: 10
        in: 168h
        operator: And
    imagesPerReference:
      last: 2
      in: 168h
      operator: And
  - references:
      branch: /^(main|master|staging|production)$/
    imagesPerReference:
      last: 10
```

Разберём каждую политику по отдельности:

1. Сохранять по одному образу для 10 последних тегов (по дате создания).
2. Сохранять по не более чем два образа, опубликованных за последнюю неделю, для не более 10 веток с активностью за последнюю неделю.
3. Сохранять по 10 образов для веток main, master, staging и production.

### Отключение политик

Если очистка по истории Git не требуются, то её можно отключить в `werf.yaml` с помощью специальной директивы:

```yaml
cleanup:
  disableGitHistoryBasedPolicy: true
```

## Особенности работы с различными container registries

По умолчанию при удалении тегов werf использует [_Docker Registry API_](https://docs.docker.com/registry/spec/api/) и от пользователя требуется только авторизация с использованием доступов с достаточным набором прав. Если же удаление посредством _Docker Registry API_ не поддерживается и оно реализуется в нативном API container registry, то от пользователя могут потребоваться специфичные для используемого container registry действия.

|                             |                             |
|-----------------------------|:---------------------------:|
| _AWS ECR_                   |     [***ок**](#aws-ecr)     |
| _Azure CR_                  |    [***ок**](#azure-cr)     |
| _Default_                   |           **ок**            |
| _Docker Hub_                |   [***ок**](#docker-hub)    |
| _GCR_                       |           **ок**            |
| _GitHub Packages_           | [***ок**](#github-packages) |
| _GitLab Registry_           | [***ок**](#gitlab-registry) |
| _Harbor_                    |           **ок**            |
| _JFrog Artifactory_         |           **ок**            |
| _Nexus_                     |           **ок**            |
| _Quay_                      |           **ок**            |
| _Yandex container registry_ |           **ок**            |
| _Selectel CRaaS_            | [***ок**](#selectel-craas)  |

werf пытается автоматически определить используемый container registry, используя заданный адрес репозитория (опция `--repo`). Пользователь может явно задать container registry опцией `--repo-container-registry` или переменной окружения `WERF_REPO_CONTAINER_REGISTRY`.

### AWS ECR

При удалении тегов werf использует _AWS SDK_, поэтому перед очисткой container registry необходимо выполнить **одно из** следующих действий:

- [Установить _AWS CLI_ и выполнить конфигурацию](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#cli-quick-configuration) (`aws configure`) или
- Определить `AWS_ACCESS_KEY_ID` и `AWS_SECRET_ACCESS_KEY` переменные окружения.

### Azure CR

При удалении тегов werf использует _Azure CLI_, поэтому перед очисткой container registry необходимо выполнить следующие действия:

- Установить [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest) (`az`).
- Выполнить авторизацию (`az login`).

> Пользователю необходимо иметь одну из следующих ролей: `Owner`, `Contributor` или `AcrDelete` (подробнее [Azure CR roles and permissions](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-roles))

### Docker Hub

При удалении тегов werf использует _Docker Hub API_, поэтому при очистке container registry необходимо определить _token_ или _username_ и _password_.

Для получения _token_ можно использовать следующий скрипт:

```shell
HUB_USERNAME=username
HUB_PASSWORD=password
HUB_TOKEN=$(curl -s -H "Content-Type: application/json" -X POST -d '{"username": "'${HUB_USERNAME}'", "password": "'${HUB_PASSWORD}'"}' https://hub.docker.com/v2/users/login/ | jq -r .token)
```

> В качестве _token_ нельзя использовать [personal access token](https://docs.docker.com/docker-hub/access-tokens/), т.к. удаление ресурсов возможно только при использовании основных учётных данных пользователя

Для того чтобы задать параметры, следует использовать следующие опции или соответствующие им переменные окружения:
- `--repo-docker-hub-token` или
- `--repo-docker-hub-username` и `--repo-docker-hub-password`.

### GitHub Packages

При организации CI/CD в Github Actions мы рекомендуем использовать [наш набор actions](https://github.com/werf/actions), который решит за вас большинство задач.

При удалении тегов werf использует _GitHub API_, поэтому при очистке container registry необходимо определить _token_ с `read:packages` и `delete:packages` scopes.

Для того чтобы задать токен, следует использовать опцию `--repo-github-token` или соответствующую переменную окружения.

### GitLab Registry

При удалении тегов werf использует _GitLab container registry API_ или _Docker Registry API_ в зависимости от версии GitLab.

> Для удаления тега прав временного токена CI-задания (`$CI_JOB_TOKEN`) недостаточно, поэтому пользователю необходимо создать специальный токен в разделе Access Token (в секции Scope необходимо выбрать `api`) и выполнить авторизацию с ним

### Selectel CRaaS

При очистке werf использует [_Selectel CR API_](https://developers.selectel.ru/docs/selectel-cloud-platform/craas_api/), поэтому при очистке container registry необходимо определить _username/password_, _account_ and _vpc_ or _vpcID_.

Для того чтобы задать параметры, следует использовать следующие опции или соответствующие им переменные окружения:
- `--repo-selectel-username`
- `--repo-selectel-password`
- `--repo-selectel-account`
- `--repo-selectel-vpc` or
- `--repo-selectel-vpc-id`

#### Известные проблемы

* Иногда Selectel не отдаёт токен при использовании VPC ID. Попробуйте использовать имя VPC.
* CR API не позволяет удалять теги, которые хранятся в корне registry.
* Небольшой лимит запросов в API. При активной разработке могут быть проблемы с очисткой.

## Сборщик мусора container registry

Зона ответственности очистки werf — удаление тегов образов (манифестов), а непосредственное удаление связанных данных выполняется с помощью сборщика мусора container registry (GC).

При вызове сборщика мусора container registry должен быть переведён в режим read-only или полностью отключен. Иначе есть высокая вероятность, что опубликованные во время процедуры образы не будут учтены сборщиком и будут повреждены.

Подробнее о сборщике мусора и способах его эксплуатации можно прочитать в документации используемого container registry. Например:

- [Docker Registry GC](https://docs.docker.com/registry/garbage-collection/#more-details-about-garbage-collection ).
- [GitLab CR GC](https://docs.gitlab.com/ee/administration/packages/container_registry.html#container-registry-garbage-collection).
- [Harbor GC](https://goharbor.io/docs/2.6.0/administration/garbage-collection/).
