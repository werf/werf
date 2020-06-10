---
title: Интеграция с GitHub Actions 
sidebar: documentation
permalink: documentation/guides/github_ci_cd_integration.html
author: Sergey Lazarev <sergey.lazarev@flant.com>, Alexey Igrychev <alexey.igrychev@flant.com>
---

## Обзор задачи

В статье рассматриваются различные варианты настройки CI/CD с использованием GitHub CI/CD и werf.

Рабочий процесс в репозитории (набор GitHub workflow конфигураций) будет строиться на базе следующих заданий:
* `converge` — задание сборки, публикации образов и выката приложения для одного из контуров кластера;
* `dismiss` — задание удаления приложения (используется только для review окружений);
* `cleanup` — задание очистки хранилища стадий и Docker registry.

Ключевыми шагами заданий будут наши комплексные GitHub Actions, [werf/actions](https://github.com/werf/actions), которые сочетают в себе все необходимые шаги по подготовке окружения и выполнения требуемых werf команд. В статье будут рассмотрены примеры использования большинства из них, больше подробностей можно найти в репозитории набора.

Набор контуров в кластере Kubernetes может варьироваться в зависимости от многих факторов.
В статье будут приведены различные варианты организации окружений для следующих:
* [Контур production]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#production).
* [Контур staging]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#staging).
* [Контур review]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#review).

Далее последовательно рассматриваются задания и различные варианты их организации. Изложение построено от общего к частному. В конце статьи приведён [полный набор конфигураций для готовых workflow](#полный-набор-конфигураций-для-готовых-workflow).

Независимо от workflow, все версии конфигураций подчиняются следующим правилам:
* Сборка и публикация является неотъемлемой частью выката.
* [*Выкат/удаление* review окружений](#варианты-организации-review-окружения):
  * *Выкат* на review окружение возможен только в рамках Pull Request (PR).
  * Review окружения удаляются автоматически при закрытии PR.
* [*Очистка*](#очистка-образов) запускается один раз в день по расписанию на master.

Для выкатов review окружения, а также staging и production окружений предложены самые популярные варианты по организации. Каждый вариант для staging и production окружений сопровождается всевозможными способами отката релиза в production.

> С общей информацией по организации CI/CD с помощью werf, а также информацией по конструированию своего workflow, можно ознакомиться в [общей статье]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html)

## Требования

* Кластер Kubernetes.
* Проект на [GitHub](https://github.com/).
* Приложение, которое успешно собирается и деплоится с помощью werf.
* Понимание [основных концептов GitHub Actions](https://help.github.com/en/actions/getting-started-with-github-actions/core-concepts-for-github-actions).
  
> Далее в примерах статьи будут использоваться виртуальные машины, предоставляемые GitHub, с OS Linux (`runs-on: ubuntu-latest`). Тем не менее, все примеры также справедливы для предустановленных self-hosted runners на базе любой OS

## Сборка, публикация образов и выкат приложения

Прежде всего необходимо описать некий шаблон задания, общую часть для выката в любой контур, что позволит сосредоточиться далее на правилах выката и предложенных workflow.

{% include /guides/github_ci_cd_integration/converge_base.md %}

> Данное задание можно разбить на два независимых, но в нашем случае (сборка и публикации вызывается не на каждый коммит, а используется только совместно с выкатом) это избыточно и ухудшит читаемость конфигурации и время выполнения.
>
{% include /guides/github_ci_cd_integration/build_and_publish_note.md %}

Первый шаг, с которого начинается задание — `Checkout code`, добавление исходных кодов приложения. При использовании сборщика werf (основная особенность которого — инкрементальная сборка) недостаточно, так называемого, `shallow clone` с единственным коммитом, который создаёт action `actions/checkout@v2` при использовании без параметров. 

Базируясь на истории git, werf создаёт стадии. При отсутствии истории каждая сборка будет проходить без ранее собранных стадий, поэтому, крайне важно, использовать параметр `fetch-depth: 0` для доступа ко всей истории для всех команд, которые используют стадии: при сборке и публикации (`werf build-and-publish`), выкате (`werf deploy`), запуске (`werf run`) и т.д. 

{% raw %}
```yaml
- name: Checkout code
  uses: actions/checkout@v2
  with:
    fetch-depth: 0
```
{% endraw %}

Следующим шагом используется action `werf/actions/converge`, который объединяет все необходимые шаги, подготавливает окружение и вызывает соответствующую команду.  

{% raw %}
```yaml
- name: Converge
  uses: werf/actions/converge@master
  with:
    env: ANY_ENV_NAME
    kube-config-base64-data: ${{ secrets.KUBE_CONFIG_BASE64_DATA }}
  env:
    WERF_SET_ENV_URL: "global.env_url=ANY_ENV_URL"
```
{% endraw %}

Среди предустановленного ПО на виртуальных машинах GitHub уже установлен `kubectl`, поэтому пользователю остаётся только:
* определиться с конфигурацией [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/);
* создать секретную переменную `KUBE_CONFIG_BASE64_DATA` с контентом файла kubeconfig (`cat ~/.kube/config | base64`) в Settings/Secrets проекта на GitHub.
* прокинуть секретную переменную `secrets.KUBE_CONFIG_BASE64_DATA` в используемый werf action:
    
{% raw %}
```yaml
- name: Converge
  uses: werf/actions/converge@master
  with:
    kube-config-base64-data: ${{ secrets.KUBE_CONFIG_BASE64_DATA }}
```
{% endraw %}

Конфигурация задания достаточно проста, поэтому хочется сделать акцент на том, чего в ней нет — явной авторизации в Docker registry, вызова `docker login`. 

В простейшем случае, при использовании встроенной [Github Packages имплементации Docker registry](https://help.github.com/en/packages/using-github-packages-with-your-projects-ecosystem/configuring-docker-for-use-with-github-packages), авторизация выполняется автоматически при вызове команды `werf ci-env`. В качестве необходимых аргументов используются переменные окружения GitHub, [секретная переменная `GITHUB_TOKEN`](https://help.github.com/en/actions/configuring-and-managing-workflows/authenticating-with-the-github_token#about-the-github_token-secret), а также имя пользователя (`GITHUB_ACTOR`) инициировавшего запуск workflow.

Если необходимо выполнить авторизацию с произвольными учётными данными или с внешним Docker registry, то необходимо использовать готовый action для вашей имплементации или просто выполнить `docker login` перед использованием werf. 

Рассмотрим оставшиеся используемые параметры на этом шаге:

{% raw %}
```yaml
- name: Converge
  uses: werf/actions/converge@master
  with:
    env: ANY_ENV_NAME
  env:
    WERF_SET_ENV_URL: "global.env_url=ANY_ENV_URL"
```
{% endraw %}

Для каждого контура необходимо определить окружение. В нашем случае оно определяется следующими параметрами: 
* именем (`ANY_ENV_NAME`) и; 
* URL (`ANY_ENV_URL`).             

Для того, чтобы по-разному конфигурировать приложение для используемых контуров кластера в helm-шаблонах можно использовать Go-шаблоны и переменную `.Values.global.env`, что соответствует значению, которое задаётся в качестве параметра у action (`env`).

> Адрес окружения является необязательным. В данной статье используется исключительно в качестве примера организации окружений и для демонстрации работы с опциями werf при использовании actions (все опции werf можно задавать через переменные окружения) 

Адрес окружения, URL для доступа к разворачиваемому в контуре приложению, который передаётся параметром `global.env_url`, может использоваться в helm-шаблонах, например, для конфигурации Ingress-ресурсов. Для того, чтобы опредилить URL используется переменная окружения `WERF_SET_ENV_URL`, которая соответствует вызову werf с опцией `--set` (`WERF_SET_<ANY_NAME>`).

> Если для шифрования значений переменных вы используете werf, то вам также необходимо добавить `encryption key` в переменную `WERF_SECRET_KEY` в Settings/Secrets проекта и добавить секрет в секцию env

Далее будут представлены популярные стратегии и практики, на базе которых мы предлагаем выстраивать ваши процессы в GitHub Actions. 

### Варианты организации review окружения

Как уже было сказано ранее, review окружение — это динамический контур, поэтому помимо выката, у этого окружения также будет и очистка.

Рассмотрим базовые GitHub workflow файлы, которые лягут в основу всех предложенных вариантов.

Сначала разберём файл `.github/workflows/review_deployment.yml`.

{% include /guides/github_ci_cd_integration/review_base.md %}

В нём пропущено условие запуска, т.к. оно зависит от выбранного варианта организации.

От базовой конфигурации задание отличается только появившимся шагом `Define environment url`. 
На этом шаге генерируется уникальный URL для каждого PR, по которому после выката будет доступно наше приложение (при соответствующей организации helm templates). 

{% raw %}
```yaml
- name: Define environment url
  run: |
    pr_id=${{ github.event.number }}
    github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
    echo ::set-env name=WERF_SET_ENV_URL::global.env_url=http://${github_repository_id}-${pr_id}.kube.DOMAIN
```
{% endraw %}

Далее файл `.github/workflows/review_deployment_dismiss.yml`.

{% include /guides/github_ci_cd_integration/review_dismiss_base.md %}

Данный GitHub workflow будет выполняться при закрытии PR.

```yaml
on:
  pull_request:
    types: [closed]
```

На шаге `Dismiss` выполняется удаление review-релиза: werf удаляет helm-релиз и namespace в Kubernetes со всем его содержимым ([werf dismiss]({{ site.baseurl }}/documentation/cli/main/dismiss.html)).

Далее разберём основные стратегии при организации выката review окружения. 

> Мы не ограничиваем вас предложенными вариантами, даже напротив — рекомендуем комбинировать их и создавать конфигурацию workflow под нужды вашей команды

#### №1 Вручную

> Данный вариант реализует подход описанный в разделе [Выкат на review из pull request по кнопке]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#выкат-на-review-из-pull-request-по-кнопке)

При таком подходе пользователь выкатывает и удаляет окружение, проставляя соответствующий лейбл (`review_start` или `review_stop`) в PR.

Он самый простой и может быть удобен в случае, когда выкаты происходят редко и review окружение не используется при разработке.
По сути, для проверки перед принятием PR. 

{% include /guides/github_ci_cd_integration/review_1.md %}

В данном варианте оба GitHub workflow ожидают проставление лейбла в PR. 

```yaml
on:
  pull_request:
    types: [labeled]
```

Если событие связано с добавлением лейбла `review_start` или `review_stop`, то выполняются задания соответствующего workflow. 
Иначе, при проставлении произвольного лейбла — workflow запускается, но ни одно задание не выполняется и он помечается как `skipped`.
Используя фильтрацию по статусу, можно проследить активность в review окружении.

Шаг `Label taking off` снимает лейбл, который инициирует запуск workflow. Он используется в качестве индикатора обработки пользовательского запроса на выкат и остановку review окружения (а бонусом, мы можем отслеживать историю изменений и выкатов по логу в PR).  

{% raw %}
```yaml
labels:
  name: Label taking off 
  runs-on: ubuntu-latest
  if: github.event.label.name == 'review_stop'
  steps:

    - name: Take off label
      uses: actions/github-script@v1
      with:
        script: "github.issues.removeLabel({...context.issue, name: github.event.label.name})"
```
{% endraw %}

#### №2 Автоматически по имени ветки

> Данный вариант реализует подход описанный в разделе [Выкат на review из ветки по шаблону автоматически]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#выкат-на-review-из-ветки-по-шаблону-автоматически)

В предложенном ниже варианте автоматический релиз выполняется для каждого коммита в PR, в случае, если имя git-ветки содержит `review`. 

{% include /guides/github_ci_cd_integration/review_2.md %}

Выкат инициируется при коммите в ветку, открытии и переоткрытии PR, что соответствует набору событий по умолчанию для `pull_request`:
```yaml
// equal conditions

on:
  pull_request:
    types:
      - opened
      - reopened
      - synchronize

on:
  pull_request:
```

#### №3 Полуавтоматический режим с лейблом (рекомендованный)

> Данный вариант реализует подход описанный в разделе [Выкат на review из pull request автоматически после ручной активации]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#выкат-на-review-из-pull-request-автоматически-после-ручной-активации)

Полуавтоматический режим с лейблом — это комплексное решение, объединяющие первые два варианта. 

При проставлении специального лейбла, в примере ниже `review`, пользователь активирует автоматический выкат в review окружения для каждого коммита. 
При снятии лейбла автоматически запускается удаление review-релиза.    

{% include /guides/github_ci_cd_integration/review_3.md %}

Выкат инициируется при коммите в ветку, добавлении и снятии лейбла в PR, что соответствует следующему набору событий для `pull_request`:

```yaml
pull_request:
  types:
    - labeled
    - unlabeled
    - synchronize
```

### Варианты организации staging и production окружений

Предложенные далее варианты являются наиболее эффективными комбинациями правил выката **staging** и **production** окружений.

В нашем случае, данные окружения являются определяющими, поэтому названия вариантов соответствуют названиям окончательных workflow, предложенных в [конце статьи](#полный-набор-конфигураций-для-готовых-workflow). 

#### №1 Fast and Furious (рекомендованный)

> Данный вариант реализует подходы описанные в разделах [Выкат на production из master автоматически]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#выкат-на-production-из-master-автоматически) и [Выкат на production-like из pull request по кнопке]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#выкат-на-production-like-из-pull-request-по-кнопке)

Выкат в **production** происходит автоматически при любых изменениях в master. Выполнить выкат в **staging** можно только проставив соответствующий лейбл в PR.

{% include /guides/github_ci_cd_integration/production_staging_1.md %}

Варианты отката изменений в production:
- [revert изменений](https://git-scm.com/docs/git-revert) в master (**рекомендованный**);
- выкат стабильного PR.

#### №2 Push the Button (*)

> Данный вариант реализует подходы описанные в разделах [Выкат на production из master по кнопке]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#выкат-на-production-из-master-по-кнопке) и [Выкат на staging из master автоматически]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#выкат-на-staging-из-master-автоматически)

{% include /guides/github_ci_cd_integration/not_recommended_approach_ru.md %}

Выкат **production** осуществляется вручную на master, а выкат в **staging** происходит автоматически при любых изменениях в master.

{% include /guides/github_ci_cd_integration/production_staging_2.md %}

Имеем следующее условие для релиза в production:

```yaml
on:
  repository_dispatch:
    types: [production_deployment]
```

Чтобы запустить данный workflow, достаточно выполнить следующий запрос:

```shell
curl \
  --location --request POST 'https://api.github.com/repos/<company>/<project>/dispatches' \
  --header 'Content-Type: application/json' \
  --header 'Accept: application/vnd.github.everest-preview+json' \
  --header "Authorization: token $GITHUB_TOKEN" \
  --data-raw '{
    "event_type": "production_deployment",
    "client_payload": {}
  }'
```

Для использования данного подхода можно добавить скрипт в репозиторий проекта, использовать postman, плагин для браузера и т.д.

Варианты отката изменений в production:
- выкат стабильного PR и инициализация выката при помощи repository_dispatch.

#### №3 Tag everything (*)

> Данный вариант реализует подходы описанные в разделах [Выкат на production из тега автоматически]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#выкат-на-production-из-тега-автоматически) и [Выкат на staging из master по кнопке]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#выкат-на-staging-из-master-по-кнопке)

{% include /guides/github_ci_cd_integration/not_recommended_approach_ru.md %}

Выкат в **production** выполняется при проставлении тега, а в **staging** осуществляется вручную на master.

{% include /guides/github_ci_cd_integration/production_staging_3.md %}

```yaml
on:
  repository_dispatch:
    types: [staging_deployment]
```

Чтобы запустить данный workflow, достаточно выполнить следующий запрос:

```shell
curl \
  --location --request POST 'https://api.github.com/repos/<company>/<project>/dispatches' \
  --header 'Content-Type: application/json' \
  --header 'Accept: application/vnd.github.everest-preview+json' \
  --header "Authorization: token $GITHUB_TOKEN" \
  --data-raw '{
    "event_type": "staging_deployment",
    "client_payload": {}
  }'
```

Для использования данного подхода можно добавить скрипт в репозиторий проекта, использовать postman, плагин для браузера и т.д.

Варианты отката изменений в production:
- создание нового тега на старый коммит.

#### №4 Branch, branch, branch!

> Данный вариант реализует подходы описанные в разделах [Выкат на production из ветки автоматически]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#выкат-на-production-из-ветки-автоматически) и [Выкат на production-like из ветки автоматически]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#выкат-на-production-like-из-ветки-автоматически)

Выкат в **production** происходит автоматически при любых изменениях в ветке production, а в **staging** при любых изменениях в ветке master.

{% include /guides/github_ci_cd_integration/production_staging_4.md %}

Варианты отката изменений в production:
- [revert изменений](https://git-scm.com/docs/git-revert) в ветке production;
- [revert изменений](https://git-scm.com/docs/git-revert) в master и fast-forward merge в ветку production;
- удаление коммита из ветки production и push-force.

## Очистка образов

{% include /guides/github_ci_cd_integration/cleanup_base.md %}

Первый шаг, с которого начинается задание — `Checkout code`, добавление исходных кодов приложения. 

Большинство политик очистки в werf базируется на примитивах git (на коммите, ветке и теге), поэтому использование action `actions/checkout@v2` без дополнительных параметров и действий может приводить к неожиданному удалению образов. Мы рекомендуем использовать следующие шаги для корректной работы. 

```yaml
- name: Checkout code
  uses: actions/checkout@v2
  
- name: Fetch all history for all tags and branches
  run: git fetch --prune --unshallow
```

В werf встроен эффективный механизм очистки, который позволяет избежать переполнения Docker registry и диска сборочного узла от устаревших и неиспользуемых образов.
Более подробно ознакомиться с функционалом очистки, встроенным в werf, можно [здесь]({{ site.baseurl }}/documentation/reference/cleaning_process.html).

## Полный набор конфигураций для готовых workflow

<div class="tabs" style="display: grid">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_github_ci_1')">№1 Fast and Furious (рекомендованный)</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_github_ci_2')">№2 Push the Button (*)</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_github_ci_3')">№3 Tag everything (*)</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_github_ci_4')">№4 Branch, branch, branch!</a>
</div>

<div id="complete_github_ci_1" class="tabs__content no_toc_section active" markdown="1">

### Детали workflow
{:.no_toc}

> Подробнее про workflow можно почитать в отдельной [статье]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#1-fast-and-furious)

* Выкат на review контур по стратегии [№3 Полуавтоматический режим с лейблом (рекомендованный)](#3-полуавтоматический-режим-с-лейблом-рекомендованный).
* Выкат на staging и production контуры осуществляется по стратегии [№1 Fast and Furious (рекомендованный)](#1-fast-and-furious-рекомендованный).
* [Очистка стадий](#очистка-образов) выполняется по расписанию раз в сутки.

### Конфигурации
{:.no_toc}

{% include /guides/github_ci_cd_integration/workflow_1.md %}

</div>

<div id="complete_github_ci_2" class="tabs__content no_toc_section" markdown="1">

### Детали workflow
{:.no_toc}

> Подробнее про workflow можно почитать в отдельной [статье]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#2-push-the-button)

{% include /guides/github_ci_cd_integration/not_recommended_approach_ru.md %}

* Выкат на review контур по стратегии [№1 Вручную](#1-вручную).
* Выкат на staging и production контуры осуществляется по стратегии [№2 Push the Button](#2-push-the-button-).
* [Очистка стадий](#очистка-образов) выполняется по расписанию раз в сутки.

### Конфигурации
{:.no_toc}

{% include /guides/github_ci_cd_integration/workflow_2.md %}

</div>

<div id="complete_github_ci_3" class="tabs__content no_toc_section" markdown="1">

### Детали workflow
{:.no_toc}

> Подробнее про workflow можно почитать в отдельной [статье]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#3-tag-everything)

{% include /guides/github_ci_cd_integration/not_recommended_approach_ru.md %}

* Выкат на review контур по стратегии [№1 Вручную](#1-вручную).
* Выкат на staging и production контуры осуществляется по стратегии [№3 Tag everything](#3-tag-everything-).
* [Очистка стадий](#очистка-образов) выполняется по расписанию раз в сутки. 

### Конфигурации
{:.no_toc}

{% include /guides/github_ci_cd_integration/workflow_3.md %}

</div>

<div id="complete_github_ci_4" class="tabs__content no_toc_section" markdown="1">

### Детали workflow
{:.no_toc}

> Подробнее про workflow можно почитать в отдельной [статье]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#4-branch-branch-branch)

* Выкат на review контур по стратегии [№2 Автоматически по имени ветки](#2-автоматически-по-имени-ветки).
* Выкат на staging и production контуры осуществляется по стратегии [№4 Branch, branch, branch!](#4-branch-branch-branch).
* [Очистка стадий](#очистка-образов) выполняется по расписанию раз в сутки.

### Конфигурации
{:.no_toc}

{% include /guides/github_ci_cd_integration/workflow_4.md %}

</div>
