---
title: Интеграция с Github Actions 
sidebar: documentation
permalink: documentation/guides/github_actions_integration.html
author: Sergey Lazarev <sergey.lazarev@flant.com>
---

## Обзор задачи

В статье рассматривается пример настройки CI/CD с использованием Github Actions (self-hosted runner) и werf.

## Требования

* Кластер Kubernetes и настроенный для работы с ним kubectl.
* Аккаунт на github (организация?);
* Docker registry (будете ли вы использовать GitHub Packages или нет решение за вами);
* Приложение, которое успешно собирается и деплоится с помощью werf.

## Инфраструктура

* Кластер Kubernetes.
* Github репозиторий и Docker registry на на базе github (Github packages).
* Узел, на котором установлен werf (узел сборки и деплоя).

Обратите внимание, что процесс сборки и процесс деплоя выполняются на одном и том же узле — пока это единственно верный путь, т.к. распределённая сборка еще не поддерживается [issue](https://github.com/flant/werf/issues/1614). Таким образом, в настоящий момент поддерживается только локальное хранилище стадий — `--stages-storage :local`.

Организовать работу werf внутри Docker-контейнера можно, но мы не поддерживаем данный способ.
Найти информацию по этому вопросу и обсудить можно в [issue](https://github.com/flant/werf/issues/1926).
В данном примере и в целом мы рекомендуем использовать _shell executor_.

Для хранения кэша сборки и служебных файлов werf использует папку `~/.werf`. Папка должна сохраняться и быть доступной на всех этапах pipeline. Это ещё одна из причин по которой мы рекомендуем отдавать предпочтение _shell executor_ вместо эфемерных окружений.

Процесс деплоя требует наличия доступа к кластеру через `kubectl`, поэтому необходимо установить и настроить `kubectl` на узле, с которого будет запускаться werf.
Если не указывать конкретный контекст опцией `--kube-context` или переменной окружения `WERF_KUBE_CONTEXT`, то werf будет использовать контекст `kubectl` по умолчанию.

В конечном счете werf требует наличия доступа:
- к Git-репозиторию кода приложения;
- к Docker registry;
- к кластеру Kubernetes.

### Установка github self-hosted runner.
_Прим.: на текущий момент, вы можете добавить self-hosted runners только к репозиториям. Способность привязать и управлять runners во всей организации будет добавлена в будущих релизах._

1. Создайте репозиторий и запуште код в него.
2. Установите docker и kubectl (если они еще не установлены)
3. Создайте пользователя ?github-runner? и добавьте его в группу docker
  ```shell
  sudo usermod -aG docker github-runner
  ```
4. Install self-hosted runner:
  * On GitHub, navigate to the main page of the repository.
  * Under your repository name, click  Settings - Actions - Add runner (Next to "Self-hosted runners").
  * Select the operating system of your self-hosted runner machine.
  * If you are using Linux, use the drop-down menu to select the architecture of your self-hosted runner machine.
  * You will see instructions showing you how to download the runner application and install it on your self-hosted runner machine. Open a shell on your self-hosted runner machine and run each shell command in the order shown.
5. Configuring the self-hosted runner application as a service:
  * Stop the self-hosted runner application if it is currently running.
  * Install the service with the following command:
  ```shell
  ./svc.sh install
  ```
6. Установим [multiwerf](https://github.com/flant/multiwerf) пользователем `github-runner`:
   ```shell
   sudo su github-runner
   mkdir -p ~/bin
   cd ~/bin
   curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash
   ```
7. Скопируйте файл конфигурации `kubectl` в домашнюю папку пользователя `github-runner`:
  ```shell
  mkdir -p /home/github-runner/.kube &&
  sudo cp -i /etc/kubernetes/admin.conf /home/github-runner/.kube/config &&
  sudo chown -R github-runner:github-runner /home/github-runner/.kube
  ```

После того как self-hosted runner настроен можно переходить к настройке github workflows.

## Pipeline

Файлы описывающие github workflows находятся по пути: `.github/workflows/`

Определим следующие стадии для нашего workflow:
* `build` — стадия сборки образов приложения;
* `deploy` — стадия деплоя приложения для одного из контуров кластера (например, stage, test, review, production или любой другой);
* `cleanup` — стадия очистки хранилища стадий и Docker registry.

### Сборка и публикация образов приложения
_Прим.: Если вы используете werf для шифрования значений переменных, то вам необходимо добавить ваш `encryption key` в переменную `WERF_SECRET_KEY` в env. Можно добавить секрет на странице вашего репозитория в Settings - Secrets - Add a new secret, далее его необходимо добавить в env всей Job или отдельного Steps. Также необходимо будет добавить пароль/токен для доступа к docker registry._

Создайте файл `.github/workflows/build.yml` в корне проекта и добавьте следующие строки:
```yaml
name: Werf build
on: push
jobs:
  build:
    name: "Build on push"
    runs-on: self-hosted
    env:
      WERF_SECRET_KEY: ${{ secrets.WERF_SECRET_KEY }}
    steps:
# как делаем чекаут?
# так
    - name: Checkout
      run: |
        cd ${{ runner.workspace }}/${GITHUB_REPOSITORY##*/}/
        git fetch origin +refs/heads/*:refs/remotes/origin/* +refs/pull/*:refs/remotes/pull/*;  git checkout --force ${{ github.sha }}
 # или так ?
    - name: Checkout
      uses: actions/checkout@v2
 
    - name: Docker login
      env:
        REGISTRY_TOKEN: ${{ secrets.REGISTRY_TOKEN }}
      run: |
        docker login docker.pkg.github.com -u robot -p ${{ secrets.REGISTRY_TOKEN }}
 
    - name: werf build and publish
      run: |
        type multiwerf && source <(multiwerf use 1.0 beta)
        werf build-and-publish --stages-storage :local --images-repo docker.pkg.github.com/<organisation>/<project> --tag-git-branch --tag-git-branch ${GITHUB_REF##*/}
```

### Выкат приложения
Набор контуров в кластере Kubernetes для деплоя приложения зависит от ваших потребностей, но наиболее используемые контуры следующие:
* Контур review. Динамический (временный) контур, используемый разработчиками в процессе работы над приложением для оценки работоспособности написанного кода, первичной оценки работоспособности приложения и т.п. Данный контур удаляется после закрытия Pull request со статусом merged либо вручную.
* Контур test. Тестовый контур, используемый разработчиками после завершения работы над какой-либо задачей для демонстрации результата, более полного тестирования приложения с тестовым набором данных и т.п. В тестовый контур можно вручную деплоить приложения из любой ветки и любой тег.
* Контур stage. Данный контур может использоваться для проведения финального тестирования приложения. Деплой в контур stage производится автоматически после принятия merge-request в ветку master, но это необязательно и вы можете настроить собственную логику.
* Контур production. Финальный контур в pipeline, предназначенный для деплоя готовой к продуктивной эксплуатации версии приложения. Мы подразумеваем, что на данный контур деплоятся только приложения из тегов и только вручную.


> Описанный набор контуров и их функционал — это не правила и вы можете описывать CI/CD процессы под свои нужны.

Для того чтобы деплоить приложение в разные контуры кластера в helm-шаблонах можно использовать переменную .Values.global.env, обращаясь к ней внутри Go-шаблона (Go template). Во время деплоя вы можете указать эту переменную через команду: --env $ENV, предварительно задав эту переменную в описании stage или job вашего workflow.

#### Контур review
Как было сказано выше, контур review — динамический контур (временный контур, контур разработчика) для оценки работоспособности написанного кода, первичной оценки работоспособности приложения и т.п.

Добавим следующие строки в файл `.github/workflows/review.yml`:
```yaml
name: Werf build and deploy review
on:
  pull_request:
    types: [labeled, closed]
jobs:
 
  stub:
    name: Greeting
    runs-on: ubuntu-latest
    steps:
    - name: Greeting
      run: echo "This job is used to prevent the workflow to fail when all other jobs are skipped."
 
  labels:
    name: Working with labels
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request' && ( github.event.label.name == 'review_start' || github.event.label.name == 'review_stop' )
    steps:
    - name: Delete label
      uses: actions/github-script@0.4.0
      with:
        github-token: ${{secrets.GITHUB_TOKEN}}
        script: |
          const eventLabelName = '${{github.event.label.name}}'
          const response = await github.issues.listLabelsOnIssue({...context.issue})
          for (const label of response.data) {
            if (label.name === eventLabelName) {
              github.issues.removeLabel({...context.issue, name: eventLabelName})
              break
            }
          }
 
  review:
    name: "Build and deploy review"
    runs-on: self-hosted
    if: github.event_name == 'pull_request' && github.event.label.name == 'review_start'
    env:
      WERF_SECRET_KEY: ${{ secrets.WERF_SECRET_KEY }}
    steps:
    - name: Checkout
      run: |
        cd ${{ runner.workspace }}/${GITHUB_REPOSITORY##*/}/
        git fetch origin +refs/heads/*:refs/remotes/origin/* +refs/pull/*:refs/remotes/pull/*;  git checkout --force ${{ github.sha }}
    - name: Docker login
      env:
        REGISTRY_TOKEN: ${{ secrets.REGISTRY_TOKEN }}
      run: |
        docker login docker.pkg.github.com -u robot -p ${{ secrets.REGISTRY_TOKEN }}
    - name: werf build and publish
      run: |
        type multiwerf && source <(multiwerf use 1.0 beta)
        werf build-and-publish --stages-storage :local --images-repo docker.pkg.github.com/${GITHUB_REPOSITORY} --tag-git-branch ${GITHUB_HEAD_REF}
    - name: werf deploy
      env:
        ENV: review/${GITHUB_HEAD_REF}
      run: |
        type multiwerf && source <(multiwerf use 1.0 beta)
        werf deploy --stages-storage :local --env $ENV --images-repo docker.pkg.github.com/${GITHUB_REPOSITORY} --tag-git-branch ${GITHUB_HEAD_REF}
 
  stop_review:
    name: "Stop review"
    runs-on: self-hosted
    if: github.event_name == 'pull_request' && ( github.event.label.name == 'review_stop' || ( github.event.action == 'closed' && github.event.pull_request.merged == true ))
    env:
      WERF_SECRET_KEY: ${{ secrets.WERF_SECRET_KEY }}
    steps:
    - name: Checkout
      run: |
        cd ${{ runner.workspace }}/${GITHUB_REPOSITORY##*/}/
        git fetch origin +refs/heads/*:refs/remotes/origin/* +refs/pull/*:refs/remotes/pull/*;  git checkout --force ${{ github.sha }}
    - name: Docker login
      env:
        REGISTRY_TOKEN: ${{ secrets.REGISTRY_TOKEN }}
      run: |
        docker login docker.pkg.github.com -u robot -p ${{ secrets.REGISTRY_TOKEN }}
    - name: werf dismiss
      env:
        ENV: review/${GITHUB_HEAD_REF}
      run: |
        type multiwerf && source <(multiwerf use 1.0 beta)
        werf dismiss --with-namespace --env $ENV
```

Для начала рассмотрим этот кусок нашего workflow:
```yaml
on:
  pull_request:
    types: [labeled, closed]
```
В нем мы задаем событие, при котором наш workflow начнет выполняться.
Таким образом, по событию присоединения лейбла на Pull request или по закрытию Pull request наш workflow приступит к выполнению задач описанных в нем.

В примере выше определены несколько заданий (jobs):
1. `Labels`. В данном задании мы читаем контекст события, которое стартует workflow. Если это было событие присоединения лейбла с названием `review_start` или `review_stop`, то мы удаляем этот лейбл с данного PR. Этот этап мы применяем в качестве индикатора обработки пользовательского запроса на выкат review окружения (а бонусом, мы можем отслеживать историю изменений и выкатов по логу в PR)
2. `Review`. В данном задании устанавливается переменная ENV со значением, использующим переменную окружения github  `review/${GITHUB_HEAD_REF}`.
3. `Stop_review`. В данном задании werf удаляет helm-релиз, и, соответственно, namespace в Kubernetes со всем его содержимым ([werf dismiss]({{ site.baseurl }}/documentation/cli/main/dismiss.html)). Это задание может быть запущено вручную после деплоя на review-контур (путем присоединения лейбла `review_stop`), а также оно может быть запущено по событию закрытия PR’a с последующим мержем.
Задание `Review` не должно запускаться при изменениях в ветке master, т.к. это контур review — контур только для разработчиков.

#### Контур test

Мы не приводим описание заданий деплоя на контур test в настоящем примере, т.к. задания деплоя на контур test очень похожи на задания деплоя на контур stage. Попробуйте описать задания деплоя на контур test по аналогии самостоятельно и надеемся вы получите удовольствие.

#### Контур stage
Как описывалось выше на контур stage можно деплоить только приложение из ветки master и это допустимо делать автоматически.
Добавим следующие строки в файл `.github/workflows/stage.yml`:
```yaml
name: Werf build and deploy stage
on:
  push:
    branches:
      - master
jobs:
  stage:
    name: "Build and deploy stage"
    runs-on: self-hosted
    env:
      WERF_SECRET_KEY: ${{ secrets.WERF_SECRET_KEY }}
    Steps:
 
    - name: Checkout
      run: |
        cd ${{ runner.workspace }}/${GITHUB_REPOSITORY##*/}/
        git fetch origin +refs/heads/*:refs/remotes/origin/* +refs/pull/*:refs/remotes/pull/*;  git checkout --force ${{ github.sha }}
 
    - name: Docker login
      env:
        REGISTRY_TOKEN: ${{ secrets.REGISTRY_TOKEN }}
      run: |
        docker login docker.pkg.github.com -u robot -p ${{ secrets.REGISTRY_TOKEN }}
 
    - name: werf build and publish
      run: |
        type multiwerf && source <(multiwerf use 1.0 beta)
        werf build-and-publish --stages-storage :local --images-repo docker.pkg.github.com/${GITHUB_REPOSITORY} --tag-git-branch master
 
    - name: werf deploy
      env:
        ENV: stage
      run: |
        type multiwerf && source <(multiwerf use 1.0 beta)
        werf deploy --stages-storage :local --env $ENV --images-repo docker.pkg.github.com/${GITHUB_REPOSITORY} --tag-git-branch master
```

#### Контур production
Контур production — последний среди рассматриваемых и самый важный, так как он предназначен для конечного пользователя. Обычно в это окружения выкатываются теги и только вручную.
Добавим следующие строки в файл `.github/workflows/production.yml`:
```yaml
name: Werf build and deploy stage
on:
  push:
    tags:
      - 'v*'
jobs:
  stage:
    name: "Build and deploy stage"
    runs-on: self-hosted
    env:
      WERF_SECRET_KEY: ${{ secrets.WERF_SECRET_KEY }}
    Steps:
 
    - name: Checkout
      run: |
        cd ${{ runner.workspace }}/${GITHUB_REPOSITORY##*/}/
        git fetch origin +refs/heads/*:refs/remotes/origin/* +refs/pull/*:refs/remotes/pull/*;  git checkout --force ${{ github.sha }}
 
    - name: Docker login
      env:
        REGISTRY_TOKEN: ${{ secrets.REGISTRY_TOKEN }}
      run: |
        docker login docker.pkg.github.com -u robot -p ${{ secrets.REGISTRY_TOKEN }}
 
    - name: werf build and publish
      run: |
        type multiwerf && source <(multiwerf use 1.0 beta)
        werf build-and-publish --stages-storage :local --images-repo docker.pkg.github.com/${GITHUB_REPOSITORY} --tag-git-branch ${GITHUB_REF#refs/tags/}
 
    - name: werf deploy
      env:
        ENV: production
      run: |
        type multiwerf && source <(multiwerf use 1.0 beta)
        werf deploy --stages-storage :local --env $ENV --images-repo docker.pkg.github.com/${GITHUB_REPOSITORY} --tag-git-branch ${GITHUB_REF#refs/tags/}
```

### Очистка образов
В werf встроен эффективный механизм очистки, который позволяет избежать переполнения Docker registry и диска сборочного узла от устаревших и неиспользуемых образов. Более подробно ознакомиться с функционалом очистки, встроенным в werf, можно [здесь]({{ site.baseurl }}/documentation/reference/cleaning_process.html).

В результате работы `werf` наполняет локальное хранилище стадий, а также Docker registry собранными образами.

Для работы очистки в мы создадим отдельный workflow — `cleanup`.

Добавим следующие строки в файл `.github/workflows/cleanup.yml`:
```yaml
name: Werf cleanup
 
on:
 schedule:
   - cron: ‘0 1 * * *’
 
jobs:
  cleanup:
    name: "Cleanup"
    runs-on: self-hosted
    env:
      WERF_SECRET_KEY: ${{ secrets.WERF_SECRET_KEY }}
 
    steps:
     - name: Checkout
       run: |
         cd ${{ runner.workspace }}/${GITHUB_REPOSITORY##*/}/
         git fetch origin +refs/heads/*:refs/remotes/origin/* +refs/pull/*:refs/remotes/pull/*;  git checkout --force master
 
    - name: Docker login
      env:
        REGISTRY_TOKEN: ${{ secrets.REGISTRY_TOKEN }}
      run: |
        docker login docker.pkg.github.com -u robot -p ${{ secrets.REGISTRY_TOKEN }}
 
    - name: werf cleanup
      run: |
        type multiwerf && source <(multiwerf use 1.0 beta)
        werf cleanup --stages-storage :local
```

Стадия очистки запускается только по расписанию, которое вы можете определить сами. Настраивается это просто - обычным cron-синтаксисом.
How it works:
* docker login использует временный конфигурационный файл для docker, путь к которому был указан ранее в переменной окружения DOCKER_CONFIG для авторизации в Docker registry;
* werf cleanup также использует временный конфигурационный файл из DOCKER_CONFIG.