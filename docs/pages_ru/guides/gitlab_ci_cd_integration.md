---
title: Интеграция с GitLab CI/CD
sidebar: documentation
permalink: documentation/guides/gitlab_ci_cd_integration.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Обзор задачи

В статье рассматривается пример настройки CI/CD с использованием GitLab CI и werf.

## Требования

* Кластер Kubernetes и настроенный для работы с ним `kubectl`.
* GitLab-сервер версии выше 10.x либо учетная запись на [gitlab.com](https://gitlab.com/).
* Docker registry, встроенный в GitLab или выделенный.
* Приложение, которое успешно собирается и деплоится с werf.

## Инфраструктура

![scheme]({% asset howto_gitlabci_scheme.png @path %})

* Кластер Kubernetes.
* GitLab со встроенным Docker registry.
* Узел, на котором установлен werf (узел сборки и деплоя).

Обратите внимание, что процесс сборки и процесс деплоя выполняются на одном и том же узле — пока это единственно верный путь, т.к. распределённая сборка еще не поддерживается ([issue](https://github.com/flant/werf/issues/1614)).
Таким образом, в настоящий момент поддерживается только локальное хранилище стадий — `--stages-storage :local`.

Организовать работу werf внутри Docker-контейнера можно, но мы не поддерживаем данный способ.
Найти информацию по этому вопросу и обсудить можно в [issue](https://github.com/flant/werf/issues/1926).
В данном примере и в целом мы рекомендуем использовать _shell executor_.

Для хранения кэша сборки и служебных файлов werf использует папку `~/.werf`. Папка должна сохраняться и быть доступной на всех этапах pipeline. Это ещё одна из причин по которой мы рекумендуем отдавать предпочтение _shell executor_ вместо эфемерных окружений.

Процесс деплоя требует наличия доступа к кластеру через `kubectl`, поэтому необходимо установить и настроить `kubectl` на узле, с которого вы будет запускаться werf.
Если не указывать конкретный контекст опцией `--kube-context` или переменной окружения `$WERF_KUBE_CONTEXT`, то werf будет использовать контекст `kubectl` по умолчанию.  

В конечном счете werf требует наличия доступа:
- к Git-репозиторию кода приложения;
- к Docker registry;
- к кластеру Kubernetes.

Также необходимо присвоить тег `werf` соответствующему GitLab-runner'у.

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

   ```bash
   sudo usermod -aG docker gitlab-runner
   ```

1. Установим [Docker](https://kubernetes.io/docs/setup/independent/install-kubeadm/#installing-docker) и настроим `kubectl`, если они не были установлены ранее.
1. Установим [зависимости werf]({{ site.baseurl }}/documentation/guides/getting_started.html#requirements).
1. Установим [multiwerf](https://github.com/flant/multiwerf) пользователем `gitlab-runner`:

   ```bash
   sudo su gitlab-runner
   mkdir -p ~/bin
   cd ~/bin
   curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash
   ```

1. Скопируем файл конфигурации `kubectl` в домашнюю папку пользователя `gitlab-runner`.
   ```bash
   mkdir -p /home/gitlab-runner/.kube &&
   sudo cp -i /etc/kubernetes/admin.conf /home/gitlab-runner/.kube/config &&
   sudo chown -R gitlab-runner:gitlab-runner /home/gitlab-runner/.kube
   ```

После того как GitLab-runner настроен можно переходить к настройке pipeline.

## Pipeline

Создадим файл `.gitlab-ci.yml` в корне проекта и добавим следующие строки:

```yaml
stages:
  - build
  - deploy
  - cleanup
```

Мы определили следующие стадии:
* `build` — стадия сборки образов приложения;
* `deploy` — стадия деплоя приложения для одного из контуров кластера (например, stage, test, review, production или любой другой);
* `cleanup` — стадия очистки хранилища стадий и Docker registry.

### Сборка приложения

Добавим следующие строки в файл `.gitlab-ci.yml`:

```yaml
Build:
  stage: build
  script:
    - type multiwerf && . $(multiwerf use 1.0 stable --as-file)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - werf build-and-publish --stages-storage :local
  tags:
    - werf
  except:
    - schedules
```

Забегая вперед, очистка хранилища стадий и Docker registry предполагает запуск соответствующего задания по расписанию.
Так как при очистке не требуется выполнять сборку образов, то указываем `except: schedules`, чтобы стадия сборки не запускалась в случае работы pipeline по расписанию.

Для авторизации в Docker registry (при выполнении push/pull образов) werf использует переменную окружения GitLab `CI_JOB_TOKEN` (подробнее про модель разграничения доступа при выполнении заданий в GitLab можно прочитать [здесь](https://docs.gitlab.com/ee/user/project/new_ci_build_permissions_model.html)).
Это не единственный, но самый рекомендуемый вариант в случае работы с GitLab (подробно про авторизацию werf в Docker registry можно прочесть [здесь]({{ site.baseurl }}/documentation/reference/registry_authorization.html)).
В простейшем случае, если вы используете встроенный в GitLab Docker registry, вам не нужно делать никаких дополнительных действий для авторизации.

Если вам нужно чтобы werf не использовал переменную `CI_JOB_TOKEN` либо вы используете невстроенный в GitLab Docker registry (например, `Google Container Registry`), то можно ознакомиться с вариантами авторизации [здесь]({{ site.baseurl }}/documentation/reference/registry_authorization.html).

### Выкат приложения

Набор контуров (а равно — окружений GitLab) в кластере Kubernetes для деплоя приложения зависит от ваших потребностей, но наиболее используемые контуры следующие:
* Контур review. Динамический (временный) контур, используемый разработчиками в процессе работы над приложением для оценки работоспособности написанного кода, первичной оценки работоспособности приложения и т.п. Данный контур удаляется (а соответствущее окружение GitLab останавливается) после удаления ветки в репозитории либо вручную.
* Контур test. Тестовый контур, используемый разработчиками после завершения работы над какой-либо задачей для демонстрации результата, более полного тестирования приложения с тестовым набором данных и т.п. В тестовый контур можно вручную деплоить приложения из любой ветки и любой тег.
* Контур stage. Данный контур может использоваться для проведения финального тестирования приложения. Деплой в контур stage производится автоматически после принятия merge-request в ветку master, но это необязательно и вы можете настроить собственную логику.
* Контур production. Финальный контур в pipeline, предназначенный для деплоя готовой к продуктивной эксплуатации версии приложения. Мы подразумеваем, что на данный контур деплоятся только приложения из тегов и только вручную.

Описанный набор контуров и их функционал — это не правила и вы можете описывать CI/CD процессы под свои нужны.

Прежде всего необходимо описать шаблон, который мы будем использовать во всех заданиях деплоя, что позволит уменьшить размер файла `.gitlab-ci.yml` и улучшит его читаемость.

Добавим следующие строки в файл `.gitlab-ci.yml`:

```yaml
.base_deploy: &base_deploy
  stage: deploy
  script:
    - type multiwerf && . $(multiwerf use 1.0 stable --as-file)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    ## Следующая команда непосредственно выполняет деплой
    - werf deploy --stages-storage :local
        --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  ## Обратите внимание, что стадия деплоя обязательно зависит от стадии сборки. В случае ошибки на стадии сборки деплой не будет выполняться.
  dependencies:
    - Build
  tags:
    - werf
```

Обратите внимание на команду `werf deploy`.
Эта команда — основной шаг в процессе деплоя. В параметре `global.ci_url` передается URL для доступа к разворачиваемому в контуре приложению.
Вы можете использовать эти данные в ваших helm-шаблонах, например, для конфигурации Ingress-ресурсов.

Для того чтобы деплоить приложение в разные контуры кластера в helm-шаблонах можно использовать переменную `.Values.global.env`, обращаясь к ней внутри Go-шаблона (Go template).
Во время деплоя werf устанавливает переменную `global.env` в соответствии с именем окружения GitLab.

#### Контур review

Как было сказано выше, контур review — динамический контур (временный контур, контур разработчика) для оценки работоспособности написанного кода, первичной оценки работоспособности приложения и т.п.

Добавим следующие строки в файл `.gitlab-ci.yml`:

```yaml
Review:
  <<: *base_deploy
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    ## Измените суффикс доменного имени (здесь — `kube.DOMAIN`) при необходимости
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
    on_stop: Stop review
  only:
    - branches
  except:
    - master
    - schedules

Stop review:
  stage: deploy
  script:
    - type multiwerf && . $(multiwerf use 1.0 stable --as-file)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - werf dismiss --with-namespace
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  tags:
    - werf
  only:
    - branches
  except:
    - master
    - schedules
  when: manual
```

В примере выше определены два задания:

1. Review.
    В данном задании устанавливается переменная `name` со значением, использующим переменную окружения GitLab `CI_COMMIT_REF_SLUG`.
    GitLab формирует уникальное значение в переменной окружения `CI_COMMIT_REF_SLUG` для каждой ветки.

    Переменная `url` может использоваться в helm-шаблонах, например, для конфигурации Ingress-ресурсов.
2. Stop review.
    В данном задании werf удаляет helm-релиз, и, соответственно, namespace в Kubernetes со всем его содержимым ([werf dismiss]({{ site.baseurl }}/documentation/cli/main/dismiss.html)). Это задание может быть запущено вручную после деплоя на review-контур, а также оно может быть запущено GitLab-сервером, например, при удалении соответствующей ветки в результате слияния ветки с master и указания соответствующей опции в интерфейсе GitLab.

Задание `Review` не должно запускаться при изменениях в ветке master, т.к. это контур review — контур только для разработчиков.

#### Контур test

Мы не приводим описание заданий деплоя на контур test в настоящем примере, т.к. задания деплоя на контур test очень похожи на задания деплоя на контур stage.
Попробуйте описать задания деплоя на контур test по аналогии самостоятельно и надеемся вы получите удовольствие.

#### Контур stage

Как описывалось выше на контур stage можно деплоить только приложение из ветки master и это допустимо делать автоматически.

Добавим следующие строки в файл `.gitlab-ci.yml`:

```yaml
Deploy to Stage:
  <<: *base_deploy
  environment:
    name: stage
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
  only:
    - master
  except:
    - schedules
```

Эта часть конфигурации использует общий шаблон и устанавливает переменные окружения, свойственные данному контуру — `name` и `url`.

#### Контур production

Контур production — последний среди рассматриваемых и самый важный, так как он предназначен для конечного пользователя. Обычно в это окружения выкатываются теги и только вручную.

Добавим следующие строки в файл `.gitlab-ci.yml`:

```yaml
Deploy to Production:
  <<: *base_deploy
  environment:
    name: production
    url: http://www.company.my
  only:
    - tags
  when: manual
  except:
    - schedules
```

Обратите внимание на переменную `url` — так как мы деплоим приложение на контур production мы знаем публичный URL, по которому доступно приложение, и явно указываем его здесь. Далее мы будем использовать переменную `.Values.global.ci_url` в helm-шаблонах (вспомните описание шаблона `base_deploy` в начале статьи).

### Очистка образов

В werf встроен эффективный механизм очистки, который позволяет избежать переполнения Docker registry и диска сборочного узла от устаревших и неиспользуемых образов.
Более подробно ознакомиться с функционалом очистки, встроенным в werf, можно [здесь]({{ site.baseurl }}/documentation/reference/cleaning_process.html).

В результате работы werf наполняет локальное хранилище стадий, а также Docker registry собранными образами.

Для работы очистки в файле `.gitlab-ci.yml` выделена отдельная стадия — `cleanup`.

Чтобы использовать очистку необходимо создать `Personal Access Token` в GitLab с необходимыми правами. С помощью данного токена будет осуществляться авторизация в Docker registry перед очисткой.

Для вашего тестового проекта вы можете просто создать `Personal Access Token` а вашей учетной записи GitLab. Для этого откройте страницу `Settings` в GitLab (настройки вашего профиля), затем откройте раздел `Access Token`. Укажите имя токена, в разделе Scope отметьте `api` и нажмите `Create personal access token` — вы получите `Personal Access Token`.

Чтобы передать `Personal Access Token` в переменную окружения GitLab откройте ваш проект, затем откройте `Settings` —> `CI/CD` и разверните `Variables`. Создайте новую переменную окружения `WERF_IMAGES_CLEANUP_PASSWORD` и в качестве ее значения укажите содержимое `Personal Access Token`. Для безопасности отметьте созданную переменную как `protected`.

Добавим следующие строки в файл `.gitlab-ci.yml`:

```yaml
Cleanup:
  stage: cleanup
  script:
    - type multiwerf && . $(multiwerf use 1.0 stable --as-file)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - docker login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_IMAGES_REPO}
    - werf cleanup --stages-storage :local
  only:
    - schedules
  tags:
    - werf
```

Стадия очистки запускается только по расписанию, которое вы можете определить открыв раздел `CI/CD` —> `Schedules` настроек проекта в GitLab. Нажмите кнопку `New schedule`, заполните описание задания и определите шаблон запуска в обычном cron-формате. В качестве ветки оставьте master (название ветки не влияет на процесс очистки), отметьте `Active` и сохраните pipeline.

> Как это работает:
   - `werf ci-env` создает временный конфигурационный файл для docker (отдельный при каждом запуске и соответственно для каждого задания GitLab), а также экспортирует переменную окружения DOCKER_CONFIG со значением пути к созданному конфигурационному файлу;
   - `docker login` использует временный конфигурационный файл для docker, путь к которому был указан ранее в переменной окружения DOCKER_CONFIG для авторизации в Docker registry;
   - `werf cleanup` также использует временный конфигурационный файл из DOCKER_CONFIG.

## .gitlab-ci.yml

```yaml
stages:
  - build
  - deploy
  - cleanup

Build:
  stage: build
  script:
    - type multiwerf && . $(multiwerf use 1.0 stable --as-file)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - werf build-and-publish --stages-storage :local
  tags:
    - werf
  except:
    - schedules

.base_deploy: &base_deploy
  stage: deploy
  script:
    - type multiwerf && . $(multiwerf use 1.0 stable --as-file)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - werf deploy --stages-storage :local
        --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  dependencies:
    - Build
  tags:
    - werf

Review:
  <<: *base_deploy
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
    on_stop: Stop review
  only:
    - branches
  except:
    - master
    - schedules

Stop review:
  stage: deploy
  script:
    - type multiwerf && . $(multiwerf use 1.0 stable --as-file)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - werf dismiss --with-namespace
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  tags:
    - werf
  only:
    - branches
  except:
    - master
    - schedules
  when: manual

Deploy to Stage:
  <<: *base_deploy
  environment:
    name: stage
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
  only:
    - master
  except:
    - schedules

Deploy to Production:
  <<: *base_deploy
  environment:
    name: production
    url: http://www.company.my
  only:
    - tags
  when: manual
  except:
    - schedules

Cleanup:
  stage: cleanup
  script:
    - type multiwerf && . $(multiwerf use 1.0 stable --as-file)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - docker login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_IMAGES_REPO}
    - werf cleanup --stages-storage :local
  only:
    - schedules
  tags:
    - werf
```
