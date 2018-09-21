---
title: Gitlab CI/CD integration
sidebar: how_to
permalink: how_to/gitlab_ci_cd_integration.html
---

На данный момент полноценно поддерживается GitLab CI. В Travis CI может не работать часть функционала, но вы всегда можете внести коррективы и прислать Pull Request.

## место dapp в CI-процессе

Dapp необходимо установить на билд-сервер, туда, где запускается gitlab-runner.

Мы рекомендуем запускать билды на отдельной виртуальной машине. Не рекомендуем запускать dapp внутри Docker-а, т.к. работа докера (порождаемого dapp-ом в процессе сборки) внутри докера плохо предсказуема.

Преимуществом нашей утилиты в разрезе сборки является то, что вы можете удобно управлять несколькими раннерами, нет необходимости в ручной разливке каких-либо конфигов и доустановке софта в процессе эксплуатации.

### Пример .gitlab-ci.yml

```yaml
stages:
  - build
  - deploy
  # Никогда не забывайте про автоматизированную чистку образов!
  - cleanup_registry
  - cleanup_builder

build:
  stage: build
  script:
    # Полезная информация, которая поможет для дебага билда, если будут проблемы (дабы найти, где был билд и узнать, что лежало в используемых переменных)
    - dapp --version; pwd; set -x
    # Собираем образ и пушим в регистри. Обратите внимание на ключ --tag-ci
    - dapp dimg bp ${CI_REGISTRY_IMAGE} --tag-ci
  tags:
    - build
  except:
    - schedules

deploy:
  stage: deploy
  script:
    # Создаём namespace под приложение, если такого ещё нет.
    - kubectl get ns "${CI_PROJECT_PATH_SLUG}-${CI_ENVIRONMENT_SLUG}" || kubectl create ns "${CI_PROJECT_PATH_SLUG}-${CI_ENVIRONMENT_SLUG}"
    # Достаём секретный ключ registrysecret для доступа куба в registry. Этот ключ нужно установить в Кубы, иначе деплой работать не будет.
    - kubectl get secret registrysecret -n kube-system -o json |
                      jq ".metadata.namespace = \"${CI_PROJECT_PATH_SLUG}-${CI_ENVIRONMENT_SLUG}\"|
                      del(.metadata.annotations,.metadata.creationTimestamp,.metadata.resourceVersion,.metadata.selfLink,.metadata.uid)" |
                      kubectl apply -f -
    # Деплой, обратите внимание на ключ --tag-ci
    - dapp kube deploy
      --tag-ci
      --namespace ${CI_PROJECT_PATH_SLUG}-${CI_ENVIRONMENT_SLUG}
      --set "global.env=${CI_ENVIRONMENT_SLUG}"
      --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
      $CI_REGISTRY_IMAGE
  environment:
    # Названия, исходя из которых будет формироваться namespace приложения в Кубах и домен.
    name: stage
    url: https://${CI_COMMIT_REF_NAME}.${CI_PROJECT_PATH_SLUG}.example.com
  dependencies:
    - build
  tags:
    - deploy
  except:
    - schedules

Cleanup registry:
  stage: cleanup_registry
  script:
    - dapp --version; pwd; set -x
    - dapp dimg cleanup repo ${CI_REGISTRY_IMAGE}
  only:
    - schedules
  tags:
    - deploy

Cleanup builder:
  stage: cleanup_builder
  script:
    - dapp --version; pwd; set -x
    - dapp dimg stages cleanup local
        --improper-cache-version
        --improper-git-commit
        --improper-repo-cache
        ${CI_REGISTRY_IMAGE}
  only:
    - schedules
  tags:
    - build
```

### параметры, получаемые от раннера

Кроме параметров, передаваемых в dapp в явном виде (как прописано выше, в `gitlab-ci.yaml`), есть переменные, которые берутся из окружения, но в явном виде не передаются.

для аутентификация docker registry могут использоваться переменные `CI_JOB_TOKEN`, `CI_REGISTRY`, если они не переопределяются пользователем вот так: `docker login --username gitlab-ci-token --password $CI_JOB_TOKEN $CI_REGISTRY`

вместо указания username/password можно указать путь до произвольного docker config-а через `DAPP_DOCKER_CONFIG`.

для определения CI проверяется наличие переменной окружения `GITLAB_CI` или `TRAVIS`

`DAPP_HELM_RELEASE_NAME` переопределяет release name при deploy

`DAPP_DIMG_CLEANUP_REGISTRY_PASSWORD` специальный token для очистки registry (`dapp dimg cleanup repo`)

`DAPP_SLUG_V2` активирует схему slug-а с лимитом в 53 символа (такой лимит у имени release helm-а)

все переменные окружение с префиксом `CI_` прокидываются в helm-шаблоны  (`.Values.global.ci.CI_...`)

`ANSIBLE_ARGS` — прокидывается во все стадии с ansible и добавляется в командную строку к ansible-playbook

#### Параметры при схеме тегирования `--tag-build-id`

Тег формируется из переменных среды:

* `CI_BUILD_ID`;
* `TRAVIS_BUILD_NUMBER`.

#### Параметры при схеме тегирования `--tag-ci`

Тег формируется из переменных среды:

* `CI_BUILD_REF_NAME`, `CI_BUILD_TAG`;
* `TRAVIS_BRANCH`, `TRAVIS_TAG`.

#### Прочие параметры

Если сомневаетесь, есть ли параметр или нет, можете решиться [посмотреть исходные коды](https://github.com/flant/dapp/search?l=Ruby&p=1&q=ENV+path%3Alib)
