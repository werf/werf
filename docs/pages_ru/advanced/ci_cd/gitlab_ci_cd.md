---
title: Интеграция с GitLab CI/CD
permalink: advanced/ci_cd/gitlab_ci_cd.html
author: Artem Kladov <artem.kladov@flant.com>, Alexey Igrychev <alexey.igrychev@flant.com>
---

## Обзор задачи

В статье рассматриваются различные варианты настройки CI/CD с использованием GitLab CI/CD и werf.

Конечный pipeline состоит из следующего набора стадий:
* `build` — стадия сборки и публикации образов приложения;
* `deploy` — стадия деплоя приложения для одного из контуров кластера;
* `dismiss` — стадия удаления приложения для review окружения;
* `cleanup` — стадия очистки хранилища стадий и container registry.

Набор контуров (а равно — окружений GitLab) в кластере Kubernetes может варьироваться в зависимости от многих факторов.
В статье будут приведены различные варианты организации окружений для следующих:
* [Контур production]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#production" | true_relative_url }}).
* [Контур staging]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#staging" | true_relative_url }}).
* [Контур review]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#review" | true_relative_url }}).

Далее последовательно рассматриваются стадии pipeline и различные варианты их организации. Изложение построено от общего к частному. В конце статьи приведены [окончательные версии `.gitlab-ci.yml`](#полный-gitlab-ciyml-для-готовых-workflow) для готовых workflow.

Независимо от workflow, все версии конфигурации подчиняются следующим правилам:
* [*Сборка и публикация*](#сборка-и-публикация-образов-приложения) выполняется при каждом push в репозитории.
* [*Выкат/удаление* review окружений](#варианты-организации-review-окружения):
  * *Выкат* на review окружение возможен только в рамках Merge Request (MR).
  * Review окружения удаляются с помощью инструментария GitHub (по кнопке в разделе Environment), автоматически при удалении ветки или отсутствии активности в MR в течение суток.
* [*Очистка*](#очистка-образов) запускается один раз в день по расписанию на master.

Для выкатов review окружения и staging и production окружений предложены самые популярные варианты по организации. Каждый вариант для staging и production окружений сопровождается всевозможными способами отката релиза в production.

> С общей информацией по организации CI/CD с помощью werf, а также информацией по конструированию своего workflow, можно ознакомиться в [общей статье]({{ "advanced/ci_cd/ci_cd_workflow_basics.html" | true_relative_url }})

## Требования

* Кластер Kubernetes и настроенный для работы с ним `kubectl`.
* GitLab-сервер версии выше 10.x либо учетная запись на [gitlab.com](https://gitlab.com/).
* Container registry, встроенный в GitLab или выделенный.
* Приложение, которое успешно собирается и деплоится с werf.
* Понимание [основных концептов GitLab CI/CD](https://docs.gitlab.com/ee/ci/).

## Инфраструктура

![scheme]({{ assets["howto_gitlabci_scheme.png"].digest_path | true_relative_url }})

* Кластер Kubernetes.
* GitLab со встроенным container registry.
* Узел или группа узлов, с предустановленным werf и зависимостями.

Процесс деплоя требует наличия доступа к кластеру через `kubectl`, поэтому необходимо установить и настроить `kubectl` на узле, с которого будет запускаться werf.
Если не указывать конкретный контекст опцией `--kube-context` или переменной окружения `WERF_KUBE_CONTEXT`, то werf будет использовать контекст `kubectl` по умолчанию.

В конечном счете werf требует наличия доступа на используемых узлах:
- к Git-репозиторию кода приложения;
- к container registry;
- к кластеру Kubernetes.

### Настройка runner

На узле, где предполагается запуск werf, установим и настроим GitLab-runner:
1. Создадим проект в GitLab и добавим push кода приложения.
1. Получим токен регистрации GitLab-runner:
   * заходим в проекте в GitLab `Settings` —> `CI/CD`;
   * во вкладке `Runners` необходимый токен находится в секции `Setup a specific Runner manually`.
1. Установим GitLab-runner [по инструкции](https://docs.gitlab.com/runner/install/linux-manually.html).
1. Зарегистрируем `gitlab-runner`, выполнив [шаги](https://docs.gitlab.com/runner/register/index.html) за исключением следующих моментов:
   * используем `werf` в качестве тега runner;
   * используем `shell` в качестве executor для runner.
1. Добавим пользователя `gitlab-runner` в группу `docker`.

   ```shell
   sudo usermod -aG docker gitlab-runner
   ```

1. [Установим werf](/installation.html).
1. Установим `kubectl` и скопируем файл конфигурации `kubectl` в домашнюю папку пользователя `gitlab-runner`.
   ```shell
   mkdir -p /home/gitlab-runner/.kube &&
   sudo cp -i /etc/kubernetes/admin.conf /home/gitlab-runner/.kube/config &&
   sudo chown -R gitlab-runner:gitlab-runner /home/gitlab-runner/.kube
   ```

После того, как GitLab-runner настроен, можно переходить к настройке pipeline.

## Активация werf

Определим команды, которые будут выполняться перед каждым заданием.

{% raw %}
```yaml
before_script:
  - type trdl && . $(trdl use werf 1.2 stable)
  - type werf && source $(werf ci-env gitlab --as-file)
```
{% endraw %}

## Сборка и публикация образов приложения

{% raw %}
```yaml
Build and Publish:
  stage: build
  script:
    - werf build
  except: [schedules]
  tags: [werf]
```
{% endraw %}

Забегая вперед, очистка хранилища стадий и container registry предполагает запуск соответствующего задания по расписанию.
Так как при очистке не требуется выполнять сборку образов, то указываем `except: [schedules]`, чтобы стадия сборки не запускалась в случае работы pipeline по расписанию.

Конфигурация задания достаточно проста, поэтому хочется сделать акцент на том, чего в ней нет — явной авторизации в container registry, вызова `docker login`.

В простейшем случае, при использовании встроенного container registry, авторизация выполняется автоматически при вызове команды `werf ci-env`. В качестве необходимых аргументов используются переменные окружения GitLab `CI_JOB_TOKEN` (подробнее про модель разграничения доступа при выполнении заданий в GitLab можно прочитать [здесь](https://docs.gitlab.com/ee/user/permissions.html)) и `CI_REGISTRY_IMAGE`.

При вызове `werf ci-env` создаётся временный docker config, который используется всеми командами в shell-сессии (в том числе docker). Таким образов, параллельные задания никак не пересекаются при использовании docker и временный токен в конфигурации не перетирается.

Если необходимо выполнить авторизацию с произвольными учётными данными, `docker login` должен выполняться после вызова `werf ci-env` (подробнее про авторизацию [в отдельной статье]({{ "advanced/supported_container_registries.html#авторизация" | true_relative_url }})).

## Выкат приложения

Прежде всего необходимо описать шаблон, общую часть для деплоя в любой контур, что позволит уменьшить размер файла `.gitlab-ci.yml`, улучшит его читаемость, а также позволит далее сосредоточиться на workflow.

{% raw %}
```yaml
.base_deploy:
  stage: deploy
  script:
    - werf converge --skip-build --set "env_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  dependencies:
    - Build and Publish
  except: [schedules]
  tags: [werf]
```
{% endraw %}

При использовании шаблона `base_deploy` для каждого контура будет определяться своё окружение GitLab:

{% raw %}
```yaml
Example:
  extends: .base_deploy
  environment:
    name: <environment name>
    url: <url>
    ...
  ...
```
{% endraw %}

При выполнении задания, `werf ci-env` устанавливает переменную `WERF_ENV` в соответствии с именем окружения GitLab (`CI_ENVIRONMENT_SLUG`).

Для того, чтобы по-разному конфигурировать приложение для используемых контуров кластера в helm-шаблонах можно использовать Go-шаблоны и переменную `.Values.werf.env`, что соответствует значению опции `--env` или переменной окружения `WERF_ENV`.

Также в шаблоне используется адрес окружения, URL для доступа к разворачиваемому в контуре приложению, который передаётся параметром `envUrl`.
Это значение может использоваться в helm-шаблонах, например, для конфигурации Ingress-ресурсов.

Далее будут представлены популярные стратегии и практики, на базе которых мы предлагаем выстраивать ваши процессы в GitLab.

### Варианты организации review окружения

Как уже было сказано ранее, review окружение — это временный контур, поэтому помимо выката, у этого окружения также должна быть и очистка.

Рассмотрим базовые конфигурации `Review` и `Stop Review` заданий, которые лягут в основу всех предложенных вариантов.

{% raw %}
```yaml
Review:
  extends: .base_deploy
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    url: http://${CI_PROJECT_NAME}-${CI_MERGE_REQUEST_ID}.kube.DOMAIN
    on_stop: Stop Review
    auto_stop_in: 1 day
  artifacts:
    paths:
      - werf.yaml
  only: [merge_requests]

Stop Review:
  stage: dismiss
  script:
    - werf dismiss --with-namespace
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]
```
{% endraw %}

В задание `Review` описывается выкат review-релиза в динамическое окружение, в основу имени которого закладывается уникальный идентификатор MR. Параметр `auto_stop_in` позволяет указать период отсутствия активности в MR, после которого окружение GitLab будет автоматически остановлено. Остановка окружения GitLab сама по себе никак не влияет на ресурсы в кластере, review-релиз, поэтому в дополнении необходимо определить задание, которое вызывается при остановке (`on_stop`). В нашем случае, это задание `Stop Review`.

Задание `Stop Review` выполняет удаление review-релиза, а также остановку окружения GitLab (`action: stop`): werf удаляет helm-релиз и namespace в Kubernetes со всем его содержимым ([werf dismiss]({{ "reference/cli/werf_dismiss.html" | true_relative_url }})). Задание `Stop Review` может быть запущено вручную после деплоя на review контур, а также автоматически GitLab-сервером, например, при удалении соответствующей ветки в результате слияния ветки с master и указания соответствующей опции в интерфейсе GitLab.

Для выполнения `werf dismiss` требуется werf.yaml, так как в нём содержаться [шаблоны имени релиза и namespace]({{ "/advanced/helm/releases/naming.html" | true_relative_url }}). При удалении ветки нет возможности использовать исходники из git, поэтому в задании `Stop Review` используется werf.yaml, сохранённый при выполнении задания `Review`, и отключено подтягивание изменений из git (`GIT_STRATEGY: none`).

Таким образом, по умолчанию закладываем следующие варианты удаления review окружения:
* вручную;
* автоматически при отсутствии активности MR в течение суток и при удалении ветки.

Далее разберём основные стратегии при организации выката review окружения.

> Мы не ограничиваем вас предложенными вариантами, даже напротив — рекомендуем комбинировать их и создавать конфигурацию workflow под нужды вашей команды

#### №1 Вручную

> Данный вариант реализует подход описанный в разделе [Выкат на review из pull request по кнопке]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#выкат-на-review-из-pull-request-по-кнопке" | true_relative_url }})

При таком подходе пользователь выкатывает и удаляет окружение по кнопке в pipeline.

Он самый простой и может быть удобен в случае, когда выкаты происходят редко и review окружение не используется при разработке.
По сути, для проверки перед принятием MR.

{% raw %}
```yaml
Review:
  extends: .base_deploy
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    url: http://${CI_PROJECT_NAME}-${CI_MERGE_REQUEST_ID}.kube.DOMAIN
    on_stop: Stop Review
    auto_stop_in: 1 day
  artifacts:
    paths:
      - werf.yaml
  only: [merge_requests]
  when: manual

Stop Review:
  stage: dismiss
  script:
    - werf dismiss --with-namespace
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]
```
{% endraw %}

#### №2 Автоматически по имени ветки

> Данный вариант реализует подход описанный в разделе [Выкат на review из ветки по шаблону автоматически]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#выкат-на-review-из-ветки-по-шаблону-автоматически" | true_relative_url }})

В предложенном ниже варианте автоматический релиз выполняется для каждого коммита в MR, в случае, если имя git-ветки имеет префикс `review-`.

{% raw %}
```yaml
Review:
  extends: .base_deploy
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    url: http://${CI_PROJECT_NAME}-${CI_MERGE_REQUEST_ID}.kube.DOMAIN
    on_stop: Stop Review
    auto_stop_in: 1 day
  artifacts:
    paths:
      - werf.yaml
  rules:
    - if: $CI_MERGE_REQUEST_ID && $CI_COMMIT_REF_NAME =~ /^review-/

Stop Review:
  stage: dismiss
  script:
    - werf dismiss --with-namespace
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]
```
{% endraw %}

#### №3 Полуавтоматический режим с лейблом (рекомендованный)

> Данный вариант реализует подход описанный в разделе [Выкат на review из pull request автоматически после ручной активации]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#выкат-на-review-из-pull-request-автоматически-после-ручной-активации" | true_relative_url }})

Полуавтоматический режим с лейблом — это комплексное решение, объединяющие первые два варианта.

При проставлении специального лейбла, в примере ниже `review`, пользователь активирует автоматический выкат в review окружения для каждого коммита.
При снятии лейбла происходит остановка окружения GitLab, удаление review-релиза.

{% raw %}
```yaml
Review:
  stage: deploy
  script:

    - >
      # do optional deploy/dismiss

      if echo $CI_MERGE_REQUEST_LABELS | tr ',' '\n' | grep -q -P '^review$'; then
        werf converge --skip-build --set "env_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
      else
        if werf helm get $(werf helm get-release) 2>/dev/null; then
          werf dismiss --with-namespace
        fi
      fi
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    url: http://${CI_PROJECT_NAME}-${CI_MERGE_REQUEST_ID}.kube.DOMAIN
    on_stop: Stop Review
    auto_stop_in: 1 day
  artifacts:
    paths:
      - werf.yaml
  dependencies:
    - Build and Publish
  only: [merge_requests]
  tags: [werf]

Stop Review:
  stage: dismiss
  script:
    - werf dismiss --with-namespace
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]
```
{% endraw %}

Для проверки наличия лейбла у MR используется переменная среды `CI_MERGE_REQUEST_LABELS`.

### Варианты организации staging и production окружений

Предложенные далее варианты являются наиболее эффективными комбинациями правил выката **staging** и **production** окружений.

В нашем случае, данные окружения являются определяющими, поэтому названия вариантов соответствуют названиям окончательных готовых workflow, предложенных в [конце статьи](#полный-gitlab-ciyml-для-готовых-workflow).

#### №1 Fast and Furious (рекомендованный)

> Данный вариант реализует подходы описанные в разделах [Выкат на production из master автоматически]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#выкат-на-production-из-master-автоматически" | true_relative_url }}) и [Выкат на production-like из pull request по кнопке]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#выкат-на-production-like-из-pull-request-по-кнопке" | true_relative_url }})

Выкат в **production** происходит автоматически при любых изменениях в master. Выполнить выкат в **staging** можно по кнопке в MR.

{% raw %}
```yaml
Deploy to Staging:
  extends: .base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}.kube.DOMAIN
  only: [merge_requests]
  when: manual

Deploy to Production:
  extends: .base_deploy
  environment:
    name: production
    url: https://www.company.org
  only: [master]
```
{% endraw %}

Варианты отката изменений в production:
- [revert изменений](https://git-scm.com/docs/git-revert) в master (**рекомендованный**);
- выкат стабильного MR или воспользовавшись кнопкой [Rollback environment](https://docs.gitlab.com/ee/ci/environments/#environment-rollback).

#### №2 Push the Button

> Данный вариант реализует подходы описанные в разделах [Выкат на production из master по кнопке]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#выкат-на-production-из-master-по-кнопке" | true_relative_url }}) и [Выкат на staging из master автоматически]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#выкат-на-staging-из-master-автоматически" | true_relative_url }})

Выкат **production** осуществляется по кнопке у коммита в master, а выкат в **staging** происходит автоматически при любых изменениях в master.

{% raw %}
```yaml
Deploy to Staging:
  extends: .base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}.kube.DOMAIN
  only: [master]

Deploy to Production:
  extends: .base_deploy
  environment:
    name: production
    url: https://www.company.org
  only: [master]
  when: manual
```
{% endraw %}

Варианты отката изменений в production:
- по кнопке у стабильного коммита или воспользовавшись кнопкой [Rollback environment](https://docs.gitlab.com/ee/ci/environments/#environment-rollback) (**рекомендованный**);
- выкат стабильного MR и нажатии кнопки.

#### №3 Tag everything (рекомендованный)

> Данный вариант реализует подходы описанные в разделах [Выкат на production из тега автоматически]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#выкат-на-production-из-тега-автоматически" | true_relative_url }}) и [Выкат на staging из master по кнопке]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#выкат-на-staging-из-master-по-кнопке" | true_relative_url }})

Выкат в **production** выполняется при проставлении тега, а в **staging** по кнопке у коммита в master.

{% raw %}
```yaml
Deploy to Staging:
  extends: .base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}.kube.DOMAIN
  only: [master]
  when: manual

Deploy to Production:
  extends: .base_deploy
  environment:
    name: production
    url: https://www.company.org
  only:
    - tags
```
{% endraw %}

Варианты отката изменений в production:
- нажатие кнопки на другом теге (**рекомендованный**);
- создание нового тега на старый коммит (так делать не надо).

#### №4 Branch, branch, branch!

> Данный вариант реализует подходы описанные в разделах [Выкат на production из ветки автоматически]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#выкат-на-production-из-ветки-автоматически" | true_relative_url }}) и [Выкат на production-like из ветки автоматически]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#выкат-на-production-like-из-ветки-автоматически" | true_relative_url }})

Выкат в **production** происходит автоматически при любых изменениях в ветке production, а в **staging** при любых изменениях в ветке master.

{% raw %}
```yaml
Deploy to Staging:
  extends: .base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}.kube.DOMAIN
  only: [master]

Deploy to Production:
  extends: .base_deploy
  environment:
    name: production
    url: https://www.company.org
  only: [production]
```
{% endraw %}

Варианты отката изменений в production:
- воспользовавшись кнопкой [Rollback environment](https://docs.gitlab.com/ee/ci/environments/#environment-rollback);
- [revert изменений](https://git-scm.com/docs/git-revert) в ветке production;
- [revert изменений](https://git-scm.com/docs/git-revert) в master и fast-forward merge в ветку production;
- удаление коммита из ветки production и push-force.

## Очистка образов

{% raw %}
```yaml
Cleanup:
  stage: cleanup
  script:
    - werf cr login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_REPO}
    - werf cleanup
  only: [schedules]
  tags: [werf]
```
{% endraw %}

В werf встроен эффективный механизм очистки, который позволяет избежать переполнения container registry и диска сборочного узла от устаревших и неиспользуемых образов.
Более подробно ознакомиться с возможностями очистки, встроенными в werf, можно [здесь]({{ "advanced/cleanup.html" | true_relative_url }}).

Чтобы использовать очистку, необходимо создать `Personal Access Token` в GitLab с необходимыми правами. С помощью данного токена будет осуществляться авторизация в container registry перед очисткой.

Для вашего тестового проекта вы можете просто создать `Personal Access Token` а вашей учетной записи GitLab. Для этого откройте страницу `Settings` в GitLab (настройки вашего профиля), затем откройте раздел `Access Token`. Укажите имя токена, в разделе Scope отметьте `api` и нажмите `Create personal access token` — вы получите `Personal Access Token`.

Чтобы передать `Personal Access Token` в переменную окружения GitLab откройте ваш проект, затем откройте `Settings` —> `CI/CD` и разверните `Variables`. Создайте новую переменную окружения `WERF_IMAGES_CLEANUP_PASSWORD` и в качестве ее значения укажите содержимое `Personal Access Token`. Для безопасности отметьте созданную переменную как `protected`.

Стадия очистки запускается только по расписанию, которое вы можете определить открыв раздел `CI/CD` —> `Schedules` настроек проекта в GitLab. Нажмите кнопку `New schedule`, заполните описание задания и определите шаблон запуска в обычном cron-формате. В качестве ветки оставьте master (название ветки не влияет на процесс очистки), отметьте `Active` и сохраните pipeline.

## Полный .gitlab-ci.yml для готовых workflow

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_gitlab_ci_1')">№1 Fast and Furious (рекомендованный)</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_gitlab_ci_2')">№2 Push the Button</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_gitlab_ci_3')">№3 Tag everything (рекомендованный)</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_gitlab_ci_4')">№4 Branch, branch, branch!</a>
</div>

<div id="complete_gitlab_ci_1" class="tabs__content no_toc_section active" markdown="1">

### Детали workflow
{:.no_toc}

> Подробнее про workflow можно почитать в отдельной [статье]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#1-fast-and-furious" | true_relative_url }})

* [Сборка и публикация](#сборка-и-публикация-образов-приложения).
* Выкат на review контур по стратегии [№3 Полуавтоматический режим с лейблом (рекомендованный)](#3-полуавтоматический-режим-с-лейблом-рекомендованный).
* Выкат на staging и production контуры осуществляется по стратегии [№1 Fast and Furious (рекомендованный)](#1-fast-and-furious-рекомендованный).
* [Очистка стадий](#очистка-образов) выполняется по расписанию раз в сутки.

### .gitlab-ci.yml
{:.no_toc}

{% raw %}
```yaml
stages:
  - build
  - deploy
  - dismiss
  - cleanup

before_script:
  - type trdl && . $(trdl use werf 1.2 stable)
  - type werf && source $(werf ci-env gitlab --as-file)

Build and Publish:
  stage: build
  script:
    - werf build
  except: [schedules]
  tags: [werf]

.base_deploy:
  stage: deploy
  script:
    - werf converge --skip-build --set "env_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  dependencies:
    - Build and Publish
  tags: [werf]

Review:
  stage: deploy
  script:
    - >
      # do optional deploy/dismiss

      if echo $CI_MERGE_REQUEST_LABELS | tr ',' '\n' | grep -q -P '^review$'; then
        werf converge --skip-build --set "env_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
      else
        if werf helm get $(werf helm get-release) 2>/dev/null; then
          werf dismiss --with-namespace
        fi
      fi
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    url: http://${CI_PROJECT_NAME}-${CI_MERGE_REQUEST_ID}.kube.DOMAIN
    on_stop: Stop Review
    auto_stop_in: 1 day
  artifacts:
    paths:
      - werf.yaml
  dependencies:
    - Build and Publish
  only: [merge_requests]
  tags: [werf]

Stop Review:
  stage: dismiss
  script:
    - werf dismiss --with-namespace
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]

Deploy to Staging:
  extends: .base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}.kube.DOMAIN
  only: [merge_requests]
  when: manual

Deploy to Production:
  extends: .base_deploy
  environment:
    name: production
    url: https://www.company.org
  only: [master]

Cleanup:
  stage: cleanup
  script:
    - werf cr login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_REPO}
    - werf cleanup
  only: [schedules]
  tags: [werf]
```
{% endraw %}

</div>
<div id="complete_gitlab_ci_2" class="tabs__content no_toc_section" markdown="1">

### Детали workflow
{:.no_toc}

> Подробнее про workflow можно почитать в отдельной [статье]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#2-push-the-button" | true_relative_url }})

* [Сборка и публикация](#сборка-и-публикация-образов-приложения).
* Выкат на review контур по стратегии [№1 Вручную](#1-вручную).
* Выкат на staging и production контуры осуществляется по стратегии [№2 Push the Button](#2-push-the-button).
* [Очистка стадий](#очистка-образов) выполняется по расписанию раз в сутки.

### .gitlab-ci.yml

{:.no_toc}

{% raw %}
```yaml
stages:
  - build
  - deploy
  - dismiss
  - cleanup

before_script:
  - type trdl && . $(trdl use werf 1.2 stable)
  - type werf && source $(werf ci-env gitlab --as-file)

Build and Publish:
  stage: build
  script:
    - werf build
  except: [schedules]
  tags: [werf]

.base_deploy:
  stage: deploy
  script:
    - werf converge --skip-build --set "env_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  dependencies:
    - Build and Publish
  except: [schedules]
  tags: [werf]

Review:
  extends: .base_deploy
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    url: http://${CI_PROJECT_NAME}-${CI_MERGE_REQUEST_ID}.kube.DOMAIN
    on_stop: Stop Review
    auto_stop_in: 1 day
  artifacts:
    paths:
      - werf.yaml
  only: [merge_requests]
  when: manual

Stop Review:
  stage: dismiss
  script:
    - werf dismiss --with-namespace
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]

Deploy to Staging:
  extends: .base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}.kube.DOMAIN
  only: [master]

Deploy to Production:
  extends: .base_deploy
  environment:
    name: production
    url: https://www.company.org
  only: [master]
  when: manual

Cleanup:
  stage: cleanup
  script:
    - werf cr login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_REPO}
    - werf cleanup
  only: [schedules]
  tags: [werf]
```
{% endraw %}

</div>

<div id="complete_gitlab_ci_3" class="tabs__content no_toc_section" markdown="1">

### Детали workflow
{:.no_toc}

> Подробнее про workflow можно почитать в отдельной [статье]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#3-tag-everything" | true_relative_url }})

* [Сборка и публикация](#сборка-и-публикация-образов-приложения).
* Выкат на review контур по стратегии [№1 Вручную](#1-вручную).
* Выкат на staging и production контуры осуществляется по стратегии [№3 Tag everything](#3-tag-everything-рекомендованный).
* [Очистка стадий](#очистка-образов) выполняется по расписанию раз в сутки.

### .gitlab-ci.yml
{:.no_toc}

{% raw %}
```yaml
stages:
  - build
  - deploy
  - dismiss
  - cleanup

before_script:
  - type trdl && . $(trdl use werf 1.2 stable)
  - type werf && source $(werf ci-env gitlab --as-file)

Build and Publish:
  stage: build
  script:
    - werf build
  except: [schedules]
  tags: [werf]

.base_deploy:
  stage: deploy
  script:
    - werf converge --skip-build --set "env_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  dependencies:
    - Build and Publish
  except: [schedules]
  tags: [werf]

Review:
  extends: .base_deploy
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    url: http://${CI_PROJECT_NAME}.kube.DOMAIN
    on_stop: Stop Review
    auto_stop_in: 1 day
  artifacts:
    paths:
      - werf.yaml
  only: [merge_requests]
  when: manual

Stop Review:
  stage: dismiss
  script:
    - werf dismiss --with-namespace
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]

Deploy to Staging:
  extends: .base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}.kube.DOMAIN
  only: [master]
  when: manual

Deploy to Production:
  extends: .base_deploy
  environment:
    name: production
    url: https://www.company.org
  only: [tags]

Cleanup:
  stage: cleanup
  script:
    - werf cr login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_REPO}
    - werf cleanup
  only: [schedules]
  tags: [werf]
```
{% endraw %}
</div>

<div id="complete_gitlab_ci_4" class="tabs__content no_toc_section" markdown="1">

### Детали workflow
{:.no_toc}

> Подробнее про workflow можно почитать в отдельной [статье]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#4-branch-branch-branch" | true_relative_url }})

* [Сборка и публикация](#сборка-и-публикация-образов-приложения).
* Выкат на review контур по стратегии [№2 Автоматически по имени ветки](#2-автоматически-по-имени-ветки).
* Выкат на staging и production контуры осуществляется по стратегии [№4 Branch, branch, branch!](#4-branch-branch-branch).
* [Очистка стадий](#очистка-образов) выполняется по расписанию раз в сутки.

### .gitlab-ci.yml
{:.no_toc}

{% raw %}
```yaml
stages:
  - build
  - deploy
  - dismiss
  - cleanup

before_script:
  - type trdl && . $(trdl use werf 1.2 stable)
  - type werf && source $(werf ci-env gitlab --as-file)

Build and Publish:
  stage: build
  script:
    - werf build
  except: [schedules]
  tags: [werf]

.base_deploy:
  stage: deploy
  script:
    - werf converge --skip-build --set "env_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  dependencies:
    - Build and Publish
  except: [schedules]
  tags: [werf]

Review:
  extends: .base_deploy
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    url: http://${CI_PROJECT_NAME}.kube.DOMAIN
    on_stop: Stop Review
    auto_stop_in: 1 day
  artifacts:
    paths:
      - werf.yaml
  rules:
    - if: $CI_MERGE_REQUEST_ID && $CI_COMMIT_REF_NAME =~ /^review-/

Stop Review:
  stage: dismiss
  script:
    - werf dismiss --with-namespace
  environment:
    name: review-${CI_MERGE_REQUEST_ID}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]

Deploy to Staging:
  extends: .base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}.kube.DOMAIN
  only: [master]

Deploy to Production:
  extends: .base_deploy
  environment:
    name: production
    url: https://www.company.org
  only: [production]

Cleanup:
  stage: cleanup
  script:
    - werf cr login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_REPO}
    - werf cleanup
  only: [schedules]
  tags: [werf]
```
{% endraw %}
</div>
