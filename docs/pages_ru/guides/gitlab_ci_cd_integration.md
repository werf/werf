---
title: Интеграция с GitLab CI/CD
sidebar: documentation
permalink: documentation/guides/gitlab_ci_cd_integration.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Обзор задачи

В статье рассматриваются различные варианты настройки CI/CD с использованием GitLab CI и werf.

Окончательные версии конфигураций .gitlab-ci.yml для различных workflow доступны [в конце статьи](#полный-gitlab-ciyml-для-различных-стратегий). 

## Требования

* Кластер Kubernetes и настроенный для работы с ним `kubectl`.
* GitLab-сервер версии выше 10.x либо учетная запись на [gitlab.com](https://gitlab.com/).
* Docker registry, встроенный в GitLab или выделенный.
* Приложение, которое успешно собирается и деплоится с werf.

## Инфраструктура

![scheme]({% asset howto_gitlabci_scheme.png @path %})

* Кластер Kubernetes.
* GitLab со встроенным Docker registry.
* Узел или группа узлов, с предустановленным werf и зависимостями.

Организовать работу werf внутри Docker-контейнера можно, но мы не поддерживаем данный способ.
Найти информацию по этому вопросу и обсудить можно в [issue](https://github.com/flant/werf/issues/1926).
В данном примере и в целом мы рекомендуем использовать _shell executor_.

Процесс деплоя требует наличия доступа к кластеру через `kubectl`, поэтому необходимо установить и настроить `kubectl` на узле, с которого будет запускаться werf.
Если не указывать конкретный контекст опцией `--kube-context` или переменной окружения `WERF_KUBE_CONTEXT`, то werf будет использовать контекст `kubectl` по умолчанию.  

В конечном счете werf требует наличия доступа на используемых узлах:
- к Git-репозиторию кода приложения;
- к Docker registry;
- к кластеру Kubernetes.

### Настройка runner

На узле, где предполагается запуск werf, установим и настроим GitLab-runner:
1. Создадим проект в GitLab и добавим push кода приложения.
1. Получим токен регистрации GitLab-runner'а:
   * заходим в проекте в GitLab `Settings` —> `CI/CD`;
   * во вкладке `Runners` необходимый токен находится в секции `Setup a specific Runner manually`.
1. Установим GitLab-runner [по инструкции](https://docs.gitlab.com/runner/install/linux-manually.html).
1. Зарегистрируем `gitlab-runner`, выполнив [шаги](https://docs.gitlab.com/runner/register/index.html) за исключением следующих моментов:
   * используем `werf` в качестве тега runner'а;
   * используем `shell` в качестве executor для runner'а.
1. Добавим пользователя `gitlab-runner` в группу `docker`.

   ```shell
   sudo usermod -aG docker gitlab-runner
   ```

1. Установим [Docker](https://kubernetes.io/docs/setup/independent/install-kubeadm/#installing-docker) и настроим `kubectl`, если они не были установлены ранее.
1. Установим [зависимости werf]({{ site.baseurl }}/documentation/guides/getting_started.html#требования).
1. Установим [multiwerf](https://github.com/flant/multiwerf) пользователем `gitlab-runner`:

   ```shell
   sudo su gitlab-runner
   mkdir -p ~/bin
   cd ~/bin
   curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash
   ```

1. Скопируем файл конфигурации `kubectl` в домашнюю папку пользователя `gitlab-runner`.
   ```shell
   mkdir -p /home/gitlab-runner/.kube &&
   sudo cp -i /etc/kubernetes/admin.conf /home/gitlab-runner/.kube/config &&
   sudo chown -R gitlab-runner:gitlab-runner /home/gitlab-runner/.kube
   ```

После того, как GitLab-runner настроен, можно переходить к настройке pipeline.

## Pipeline

В статье будет рассмотрен pipeline состоящий из следующего набора стадий:
* `build-and-publish` — стадия сборки и публикации образов приложения;
* `deploy` — стадия деплоя приложения для одного из контуров кластера;
* `dismiss` — стадия удаления приложения для review окружения;
* `cleanup` — стадия очистки хранилища стадий и Docker registry.

### Сборка и публикация образов приложения

```yaml
Build and Publish:
  stage: build-and-publish
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf build-and-publish
  except: [schedules]
  tags: [werf]
```

> Забегая вперед, очистка хранилища стадий и Docker registry предполагает запуск соответствующего задания по расписанию.
> Так как при очистке не требуется выполнять сборку образов, то указываем `except: [schedules]`, чтобы стадия сборки не запускалась в случае работы pipeline по расписанию.

Конфигурация задания достаточно проста, поэтому хочется сделать акцент на том, чего в ней нет — явной авторизации в Docker registry, вызова `docker login`. 

В простейшем случае, при использовании встроенного Docker registry, авторизация выполняется автоматически при вызове команды `werf ci-env`. В качестве необходимых аргументов используются переменные окружения GitLab `CI_JOB_TOKEN` (подробнее про модель разграничения доступа при выполнении заданий в GitLab можно прочитать [здесь](https://docs.gitlab.com/ee/user/project/new_ci_build_permissions_model.html)) и `CI_REGISTRY_IMAGE`.

При вызове `werf ci-env` создаётся временный docker config, который используется всеми командами в shell-сессии (в том числе docker). Таким образов, параллельные задания никак не пересекаются при использовании docker и временный токен в конфигурации не перетирается. 

Если необходимо выполнить авторизацию с произвольными учётными данными, `docker login` должен выполняться после вызова `werf ci-env` (подробнее про авторизацию [в отдельной статье]({{ site.baseurl }}/documentation/reference/working_with_docker_registries.html#авторизация-docker)).   

> Для удаления образов из встроенного в GitLab Docker registry требуется `Personal Access Token`. Подробнее далее в разделе посвящённом [очистке](#очистка-образов)

### Выкат приложения

Набор контуров (а равно — окружений GitLab) в кластере Kubernetes для деплоя приложения зависит от ваших потребностей, но наиболее используемые контуры следующие:
* Контур production. Финальный контур в pipeline, предназначенный для эксплуатации версии приложения, доставки конечному пользователю.
* Контур staging. Контур, который может использоваться для проведения финального тестирования приложения в приближенной к production среде. 
* Контур review. Динамический (временный) контур, используемый разработчиками при разработке для оценки работоспособности написанного кода, первичной оценки работоспособности приложения и т.п.

> Описанный набор и их функции — это не истина в последней инстанции и вы можете описывать CI/CD процессы под ваши нужны с произвольным количеством контуров со своей логикой. 
> Далее будут представлены популярные стратегии и практики, на базе которых мы предлагаем выстраивать ваши процессы в GitLab CI 

Прежде всего необходимо описать шаблон, общую часть для деплоя в любой контур, что позволит уменьшить размер файла `.gitlab-ci.yml`, улучшит его читаемость, а также позволит далее сосредоточиться на workflow.

```yaml
.base_deploy: &base_deploy
  stage: deploy
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf deploy --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  dependencies:
    - Build and Publish
  except: [schedules]
  tags: [werf]
```

При использовании шаблона `base_deploy` для каждого контура будет определяться своё окружение GitLab:                

```yaml
environment:
  name: <environment name>
  url: <url>
  ...
```

При выполнении задания, `werf ci-env` устанавливает переменную `global.env` в соответствии с именем окружения GitLab (`CI_ENVIRONMENT_SLUG`).

Для того, чтобы по-разному конфигурировать приложение для используемых контуров кластера в helm-шаблонах можно использовать Go-шаблоны и переменную `.Values.global.env`.

Также в шаблоне используется адрес окружения, URL для доступа к разворачиваемому в контуре приложению, который передаётся параметром `global.ci_url`. 
Это значение может использоваться в helm-шаблонах, например, для конфигурации Ingress-ресурсов.

#### Варианты организации review окружения

Как уже было сказано ранее, review окружение это временный контур, поэтому помимо выката у этого окружения также должна быть и очистка.

Рассмотрим базовые конфигурации `Review` и `Stop Review` заданий, которые лягут в основу предложенных стратегий.

```yaml
Review:
  <<: *base_deploy
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
    on_stop: Stop Review
    auto_stop_in: 1 day
  artifacts:
    paths:
      - werf.yaml
  only: [merge_requests]

Stop Review: 
  stage: dismiss
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf dismiss --with-namespace
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]
```

В задание `Review` описывается выкат review-релиза в динамическое окружение, в основу имени которого закладывается имя используемой git-ветки. Параметр `auto_stop_in` позволяет указать период отсутствия активности в MR, после которого окружение GitLab будет автоматически остановлено. Остановка окружения GitLab сама по себе никак не влияет на ресурсы в кластере, review-релиз, поэтому в дополнении необходимо определить задание, которое вызывается при остановке (`on_stop`). В нашем случае, это задание `Stop Review`. 

Задание `Stop Review` выполняет удаление review-релиза, а также остановку окружения GitLab (`action: stop`): werf удаляет helm-релиз, и, соответственно, namespace в Kubernetes со всем его содержимым ([werf dismiss]({{ site.baseurl }}/documentation/cli/main/dismiss.html)). Задание `Stop Review` может быть запущено вручную после деплоя на review контур, а также автоматически GitLab-сервером, например, при удалении соответствующей ветки в результате слияния ветки с master и указания соответствующей опции в интерфейсе GitLab.

Для выполнения `werf dismiss` требуется werf.yaml, так как в нём содержаться [шаблоны имени релиза и namespace](https://werf.io/documentation/configuration/deploy_into_kubernetes.html). Так как при удалении ветки нет возможности использовать исходники из git, в задании `Stop Review` используется werf.yaml, сохранённый при выполнении задания `Review`, и отключено подтягивание изменений из git (`GIT_STRATEGY: none`). 

Таким образом, по умолчанию закладываем следующие варианты удаления review окружения:
* вручную;
* автоматически при отсутствии активности MR в течение суток и при удалении ветки.

Далее разберём основные стратегии при организации выката review окружения. 

> Мы не ограничиваем вас предложенными вариантами, даже напротив — рекомендуем комбинировать предложенные варианты, создавать стратегию под нужды вашей команды

##### №1 Вручную 

При таком подходе пользователь выкатывает и удаляет окружение по кнопке в pipeline.

Он самый простой и может быть удобен в случае, когда review окружение не используется в разработке, выкаты происходят редко.
По сути для проверки перед принятием MR. 

```yaml
Review:
  <<: *base_deploy
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
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
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf dismiss --with-namespace
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]
```

##### №2 Автоматически по имени ветки

В предложенном ниже варианте автоматический релиз выполняется для каждого комита в MR, в случае, если имя git-ветки имеет префикс `review-`. 

```yaml
Review:
  <<: *base_deploy
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
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
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf dismiss --with-namespace
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]
```

##### №3 Полуавтоматический режим с лейблом (рекомендованный)

Полуавтоматический режим с лейблом — это комплексное решение, объединяющие первые два варианта. 

При проставлении специального лейбла, в примере ниже `review`, пользователь активирует автоматический выкат в review окружения для каждого комита. 
При снятии лейбла происходит остановка окружения GitLab, удаление review-релиза.    

```yaml
Review:
  stage: deploy
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - >
      if [ -z "$PRIVATE_TOKEN" ]; then
        echo "\$PRIVATE_TOKEN is not defined" >&2
        exit 1
      fi

      if curl -sS --header "PRIVATE-TOKEN: ${PRIVATE_TOKEN}" ${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/merge_requests/${CI_MERGE_REQUEST_IID} | jq .labels[] | grep -q '^"review"$'; then
        werf deploy --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
      else
        if werf helm get $(werf helm get-release) 2>/dev/null; then
          werf dismiss --with-namespace
        fi
      fi
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
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
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf dismiss --with-namespace
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]
```

Для проверки наличия лейбла у MR используется GitLab API. 
Так как токена `CI_JOB_TOKEN` не достаточно для работы с private репозиториями, необходимо сгенерировать специальный токен `PRIVATE_TOKEN`

#### Варианты организации staging и production окружений

##### №1 Fast and Furious или True CI/CD (рекомендованный)

Выкат в **production** происходит автоматически при любых изменениях в master. Выполнить выкат в **staging** можно по кнопке в MR.

```yaml
Deploy to Staging:
  <<: *base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
  only: [merge_requests]
  when: manual

Deploy to Production:
  <<: *base_deploy
  environment:
    name: production
    url: http://www.company.my
  only: [master]
```

Варианты отката изменений в production:
- [revert изменений](https://git-scm.com/docs/git-revert) в master (**рекомендованный**);
- выкат стабильного MR или воспользовавшись кнопкой [Rollback](https://docs.gitlab.com/ee/ci/environments.html#what-to-expect-with-a-rollback).

##### №2 Push the Button

Выкат **production** осуществляется по кнопке у комита в master, а выкат в **staging** происходит автоматически при любых изменениях в master.

```yaml
Deploy to Staging:
  <<: *base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
  only: [master]

Deploy to Production:
  <<: *base_deploy
  environment:
    name: production
    url: http://www.company.my
  only: [master]
  when: manual  
```

Варианты отката изменений в production:
- по кнопке у стабильного комита или воспользовавшись кнопкой [Rollback](https://docs.gitlab.com/ee/ci/environments.html#what-to-expect-with-a-rollback) (**рекомендованный**);
- выкат стабильного MR и нажатии кнопки.

##### №3 Tag everything (рекомендованный)

Выкат в **production** выполняется при проставлении тега, а в **staging** по кнопке у комита в master.

```yaml
Deploy to Staging:
  <<: *base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
  only: [master]
  when: manual

Deploy to Production:
  <<: *base_deploy
  environment:
    name: production
    url: http://www.company.my
  only:
    - tags
```

Варианты отката изменений в production:
- нажатие кнопки на другом теге (**рекомендованный**);
- создание нового тега на старый комит (так делать не надо).

##### №4 Branch, branch, branch!

Выкат в **production** происходит автоматически при любых изменениях в ветке production, а в **staging** при любых изменениях в ветке master.

```yaml
Deploy to Staging:
  <<: *base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
  only: [master]

Deploy to Production:
  <<: *base_deploy
  environment:
    name: production
    url: http://www.company.my
  only: [production]
```

Варианты отката изменений в production:
- воспользовавшись кнопкой [Rollback](https://docs.gitlab.com/ee/ci/environments.html#what-to-expect-with-a-rollback);
- [revert изменений](https://git-scm.com/docs/git-revert) в ветке production;
- [revert изменений](https://git-scm.com/docs/git-revert) в master и fast-forward merge в ветку production;
- удаление коммита из ветки production и push-force.

### Очистка образов

```yaml
Cleanup:
  stage: cleanup
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - docker login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_IMAGES_REPO}
    - werf cleanup
  only: [schedules]
  tags: [werf]
```

В werf встроен эффективный механизм очистки, который позволяет избежать переполнения Docker registry и диска сборочного узла от устаревших и неиспользуемых образов.
Более подробно ознакомиться с функционалом очистки, встроенным в werf, можно [здесь]({{ site.baseurl }}/documentation/reference/cleaning_process.html).

Чтобы использовать очистку, необходимо создать `Personal Access Token` в GitLab с необходимыми правами. С помощью данного токена будет осуществляться авторизация в Docker registry перед очисткой.

Для вашего тестового проекта вы можете просто создать `Personal Access Token` а вашей учетной записи GitLab. Для этого откройте страницу `Settings` в GitLab (настройки вашего профиля), затем откройте раздел `Access Token`. Укажите имя токена, в разделе Scope отметьте `api` и нажмите `Create personal access token` — вы получите `Personal Access Token`.

Чтобы передать `Personal Access Token` в переменную окружения GitLab откройте ваш проект, затем откройте `Settings` —> `CI/CD` и разверните `Variables`. Создайте новую переменную окружения `WERF_IMAGES_CLEANUP_PASSWORD` и в качестве ее значения укажите содержимое `Personal Access Token`. Для безопасности отметьте созданную переменную как `protected`.

Стадия очистки запускается только по расписанию, которое вы можете определить открыв раздел `CI/CD` —> `Schedules` настроек проекта в GitLab. Нажмите кнопку `New schedule`, заполните описание задания и определите шаблон запуска в обычном cron-формате. В качестве ветки оставьте master (название ветки не влияет на процесс очистки), отметьте `Active` и сохраните pipeline.

## Полный .gitlab-ci.yml для различных стратегий

<div class="tabs" style="display: grid">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_gitlab_ci_1')">№1 Fast and Furious (рекомендованный)</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_gitlab_ci_2')">№2 Push the Button</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_gitlab_ci_3')">№3 Tag everything (рекомендованный)</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_gitlab_ci_4')">№4 Branch, branch, branch!</a>
</div>

<div id="complete_gitlab_ci_1" class="tabs__content no_toc_section active" markdown="1">

#### Детали конфигурации 
{:.no_toc}

* [Сборка и публикация](#сборка-и-публикация-образов-приложения).
* Выкат на review контур по стратегии [№3 Полуавтоматический режим с лейблом](#3-полуавтоматический-режим-с-лейблом-рекомендованный).
* Выкат на staging и production контуры осуществляется по стратегии [№1 Fast and Furious или True CI/CD](#1-fast-and-furious-или-true-cicd-рекомендованный).
* [Очистка стадий](#очистка-образов). 

#### .gitlab-ci.yml 
{:.no_toc}

{% raw %}
```yaml
stages:
  - build-and-publish
  - deploy
  - dismiss
  - cleanup

Build and Publish:
  stage: build-and-publish
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf build-and-publish
  except: [schedules]
  tags: [werf]

.base_deploy: &base_deploy
  stage: deploy
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf deploy --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  dependencies:
    - Build and Publish
  tags: [werf]

Review:
  stage: deploy
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - >
      if [ -z "$PRIVATE_TOKEN" ]; then
        echo "\$PRIVATE_TOKEN is not defined" >&2
        exit 1
      fi

      if curl -sS --header "PRIVATE-TOKEN: ${PRIVATE_TOKEN}" ${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/merge_requests/${CI_MERGE_REQUEST_IID} | jq .labels[] | grep -q '^"review"$'; then
        werf deploy --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
      else
        if werf helm get $(werf helm get-release) 2>/dev/null; then
          werf dismiss --with-namespace
        fi
      fi
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
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
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf dismiss --with-namespace
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]

Deploy to Staging:
  <<: *base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
  only: [merge_requests]
  when: manual

Deploy to Production:
  <<: *base_deploy
  environment:
    name: production
    url: http://www.company.my
  only: [master]

Cleanup:
  stage: cleanup
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - docker login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_IMAGES_REPO}
    - werf cleanup
  only: [schedules]
  tags: [werf]
```
{% endraw %}

</div>
<div id="complete_gitlab_ci_2" class="tabs__content no_toc_section" markdown="1">

#### Детали конфигурации
{:.no_toc}

* [Сборка и публикация](#сборка-и-публикация-образов-приложения).
* Выкат на review контур по стратегии [№1 Вручную](#1-вручную).
* Выкат на staging и production контуры осуществляется по стратегии [№2 Push the Button](#2-push-the-button).
* [Очистка стадий](#очистка-образов). 

#### .gitlab-ci.yml
{:.no_toc}

{% raw %}
```yaml
stages:
  - build-and-publish
  - deploy
  - dismiss
  - cleanup

Build and Publish:
  stage: build-and-publish
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf build-and-publish
  except: [schedules]
  tags: [werf]

.base_deploy: &base_deploy
  stage: deploy
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf deploy --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  dependencies:
    - Build and Publish
  except: [schedules]
  tags: [werf]

Review:
  <<: *base_deploy
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
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
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf dismiss --with-namespace
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]

Deploy to Staging:
  <<: *base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
  only: [master]

Deploy to Production:
  <<: *base_deploy
  environment:
    name: production
    url: http://www.company.my
  only: [master]
  when: manual  

Cleanup:
  stage: cleanup
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - docker login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_IMAGES_REPO}
    - werf cleanup
  only: [schedules]
  tags: [werf]
```
{% endraw %}
</div>

<div id="complete_gitlab_ci_3" class="tabs__content no_toc_section" markdown="1">

#### Детали конфигурации
{:.no_toc}

* [Сборка и публикация](#сборка-и-публикация-образов-приложения).
* Выкат на review контур по стратегии [№1 Вручную](#1-вручную).
* Выкат на staging и production контуры осуществляется по стратегии [№3 Tag everything](#3-tag-everything-рекомендованный).
* [Очистка стадий](#очистка-образов). 

#### .gitlab-ci.yml
{:.no_toc}

{% raw %}
```yaml
stages:
  - build-and-publish
  - deploy
  - dismiss
  - cleanup

Build and Publish:
  stage: build-and-publish
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf build-and-publish
  except: [schedules]
  tags: [werf]

.base_deploy: &base_deploy
  stage: deploy
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf deploy --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  dependencies:
    - Build and Publish
  except: [schedules]
  tags: [werf]

Review:
  <<: *base_deploy
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
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
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf dismiss --with-namespace
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]

Deploy to Staging:
  <<: *base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
  only: [master]
  when: manual

Deploy to Production:
  <<: *base_deploy
  environment:
    name: production
    url: http://www.company.my
  only: [tags]

Cleanup:
  stage: cleanup
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - docker login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_IMAGES_REPO}
    - werf cleanup
  only: [schedules]
  tags: [werf]
```
{% endraw %}
</div>

<div id="complete_gitlab_ci_4" class="tabs__content no_toc_section" markdown="1">

#### Детали конфигурации
{:.no_toc}

* [Сборка и публикация](#сборка-и-публикация-образов-приложения).
* Выкат на review контур по стратегии [№2 Автоматически по имени ветки](#2-автоматически-по-имени-ветки).
* Выкат на staging и production контуры осуществляется по стратегии [№4 Branch, branch, branch!](#4-branch-branch-branch).
* [Очистка стадий](#очистка-образов). 

#### .gitlab-ci.yml
{:.no_toc}

{% raw %}
```yaml
stages:
  - build-and-publish
  - deploy
  - dismiss
  - cleanup

Build and Publish:
  stage: build-and-publish
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf build-and-publish
  except: [schedules]
  tags: [werf]

.base_deploy: &base_deploy
  stage: deploy
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf deploy --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  dependencies:
    - Build and Publish
  except: [schedules]
  tags: [werf]

Review:
  <<: *base_deploy
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
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
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - werf dismiss --with-namespace
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  variables:
    GIT_STRATEGY: none
  dependencies:
    - Review
  only: [merge_requests]
  when: manual
  tags: [werf]

Deploy to Staging:
  <<: *base_deploy
  environment:
    name: staging
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
  only: [master]

Deploy to Production:
  <<: *base_deploy
  environment:
    name: production
    url: http://www.company.my
  only: [production]
    
Cleanup:
  stage: cleanup
  script:
    - type multiwerf && . $(multiwerf use 1.1 stable --as-file)
    - type werf && source $(werf ci-env gitlab --verbose --as-file)
    - docker login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_IMAGES_REPO}
    - werf cleanup
  only: [schedules]
  tags: [werf]
```
{% endraw %}
</div>
