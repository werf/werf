---
title: Интеграция с GitLab CI/CD
sidebar: documentation
permalink: documentation/guides/gitlab_ci_cd_integration.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Обзор задачи

В статье рассматривается пример настройки CI/CD процессы с использованием GitLab CI и Werf.

## Требования

* Кластер Kubernetes и настроенный для работы с ним `kubectl`;
* GitLab-сервер версии выше 10.x (либо учетная запись на [gitLab.com](https://gitlab.com/));
* Docker Registry (встроенный в GitLab или выделенный);
* Какое-либо приложение, Docker-образ которого вы умеете успешно собирать и деплоить с помощью Werf.

## Инфраструктура

![scheme]({{ site.baseurl }}/images/howto_gitlabci_scheme.png)

* Кластер Kubernetes
* GitLab со встроенным Docker Registry
* Узел, на котором установлен Werf (узел сборки и деплоя)

Обратите внимание, что и процесс сборки и процесс деплоя выполняется на одном и том же узле — пока, это единственно верный путь, т.к. использование распределенного кеша стадий еще не поддерживается (мы заканчиваем работу над ним). Таким образом, в настоящий момент вы должны использовать опцию указания локального хранилища кэша стадий — `--stages-storage :local`.

Мы не рекомендуем запускать Werf внутри Docker, т.к. это может иногда приводить к непрогнозируемым результатам.

Чтобы настроить процессы CI/CD, вам необходимо описать стадии сборки, деплоя и очистки. Чтобы Gitlab-runner смог запускать `werf`, необходимо чтобы он был настроен с executor `shell`.

Для хранения кэша сборки и служебных файлов, Werf использует папку `~/.werf/`, подразумевая, что эта папка сохраняется для всех pipeline'ов (не удаляется). В частности из-за этого, мы не рекомендуем для процесса сборки использовать окружения, которые не сохраняют свое состояние (файлы между запусками). Примерами таких окружений могут являться некоторые варианты запуска runner'а в облачной среде и т.п.

Процесс деплоя требует наличия доступа к кластеру через `kubectl`, по-этому вам необходимо установить и настроить `kubectl` на узле, с которого вы будете запускать Werf. Werf использует выбранный по умолчанию контекст kubectl, если не указать какой-либо конкретный контекст (например, это можно сделать с помощью опции `--kube-context` или переменной окружения `$WERF_KUBE_CONTEXT`).

В конечном счете, Werf требует наличия доступа:
- к git-репозиторию кода приложения
- к Docker Registry
- к кластеру kubernetes.

Для работы примера вам также необходимо присвоить тэг `werf` соответствующему Gitlab-runner'у.

### Настройка runner'а

На узле, где предполагается запуск Werf, установите и настройте Gitlab-runner:
1. Создайте проект в Gitlab и сделайте push кода приложения.
1. Получите токен регистрации Gitlab-runner'а: в проекте в GitLab откройте `Settings` —> `CI/CD`, раскройте `Runners` и найдите необходимый токен в секции `Setup a specific Runner manually`.
1. [Установите gitlab-runner](https://docs.gitlab.com/runner/install/linux-manually.html).
1. Зарегистрируйте `gitlab-runner`.

    Для регистрации runner'a используйте [следующие шаги](https://docs.gitlab.com/runner/register/index.html), за исключением следующего:
    * укажите `werf` в качестве тэга runner'а;
    * укажите `shell` в качестве executor'а runner'а;
1. Добавьте пользователя `gitlab-runner` в группу `docker`.

   ```bash
   sudo usermod -aG docker gitlab-runner
   ```

1. [Установите](https://kubernetes.io/docs/setup/independent/install-kubeadm/#installing-docker) Docker и настройте kubectl (конечно если они не были установлены ранее).
1. Установите [зависимости Werf]({{ site.baseurl }}/documentation/guides/getting_started.html#requirements).
1. Установите [Multiwerf](https://github.com/flant/multiwerf) от пользователя `gitlab-runner`:

   ```bash
   sudo su gitlab-runner
   mkdir -p ~/bin
   cd ~/bin
   curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash
   ```

1. Скопируйте файл конфигурации kubectl в домашнюю папку пользователя `gitlab-runner`.
   ```bash
   mkdir -p /home/gitlab-runner/.kube &&
   sudo cp -i /etc/kubernetes/admin.conf /home/gitlab-runner/.kube/config &&
   sudo chown -R gitlab-runner:gitlab-runner /home/gitlab-runner/.kube
   ```

## Pipeline

Когда у вас есть настроенный Gitlab-runner, можно переходить к настройке pipeline в GitLab.

Когда GitLab запускает задание, устанавливается ряд [переменных окружения](https://docs.gitlab.com/ee/ci/variables/README.html) runner'a, часть из которых мы будем использовать.

Создайте файл `.gitlab-ci.yml` в корне проекта, в который добавьте следующие строки:

```yaml
stages:
  - build
  - deploy
  - cleanup
```

Мы определили следующие стадии:
* `build` — стадия сборки образов компонентов приложения;
* `deploy` — стадия деплоя собранных образов приложения в какой-либо контур кластера (например — stage, test, review, production или любой другой);
* `cleanup` — стадия для очистки сборочного кэша и Docker Registry.

### Стадия сборки

Добавьте следующие строки в файл `.gitlab-ci.yml`:

```yaml
Build:
  stage: build
  script:
    - type multiwerf && source <(multiwerf use 1.0 beta)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - werf build-and-publish --stages-storage :local
  tags:
    - werf
  except:
    - schedules
```

Забегая вперед скажем, что очистка сборочного кэша и Docker Registry предполагает запуск соответствующей стадии по расписанию. И, так как на стадии очистки не нужно выполнять повторную сборку образов, то в стадии сборки нужно указать `except: schedules`, чтобы стадия сборки не запускалась в случае работы pipeline по расписанию.

Для авторизации в registy (при выполнении push/pull образов) werf использует переменную окружения GitLab `CI_JOB_TOKEN` (читай подробнее про модель разграничения доступа при выполнении заданий в GitLab [здесь](https://docs.gitlab.com/ee/user/project/new_ci_build_permissions_model.html)). Это не единственный, но самый рекомендуемый вариант в случае работы с GitLab (читай более подробно про авторизацию Werf в Docker Registry [здесь]({{ site.baseurl }}/documentation/reference/registry_authorization.html)). В простейшем случае, если вы используете встроенный в GitLab registry, вам не нужно делать никаких дополнительных действий для авторизации.

Если вам нужно чтобы Werf не использовал переменную `CI_JOB_TOKEN`, либо вы используете не встроенный в GitLab Docker Registry (например, `Google Container Registry`), читайте подробнее о вариантах авторизации в Docker Registry [здесь]({{ site.baseurl }}/documentation/reference/registry_authorization.html).

### Стадии деплоя

Набор контуров (а равно — окружений GitLab) в кластере Kubernetes для деплоя приложения зависит от ваших потребностей, но наиболее используемые контуры — следующие:
* Контур review. Динамический (временный) контур, используется разработчиками в процессе работы над приложением — для оценки работоспособности написанного кода, первичной оценки работоспособности приложения и т.п. Данный контур удаляется (а соответствущее окружение GitLab останавливается) после удаления ветки в репозитории, либо вручную.
* Контур test. Тестовый контур, используемый разработчиками после завершения работы над какой-либо задачей для демонстрации результата, более полного тестирования приложения с тестовым набором данных и т.п. В тестовый контур можно вручную деплоить приложения из любой ветки и любой тэг.
* Контур stage. Данный контур может использоваться для проведения финального тестирования приложения. Деплой в контур stage производится автоматически после принятия merge-request'a в ветку master (но это не обязательно, вы можете настроить собственную логику).
* Контур production. Финальный контур в pipeline, предназначенный для деплоя готовой к продуктивной эксплуатации версии приложения. Мы подразумеваем что на данный контур деплоятся только приложения из тэгов, и только вручную.

Конечно, описанный набор контуров и их функционал — это не правила, и вы можете описывать свои CI/CD процессы.

Прежде всего, необходимо описать шаблон, который мы будем использовать во всех заданиях деплоя. Это позволит уменьшить размер файла `.gitlab-ci.yml` и улучшит его читаемость.

Добавьте следующие строки в файл `.gitlab-ci.yml`:

```yaml
.base_deploy: &base_deploy
  stage: deploy
  script:
    - type multiwerf && source <(multiwerf use 1.0 beta)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    ## Следующая команда непосредственно выполняет деплой
    - werf deploy --stages-storage :local
        --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  ## Обратите внимание, что стадия деплоя обязательно зависит от стадии сборки. В случае ошибки на стадии сборки - стадия деплоя не будет выполняться.
  dependencies:
    - Build
  tags:
    - werf
```

Обратите внимание на команду `werf deploy`. Эта команда — основной шаг в процессе деплоя. Обратите внимание также на параметры команды, — в параметре `global.ci_url` передается URL для доступа к разворачиваемому в контуре приложению. Вы можете использовать эти данные в ваших helm-шаблонах, например, для конфигурации Ingress-ресурсов.

Чтобы по-разному деплоить приложение в разные контуры кластера, в helm-шаблонах вы можете использовать переменную `.Values.global.env`, обращаясь к ней внутри Go-шаблона (go-template). Во время деплоя, Werf устанавливает переменную `global.env` в соответствии с именем окружения GitLab.

#### Контур review

Как было сказано выше, контур review — динамический контур (временный контур, контур разработчика) для оценки работоспособности написанного кода, первичной оценки работоспособности приложения и т.п.

Добавьте следующие строки в файл `.gitlab-ci.yml`:

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
    - type multiwerf && source <(multiwerf use 1.0 beta)
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

В примере выше, мы определили два задания:
1. Review.
    В данном задании мы установили переменную `name` согласно содержимому переменной окружения GitLab `CI_COMMIT_REF_SLUG`. GitLab формирует уникальное значение в переменной окружения `CI_COMMIT_REF_SLUG` для каждой ветки.

    Переменная `url` устанавливается для того, чтобы вы могли использовать её в helm-шаблонах, например, для конфигурации Ingress-ресурсов.
2. Stop review.
    В данном задании Werf удалит helm-релиз, и, соответственно namespace в Kubernetes со всем его содержимым (читай подробнее про команду [werf dismiss]({{ site.baseurl }}/documentation/cli/main/dismiss.html)). Это задание может быть запущено вручную после деплоя на review-контур, а также оно может быть запущено GitLab-сервером, например, при удалении соответствующей ветки в результате слияния ветки с master и указания соответствующей опции в интерфейсе GitLab.

Задание `Review` не должно запускаться при изменениях в ветке мастер, т.к. это контур review — контур только для разработчиков.

#### Контур test

Мы не приводим описание заданий деплоя на контур test в настоящем примере, т.к. задания деплоя на контур test очень похожи на задания деплоя на контур stage (смотри ниже). Попробуйте описать задания деплоя на контур test по аналогии самостоятельно, и надеемся вы получите удовольствие.

#### Контур stage

Как описывалось выше, на контур stage можно деплоить только приложение из ветки master и допустимо делать это автоматически.

Добавьте следующие строки в файл `.gitlab-ci.yml`:

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

We use `base_deploy` template and define only unique to stage environment variables such as — environment name and URL. Because of this approach, we get a small and readable job description and as a result — more compact and readable `.gitlab-ci.yml`.

#### Контур production

The production environment is the last and important environment as it is public accessible! We only deploy tags on production environment and only by manual action (maybe at night?).

Добавьте следующие строки в файл `.gitlab-ci.yml`:

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

Обратите внимание на переменную `url` — так как мы деплоим приложение на продуктивный контур, мы знаем публичный URL по которому доступно приложение, и явно указываем его здесь. Далее мы будем также использовать переменную `.Values.global.ci_url` в helm-шаблонах (вспомните описание шаблона `base_deploy` в начале статьи).

### Стадия очистки

В Werf встроен эффективный механизм очистки, который позволяет избежать переполнения registry и диска сборочного узла устаревшими, ненужными образами.
Вы можете более подробно ознакомиться с функционалом очистки, встроенным в Werf, [здесь]({{ site.baseurl }}/documentation/reference/cleaning_process.html).

Как результат работы Werf, мы получаем образы в Docker Registry и сборочный кэш. Сборочный кэш находится только на сборочном узле, а в Docker Registry Werf push'ит только собранные образы.

Для работы очистки, в файле `.gitlab-ci.yml` выделена отдельная стадия — `cleanup`.

Чтобы использовать очистку, необходимо создать `Personal Access Token` в GitLab с необходимыми правами. С помощью данного токена будет осуществляться авторизация в Docker Registry перед очисткой.

Для вашего тестового проекта вы можете просто сделать `Personal Access Token` для вашей учетной записи GitLab. Для этого, откройте страницу `Settings` в GitLab (настройки вашего профиля), затем откройте раздел `Access Token`. Укажите имя токена, в разделе Scope отметьте `api` и нажмите `Create personal access token` — вы получите `Personal Access Token`.

Чтобы передать `Personal Access Token` в переменную окружения GitLab, откройте ваш проект в GitLab, затем откройте `Settings` —> `CI/CD` и разверните `Variables`. Затем, создайте новую переменную окружения `WERF_IMAGES_CLEANUP_PASSWORD` и в качестве ее значений укажите содержимое `Personal Access Token`. Отметьте созданную переменную как `protected`, для безопасности.

Добавьте следующие строки в файл `.gitlab-ci.yml`:

```yaml
Cleanup:
  stage: cleanup
  script:
    - type multiwerf && source <(multiwerf use 1.0 beta)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - docker login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_IMAGES_REPO}
    - werf cleanup --stages-storage :local
  only:
    - schedules
  tags:
    - werf
```

Стадия очистки запускается только по расписанию, которое вы можете определить открыв раздел `CI/CD` —> `Schedules` настроек проекта в GitLab. Нажмите кнопку `New schedule`, заполните описание задания и определите шаблон запуска в обычном cron-формате. В качестве ветки оставьте — master (название ветки не влияет на процесс очистки), отметьте `Active` и сохраните pipeline.

> Как это работает:
   - `werf ci-env` создает временный конфигурационный файл для docker (отдельный при каждом запуске и соответственно для каждого задания GitLab), а также экспортирует переменную окружения DOCKER_CONFIG со значением пути к созданному конфигурационному файлу;
   - `docker login` использует временный конфигурационный файл для docker, путь к которому был указан ранее в переменной окружения DOCKER_CONFIG, для авторизации в registry;
   - `werf cleanup` также использует временный конфигурационный файл из DOCKER_CONFIG.


## Complete .gitlab-ci.yml file

```yaml
stages:
  - build
  - deploy
  - cleanup

Build:
  stage: build
  script:
    - type multiwerf && source <(multiwerf use 1.0 beta)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - werf build-and-publish --stages-storage :local
  tags:
    - werf
  except:
    - schedules

.base_deploy: &base_deploy
  stage: deploy
  script:
    - type multiwerf && source <(multiwerf use 1.0 beta)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    ## Next command makes deploy and will be discussed further
    - werf deploy --stages-storage :local
        --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  ## It is important that the deploy stage depends on the build stage. If the build stage fails, deploy stage should not start.
  dependencies:
    - Build
  tags:
    - werf

Review:
  <<: *base_deploy
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    ## Of course, you need to change domain suffix (`kube.DOMAIN`) of the url if you want to use it in you helm templates.
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
    - type multiwerf && source <(multiwerf use 1.0 beta)
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
    - type multiwerf && source <(multiwerf use 1.0 beta)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - docker login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_IMAGES_REPO}
    - werf cleanup --stages-storage :local
  only:
    - schedules
  tags:
    - werf
```
