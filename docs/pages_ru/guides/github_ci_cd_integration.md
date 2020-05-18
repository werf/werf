---
title: Интеграция с GitHub Actions 
sidebar: documentation
permalink: documentation/guides/github_ci_cd_integration.html
author: Sergey Lazarev <sergey.lazarev@flant.com>, Alexey Igrychev <alexey.igrychev@flant.com>
---

## Обзор задачи

В статье рассматриваются различные варианты настройки CI/CD с использованием GitLab CI/CD и werf.

Рабочий процесс в репозитории (набор GitHub workflow конфигураций) будет строиться на базе следующих заданий:
* `build-publish-deploy` — задание сборки, публикации образов и выката приложения для одного из контуров кластера;
* `dismiss` — задание удаления приложения для review окружения;
* `cleanup` — задание очистки хранилища стадий и Docker registry.

Набор контуров в кластере Kubernetes может варьироваться в зависимости от многих факторов.
В статье будут приведены различные варианты организации окружений для следующих:
* [Контур production]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#production).
* [Контур staging]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#staging).
* [Контур review]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#review).

Далее последовательно рассматриваются задания и различные варианты их организации. Изложение построено от общего к частному. В конце статьи приведён [полный набор конфигураций для готовых workflow](#полный-набор-конфигураций-для-готовых-workflow).

Независимо от workflow, все версии конфигураций подчиняются следующим правилам:
* [*Выкат/удаление* review окружений](#варианты-организации-review-окружения):
  * *Выкат* на review окружение возможен только в рамках Pull Request (PR).
  * Review окружения удаляются автоматически при закрытии PR.
* [*Очистка*](#очистка-образов) запускается один раз в день по расписанию на master.

Для выкатов review окружения и staging и production окружений предложены самые популярные варианты по организации. Каждый вариант для staging и production окружений сопровождается всевозможными способами отката релиза в production.

> С общей информацией по организации CI/CD с помощью werf, а также информацией по конструированию своего workflow, можно ознакомиться в [общей статье]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html)
>

## Требования

* Кластер Kubernetes.
* Проект на [GitHub](https://github.com/).
* Приложение, которое успешно собирается и деплоится с помощью werf.
* Понимание [основных концептов GitHub Actions](https://help.github.com/en/actions/getting-started-with-github-actions/core-concepts-for-github-actions).
  
> Далее в примерах статьи будут использоваться виртуальные машины, предоставляемые GitHub, с OS Linux (`runs-on: ubuntu-latest`). Тем не менее, все примеры также справедливы для предустановленных self-hosted runners на базе любой OS

## Сборка, публикация образов и выкат приложения

Прежде всего необходимо описать некий шаблон, общую часть для выката в любой контур, что позволит сосредоточиться далее на правилах выката и предложенных workflow.

{% raw %}
```yaml
build-publish-deploy:
  name: Build, Publish and Deploy
  runs-on: ubuntu-latest
  steps:
  
    - name: Checkout code
      uses: actions/checkout@v2
      with:
        fetch-depth: 0
  
    - name: Install multiwerf
      run: |
        # add ~/bin into PATH
        export PATH=$PATH:$HOME/bin
        echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
  
        # install multiwerf into ~/bin directory
        mkdir -p ~/bin
        cd ~/bin
        curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash
  
        echo "::add-path::$HOME/bin"
  
    - name: Create kube config
      run: |
        KUBECONFIG=$(mktemp -d)/config
        echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
        echo ::set-env name=KUBECONFIG::$KUBECONFIG
      env:
        BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}
  
    - name: Build and Publish
      run: |
        . $(multiwerf use 1.1 alpha --as-file)
        . $(werf      ci-env github --as-file --verbose)
        werf build-and-publish
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

    - name: Deploy
      run: |
        . $(multiwerf use 1.1 alpha --as-file)
        . $(werf      ci-env github --as-file --verbose)
        werf deploy --set "global.ci_url=$WERF_ENV_URL"
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        WERF_ENV: ANY_ENV_NAME
        WERF_ENV_URL: ANY_ENV_URL
```
{% endraw %}

> Данное задание можно разбить на два независимых, но в нашем случае (сборка и публикации вызывается не на каждый коммит, а используется только совместно с выкатом) это избыточно и ухудшит читаемость конфигурации и время выполнения.
>    
> <div class="details">
> <a href="javascript:void(0)" class="details__summary">build-and-publish и deploy задания</a>
> <div class="details__content" markdown="1">
> 
> {% raw %}
> ```yaml
> build-and-publish:
>   name: Build and Publish
>   runs-on: ubuntu-latest
>   steps:
>   
>     - name: Checkout code
>       uses: actions/checkout@v2
>       with:
>         fetch-depth: 0
>   
>     - name: Install multiwerf
>       run: |
>         # add ~/bin into PATH
>         export PATH=$PATH:$HOME/bin
>         echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
>   
>         # install multiwerf into ~/bin directory
>         mkdir -p ~/bin
>         cd ~/bin
>         curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash
>   
>         echo "::add-path::$HOME/bin"
>   
>     - name: Create kube config
>       run: |
>         KUBECONFIG=$(mktemp -d)/config
>         echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
>         echo ::set-env name=KUBECONFIG::$KUBECONFIG
>       env:
>         BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}
>   
>     - name: Build and Publish
>       run: |
>         . $(multiwerf use 1.1 alpha --as-file)
>         . $(werf      ci-env github --as-file --verbose)
>         werf build-and-publish
>       env:
>         GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
> 
> deploy:
>   name: Deploy
>   needs: build-and-publish
>   runs-on: ubuntu-latest
>   steps:
>   
>     - name: Checkout code
>       uses: actions/checkout@v2
>       with:
>         fetch-depth: 0
>   
>     - name: Install multiwerf
>       run: |
>         # add ~/bin into PATH
>         export PATH=$PATH:$HOME/bin
>         echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
>   
>         # install multiwerf into ~/bin directory
>         mkdir -p ~/bin
>         cd ~/bin
>         curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash
>   
>         echo "::add-path::$HOME/bin"
>   
>     - name: Create kube config
>       run: |
>         KUBECONFIG=$(mktemp -d)/config
>         echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
>         echo ::set-env name=KUBECONFIG::$KUBECONFIG
>       env:
>         BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}
> 
>     - name: Deploy
>       run: |
>         . $(multiwerf use 1.1 alpha --as-file)
>         . $(werf      ci-env github --as-file --verbose)
>         werf deploy --set "global.ci_url=$WERF_ENV_URL"
>       env:
>         GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
>         WERF_ENV: ANY_ENV_NAME
>         WERF_ENV_URL: ANY_ENV_URL
> ```
> {% endraw %}
> 
> </div>
> </div>

Первый шаг, с которого начинается задание — `Checkout code`, добавление исходных кодов приложения. При использовании сборщика werf (основная особенность которого — инкрементальная сборка) недостаточно, так называемого, `shallow clone` с один коммитом, который создаёт action `actions/checkout@v2` при использовании без параметров. werf создаёт стадии, базируясь на истории git, поэтому без истории, каждая сборка будет проходить без ранее собранных стадий. Поэтому, крайне важно, использовать параметр `fetch-depth: 0` для доступа ко всей истории при сборке, публикации (`werf build-and-publish`), выкате (`werf deploy`) и запуске (`werf run`). Т.е. для всех команд, которые используют стадии. 

{% raw %}
```yaml
- name: Checkout code
  uses: actions/checkout@v2
  with:
    fetch-depth: 0
```
{% endraw %}

Среди предустановленного ПО на виртуальных машинах GitHub уже установлен `kubectl`, поэтому пользователю остаётся только:
* определиться с конфигурацией [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/);
* создать секретную переменную `BASE64_KUBECONFIG` с контентом файла kubeconfig (`cat ~/.kube/config | base64 -w0`) в Settings/Secrets проекта на GitHub.
* добавить шаг `Create kube config` c созданием файла kubeconfig в каждом [job](https://help.github.com/en/actions/getting-started-with-github-actions/core-concepts-for-github-actions#job), где он требуется:
    
{% raw %}  
```yaml
- name: Create kube config
  run: |
    KUBECONFIG=$(mktemp -d)/config
    echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
    echo ::set-env name=KUBECONFIG::$KUBECONFIG
  env:
    BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}
```
{% endraw %}

Конфигурация задания достаточно проста, поэтому хочется сделать акцент на том, чего в ней нет — явной авторизации в Docker registry, вызова `docker login`. 

В простейшем случае, при использовании встроенной [Github Packages имплементации Docker registry](https://help.github.com/en/packages/using-github-packages-with-your-projects-ecosystem/configuring-docker-for-use-with-github-packages), авторизация выполняется автоматически при вызове команды `werf ci-env`. В качестве необходимых аргументов используются переменные окружения GitHub, [секретная переменная `GITHUB_TOKEN`](https://help.github.com/en/actions/configuring-and-managing-workflows/authenticating-with-the-github_token#about-the-github_token-secret), которую необходимо явно добавлять в шаг, а также имя пользователя (`GITHUB_ACTOR`) инициировавшего запуск workflow.

Если необходимо выполнить авторизацию с произвольными учётными данными или с внешним Docker registry, то необходимо использовать готовый action для вашей имплементации или просто выполнить `docker login`. 

Далее рассмотрим шаг `Deploy` с выкатом:
{% raw %}
```yaml    
- name: Deploy
  run: |
    . $(multiwerf use 1.1 alpha --as-file)
    . $(werf      ci-env github --as-file --verbose)
    werf deploy --set "global.ci_url=$WERF_ENV_URL"
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    WERF_ENV: ANY_ENV_NAME
    WERF_ENV_URL: ANY_ENV_URL
```
{% endraw %}

Для каждого контура необходимо определить окружение. В нашем случае оно определяется следующими параметрами: 
* именем (`WERF_ENV`) и; 
* URL (`WERF_ENV_URL`).             

Для того, чтобы по-разному конфигурировать приложение для используемых контуров кластера в helm-шаблонах можно использовать Go-шаблоны и переменную `.Values.global.env`, что соответствует значению опции `--env` или переменной окружения `WERF_ENV`.

Адрес окружения, URL для доступа к разворачиваемому в контуре приложению, который передаётся параметром `global.ci_url`, может использоваться в helm-шаблонах, например, для конфигурации Ingress-ресурсов.

> Если для шифрования значений переменных вы используете werf, то вам также необходимо добавить `encryption key` в переменную `WERF_SECRET_KEY` в Settings/Secrets проекта

Далее будут представлены популярные стратегии и практики, на базе которых мы предлагаем выстраивать ваши процессы в GitHub Actions. 

### Варианты организации review окружения

Как уже было сказано ранее, review окружение — это динамический контур, поэтому помимо выката, у этого окружения также будет и очистка.

Рассмотрим базовые GitHub workflow файлы, которые лягут в основу всех предложенных вариантов.

Сначала разберём файл `.github\workflows\review_deployment.yml`.

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Define environment url
        run: |
          pr_id=${{ github.event.number }}
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}-${pr_id}.kube.DOMAIN

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: review-${{ github.event.number }}
```
{% endraw %}

</div>
</div>

В нём пропущено условие запуска, т.к. оно зависит от выбранного варианта организации.

От базовой конфигурации задание отличается только появившимся шагом `Define environment url`. 
На этом шаге генерируется уникальный URL, по которому после выката будет доступно наше приложение (при соответствующей организации helm templates). 

```yaml
- name: Define environment url
  run: |
    pr_id=${{ github.event.number }}
    github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
    echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}-${pr_id}.kube.DOMAIN
```

Далее файл `.github\workflows\review_deployment_dismiss.yml`.

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment_dismiss.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment Dismiss
on:
  pull_request:
    types: [closed]
jobs:

  dismiss:
    name: Dismiss
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Optional Dismiss
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)

          release_name=$(werf helm get-release)
          if werf helm get $release_name 2>/dev/null; then
            werf dismiss --with-namespace
          else
            echo "Release ${release_name} was not found"
          fi
        env:
          WERF_ENV: review-${{ github.event.number }}
```
{% endraw %}

</div>
</div>

Данный GitHub workflow будет выполняться при закрытии PR.

```yaml
on:
  pull_request:
    types: [closed]
```

На шаге `Optional Dismiss` выполняется удаление review-релиза: werf удаляет helm-релиз и namespace в Kubernetes со всем его содержимым ([werf dismiss]({{ site.baseurl }}/documentation/cli/main/dismiss.html)).

{% raw %}
```yaml
- name: Optional Dismiss
  run: |
    . $(multiwerf use 1.1 alpha --as-file)
    . $(werf      ci-env github --as-file --verbose)

    release_name=$(werf helm get-release)
    if werf helm get $release_name 2>/dev/null; then
      werf dismiss --with-namespace
    else
      echo "Release ${release_name} was not found"
    fi
  env:
    WERF_ENV: review-${{ github.event.number }}
```
{% endraw %}

Далее разберём основные стратегии при организации выката review окружения. 

> Мы не ограничиваем вас предложенными вариантами, даже напротив — рекомендуем комбинировать их и создавать конфигурацию workflow под нужды вашей команды

#### №1 Вручную

> Данный вариант реализует подход описанный в разделе [Выкат на review из pull request по кнопке]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#выкат-на-review-из-pull-request-по-кнопке)

При таком подходе пользователь выкатывает и удаляет окружение, проставляя соответствующий лейбл (`review_start` или `review_stop`) в PR.

Он самый простой и может быть удобен в случае, когда выкаты происходят редко и review окружение не используется при разработке.
По сути, для проверки перед принятием PR. 

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment
on:
  pull_request:
    types: [labeled]
jobs:

  labels:
    name: Label taking off
    runs-on: ubuntu-latest
    if: github.event.label.name == 'review_start'
    steps:
      - name: Take off label
        uses: actions/github-script@0.4.0
        with:
          github-token: ${{secrets.GITHUB_TOKEN}}
          script: "github.issues.removeLabel({...context.issue, name: github.event.label.name})"

  build-publish-deploy:
    name: Build, Publish and Deploy
    if: github.event.label.name == 'review_start'
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Define environment
        run: |
          pr_id=${{ github.event.number }}
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}-${pr_id}.kube.DOMAIN

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: review-${{ github.event.number }}
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment_dismiss.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment Dismiss
on:
  pull_request:
    types: [labeled, closed]
jobs:

  labels:
    name: Label taking off 
    runs-on: ubuntu-latest
    if: github.event.label.name == 'review_stop'
    steps:
      - name: Take off label
        uses: actions/github-script@0.4.0
        with:
          github-token: ${{secrets.GITHUB_TOKEN}}
          script: "github.issues.removeLabel({...context.issue, name: github.event.label.name})"

  dismiss:
    name: Dismiss
    if: github.event.label.name == 'review_stop' || github.event.action == 'closed'
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Optional Dismiss
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)

          release_name=$(werf helm get-release)
          if werf helm get $release_name 2>/dev/null; then
            werf dismiss --with-namespace
          else
            echo "Release ${release_name} was not found"
          fi
        env:
          WERF_ENV: review-${{ github.event.number }}
```
{% endraw %}

</div>
</div>

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

```yaml
labels:
  name: Label taking off 
  runs-on: ubuntu-latest
  if: github.event.label.name == 'review_stop'
  steps:
    - name: Take off label
      uses: actions/github-script@0.4.0
      with:
        github-token: ${{secrets.GITHUB_TOKEN}}
        script: "github.issues.removeLabel({...context.issue, name: github.event.label.name})"
```

#### №2 Автоматически по имени ветки

> Данный вариант реализует подход описанный в разделе [Выкат на review из ветки по шаблону автоматически]({{ site.baseurl }}/documentation/reference/ci_cd_workflow_overview.html#выкат-на-review-из-ветки-по-шаблону-автоматически)

В предложенном ниже варианте автоматический релиз выполняется для каждого коммита в PR, в случае, если имя git-ветки содержит `review`. 

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment
on:
  pull_request:
    types:
      - opened
      - reopened
      - synchronize
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    if: ${{ contains( github.head_ref, 'review' ) }}
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Define environment
        run: |
          pr_id=${{ github.event.number }}
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}-${pr_id}.kube.DOMAIN

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: review-${{ github.event.number }}
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment_dismiss.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment Dismiss
on:
  pull_request:
    types: [closed]
jobs:

  dismiss:
    name: Dismiss
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Optional Dismiss
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)

          release_name=$(werf helm get-release)
          if werf helm get $release_name 2>/dev/null; then
            werf dismiss --with-namespace
          else
            echo "Release ${release_name} was not found"
          fi
        env:
          WERF_ENV: review-${{ github.event.number }}
```
{% endraw %}

</div>
</div>

Выкат инициируется при коммите в ветку, открытии и переоткрытии PR, что соответствует набору событие по умолчанию для `pull_request`:
```yaml
on:
  pull_request:
    types:
      - opened
      - reopened
      - synchronize

// == 

on:
  pull_request:
```

#### №3 Полуавтоматический режим с лейблом (рекомендованный)

> Данный вариант реализует подход описанный в разделе [Выкат на review из pull request автоматически после ручной активации]({{ site.baseurl }}/documentation/reference/ci_cd_workflow_overview.html#выкат-на-review-из-pull-request-автоматически-после-ручной-активации)

Полуавтоматический режим с лейблом — это комплексное решение, объединяющие первые два варианта. 

При проставлении специального лейбла, в примере ниже `review`, пользователь активирует автоматический выкат в review окружения для каждого коммита. 
При снятии лейбла автоматически запускается удаление review-релиза.    

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\optional_review_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Optional Review Deployment
on:
  pull_request:
    types:
      - labeled
      - unlabeled
      - synchronize
jobs:

  deploy_dismiss:
    name: Optional Deploy/Dismiss
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        if contains( github.event.issue.labels.*.name, 'review' )

      - name: Define environment url
        run: |
          pr_id=${{ github.event.number }}
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}-${pr_id}.kube.DOMAIN
        if contains( github.event.issue.labels.*.name, 'review' )

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: review-${{ github.event.number }}
        if contains( github.event.issue.labels.*.name, 'review' )

      - name: Optional Dismiss
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)

          release_name=$(werf helm get-release)
          if werf helm get $release_name 2>/dev/null; then
            werf dismiss --with-namespace
          else
            echo "Release ${release_name} was not found"
          fi
        env:
          WERF_ENV: review-${{ github.event.number }}
        if !contains( github.event.issue.labels.*.name, 'review' )
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment_dismiss.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment Dismiss
on:
  pull_request:
    types: [closed]
jobs:

  dismiss:
    name: Dismiss
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Optional Dismiss
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)

          release_name=$(werf helm get-release)
          if werf helm get $release_name 2>/dev/null; then
            werf dismiss --with-namespace
          else
            echo "Release ${release_name} was not found"
          fi
        env:
          WERF_ENV: review-${{ github.event.number }}
```
{% endraw %}

</div>
</div>

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

В нашем случае, данные окружения являются определяющими, поэтому названия вариантов соответствуют названиям окончательных workflow, предложенных в [конце статьи](#полный-gitlab-ciyml-для-различных-workflow). 

#### №1 Fast and Furious (рекомендованный)

> Данный вариант реализует подходы описанные в разделах [Выкат на production из master автоматически]({{ site.baseurl }}/documentation/reference/ci_cd_workflow_overview.html#выкат-на-production-из-master-автоматически) и [Выкат на production-like из pull request по кнопке]({{ site.baseurl }}/documentation/reference/ci_cd_workflow_overview.html#выкат-на-production-like-из-pull-request-по-кнопке)

Выкат в **production** происходит автоматически при любых изменениях в master. Выполнить выкат в **staging** можно по кнопке в MR.

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\staging_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Staging Deployment
on:
  pull_request:
    types: [labeled]
jobs:

  labels:
    name: Label taking off
    runs-on: ubuntu-latest
    if: github.event.label.name == 'staging_deploy'
    steps:
      
      - name: Take off label
        uses: actions/github-script@0.4.0
        with:
          github-token: ${{secrets.GITHUB_TOKEN}}
          script: "github.issues.removeLabel({...context.issue, name: github.event.label.name})"

  build-publish-deploy:
    name: Build, Publish and Deploy
    if: github.event.label.name == 'staging_deploy'
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Define environment url
        run: |
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}.kube.DOMAIN

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: staging
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\production_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Production Deployment
on:
  push:
    branches: [master]
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: production
          WERF_ENV_URL: https://www.company.org
```
{% endraw %}

</div>
</div>

Варианты отката изменений в production:
- [revert изменений](https://git-scm.com/docs/git-revert) в master (**рекомендованный**);
- выкат стабильного MR или воспользовавшись кнопкой [Rollback](https://docs.gitlab.com/ee/ci/environments.html#what-to-expect-with-a-rollback).

#### №2 Push the Button

> Данный вариант реализует подходы описанные в разделах [Выкат на production из master по кнопке]({{ site.baseurl }}/documentation/reference/ci_cd_workflow_overview.html#выкат-на-production-из-master-по-кнопке) и [Выкат на staging из master автоматически]({{ site.baseurl }}/documentation/reference/ci_cd_workflow_overview.html#выкат-на-staging-из-master-автоматически)

Выкат **production** осуществляется по кнопке у коммита в master, а выкат в **staging** происходит автоматически при любых изменениях в master.

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\staging_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Staging Deployment
on:
  push:
    branches: [master]
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Define environment url
        run: |
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}.kube.DOMAIN

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: staging
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\production_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Production Deployment
on:
  repository_dispatch:
    types: [production_deployment]
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: production
          WERF_ENV_URL: https://www.company.org
```
{% endraw %}

</div>
</div>

Варианты отката изменений в production:
- по кнопке у стабильного коммита или воспользовавшись кнопкой [Rollback](https://docs.gitlab.com/ee/ci/environments.html#what-to-expect-with-a-rollback) (**рекомендованный**);
- выкат стабильного MR и нажатии кнопки.

#### №3 Tag everything (рекомендованный)

> Данный вариант реализует подходы описанные в разделах [Выкат на production из тега автоматически]({{ site.baseurl }}/documentation/reference/ci_cd_workflow_overview.html#выкат-на-production-из-тега-автоматически) и [Выкат на staging из master по кнопке]({{ site.baseurl }}/documentation/reference/ci_cd_workflow_overview.html#выкат-на-staging-из-master-по-кнопке)

Выкат в **production** выполняется при проставлении тега, а в **staging** по кнопке у коммита в master.

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\staging_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Staging Deployment
on:
  repository_dispatch:
    types: [staging_deployment]
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Define environment url
        run: |
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}.kube.DOMAIN

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: staging
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\production_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Production Deployment
on:
  push:
    tags:
    - v*
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: production
          WERF_ENV_URL: https://www.company.org
```
{% endraw %}

</div>
</div>

Варианты отката изменений в production:
- нажатие кнопки на другом теге (**рекомендованный**);
- создание нового тега на старый коммит (так делать не надо).

#### №4 Branch, branch, branch!

> Данный вариант реализует подходы описанные в разделах [Выкат на production из ветки автоматически]({{ site.baseurl }}/documentation/reference/ci_cd_workflow_overview.html#выкат-на-production-из-ветки-автоматически) и [Выкат на production-like из ветки автоматически]({{ site.baseurl }}/documentation/reference/ci_cd_workflow_overview.html#выкат-на-production-like-из-ветки-автоматически)

Выкат в **production** происходит автоматически при любых изменениях в ветке production, а в **staging** при любых изменениях в ветке master.

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\staging_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Staging Deployment
on:
  push:
    branches:
    - master
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Define environment url
        run: |
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}.kube.DOMAIN

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: staging
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\production_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Production Deployment
on:
  push:
    branches:
    - production
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: production
          WERF_ENV_URL: https://www.company.org
```
{% endraw %}

</div>
</div>

Варианты отката изменений в production:
- воспользовавшись кнопкой [Rollback](https://docs.gitlab.com/ee/ci/environments.html#what-to-expect-with-a-rollback);
- [revert изменений](https://git-scm.com/docs/git-revert) в ветке production;
- [revert изменений](https://git-scm.com/docs/git-revert) в master и fast-forward merge в ветку production;
- удаление коммита из ветки production и push-force.

## Очистка образов

{% raw %}
```yaml
name: Cleanup Docker registry
on:
  schedule:
    - cron:  '0 6 * * *'
  repository_dispatch:
    types: [cleanup]

jobs:
  cleanup:
    name: Cleanup
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Fetch all history for all tags and branches
        run: git fetch --prune --unshallow

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Cleanup
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf cleanup
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_REPO_GITHUB_TOKEN: ${{ secrets.WERF_GITHUB_TOKEN }}
```
{% endraw %}

Первый шаг, с которого начинается задание — `Checkout code`, добавление исходных кодов приложения. Большинство политик очистки в werf базируется на примитивах git (на коммите, ветке и теге), поэтому использование action `actions/checkout@v2` без дополнительных параметров и действий может приводить к неожиданному удалению образов. Мы рекомендуем использовать следующие шаги для корректной работы. 

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
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_github_ci_2')">№2 Push the Button</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_github_ci_3')">№3 Tag everything (рекомендованный)</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_github_ci_4')">№4 Branch, branch, branch!</a>
</div>

<div id="complete_github_ci_1" class="tabs__content no_toc_section active" markdown="1">

### Детали workflow
{:.no_toc}

> Подробнее про workflow можно почитать в отдельной [статье]({{ site.baseurl }}/documentation/reference/ci_cd_workflow_overview.html#1-fast-and-furious)

* Выкат на review контур по стратегии [№3 Полуавтоматический режим с лейблом (рекомендованный)](#3-полуавтоматический-режим-с-лейблом-рекомендованный).
* Выкат на staging и production контуры осуществляется по стратегии [№1 Fast and Furious (рекомендованный)](#1-fast-and-furious-рекомендованный).
* [Очистка стадий](#очистка-образов).

### Конфигурации
{:.no_toc}

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\optional_review_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Optional Review Deployment
on:
  pull_request:
    types:
      - labeled
      - unlabeled
      - synchronize
jobs:

  deploy_dismiss:
    name: Optional Deploy/Dismiss
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        if contains( github.event.issue.labels.*.name, 'review' )

      - name: Define environment url
        run: |
          pr_id=${{ github.event.number }}
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}-${pr_id}.kube.DOMAIN
        if contains( github.event.issue.labels.*.name, 'review' )

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: review-${{ github.event.number }}
        if contains( github.event.issue.labels.*.name, 'review' )

      - name: Optional Dismiss
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)

          release_name=$(werf helm get-release)
          if werf helm get $release_name 2>/dev/null; then
            werf dismiss --with-namespace
          else
            echo "Release ${release_name} was not found"
          fi
        env:
          WERF_ENV: review-${{ github.event.number }}
        if !contains( github.event.issue.labels.*.name, 'review' )
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment_dismiss.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment Dismiss
on:
  pull_request:
    types: [closed]
jobs:

  dismiss:
    name: Dismiss
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Optional Dismiss
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)

          release_name=$(werf helm get-release)
          if werf helm get $release_name 2>/dev/null; then
            werf dismiss --with-namespace
          else
            echo "Release ${release_name} was not found"
          fi
        env:
          WERF_ENV: review-${{ github.event.number }}
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\staging_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Staging Deployment
on:
  pull_request:
    types: [labeled]
jobs:

  labels:
    name: Label taking off
    runs-on: ubuntu-latest
    if: github.event.label.name == 'staging_deploy'
    steps:
      
      - name: Take off label
        uses: actions/github-script@0.4.0
        with:
          github-token: ${{secrets.GITHUB_TOKEN}}
          script: "github.issues.removeLabel({...context.issue, name: github.event.label.name})"

  build-publish-deploy:
    name: Build, Publish and Deploy
    if: github.event.label.name == 'staging_deploy'
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Define environment url
        run: |
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}.kube.DOMAIN

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: staging
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\production_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Production Deployment
on:
  push:
    branches: [master]
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: production
          WERF_ENV_URL: https://www.company.org
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\docker_registry_cleanup.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Docker Registry Cleanup 
on:
  schedule:
    - cron:  '0 6 * * *'
  repository_dispatch:
    types: [cleanup]

jobs:
  cleanup:
    name: Cleanup
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Fetch all history for all tags and branches
        run: git fetch --prune --unshallow

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Cleanup
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf cleanup
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_REPO_GITHUB_TOKEN: ${{ secrets.WERF_GITHUB_TOKEN }}
```
{% endraw %}

</div>
</div>

</div>

<div id="complete_github_ci_2" class="tabs__content no_toc_section" markdown="1">

### Детали workflow
{:.no_toc}

> Подробнее про workflow можно почитать в отдельной [статье]({{ site.baseurl }}/documentation/reference/ci_cd_workflow_overview.html#2-push-the-button)

* Выкат на review контур по стратегии [№1 Вручную](#1-вручную).
* Выкат на staging и production контуры осуществляется по стратегии [№2 Push the Button](#2-push-the-button).
* [Очистка стадий](#очистка-образов). 

### Конфигурации
{:.no_toc}

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment
on:
  pull_request:
    types: [labeled]
jobs:

  labels:
    name: Label taking off
    runs-on: ubuntu-latest
    if: github.event.label.name == 'review_start'
    steps:
      - name: Take off label
        uses: actions/github-script@0.4.0
        with:
          github-token: ${{secrets.GITHUB_TOKEN}}
          script: "github.issues.removeLabel({...context.issue, name: github.event.label.name})"

  build-publish-deploy:
    name: Build, Publish and Deploy
    if: github.event.label.name == 'review_start'
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Define environment
        run: |
          pr_id=${{ github.event.number }}
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}-${pr_id}.kube.DOMAIN

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: review-${{ github.event.number }}
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment_dismiss.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment Dismiss
on:
  pull_request:
    types: [labeled, closed]
jobs:

  labels:
    name: Label taking off 
    runs-on: ubuntu-latest
    if: github.event.label.name == 'review_stop'
    steps:
      - name: Take off label
        uses: actions/github-script@0.4.0
        with:
          github-token: ${{secrets.GITHUB_TOKEN}}
          script: "github.issues.removeLabel({...context.issue, name: github.event.label.name})"

  dismiss:
    name: Dismiss
    if: github.event.label.name == 'review_stop' || github.event.action == 'closed'
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Optional Dismiss
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)

          release_name=$(werf helm get-release)
          if werf helm get $release_name 2>/dev/null; then
            werf dismiss --with-namespace
          else
            echo "Release ${release_name} was not found"
          fi
        env:
          WERF_ENV: review-${{ github.event.number }}
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\staging_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Staging Deployment
on:
  push:
    branches: [master]
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Define environment url
        run: |
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}.kube.DOMAIN

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: staging
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\production_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Production Deployment
on:
  repository_dispatch:
    types: [production_deployment]
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: production
          WERF_ENV_URL: https://www.company.org
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\docker_registry_cleanup.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Docker Registry Cleanup 
on:
  schedule:
    - cron:  '0 6 * * *'
  repository_dispatch:
    types: [cleanup]

jobs:
  cleanup:
    name: Cleanup
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Fetch all history for all tags and branches
        run: git fetch --prune --unshallow

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Cleanup
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf cleanup
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_REPO_GITHUB_TOKEN: ${{ secrets.WERF_GITHUB_TOKEN }}
```
{% endraw %}

</div>
</div>

</div>

<div id="complete_github_ci_3" class="tabs__content no_toc_section" markdown="1">

### Детали workflow
{:.no_toc}

> Подробнее про workflow можно почитать в отдельной [статье]({{ site.baseurl }}/documentation/reference/ci_cd_workflow_overview.html#3-tag-everything)

* Выкат на review контур по стратегии [№1 Вручную](#1-вручную).
* Выкат на staging и production контуры осуществляется по стратегии [№3 Tag everything](#3-tag-everything-рекомендованный).
* [Очистка стадий](#очистка-образов). 

### Конфигурации
{:.no_toc}

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment
on:
  pull_request:
    types: [labeled]
jobs:

  labels:
    name: Label taking off
    runs-on: ubuntu-latest
    if: github.event.label.name == 'review_start'
    steps:
      - name: Take off label
        uses: actions/github-script@0.4.0
        with:
          github-token: ${{secrets.GITHUB_TOKEN}}
          script: "github.issues.removeLabel({...context.issue, name: github.event.label.name})"

  build-publish-deploy:
    name: Build, Publish and Deploy
    if: github.event.label.name == 'review_start'
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Define environment
        run: |
          pr_id=${{ github.event.number }}
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}-${pr_id}.kube.DOMAIN

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: review-${{ github.event.number }}
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment_dismiss.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment Dismiss
on:
  pull_request:
    types: [labeled, closed]
jobs:

  labels:
    name: Label taking off 
    runs-on: ubuntu-latest
    if: github.event.label.name == 'review_stop'
    steps:
      - name: Take off label
        uses: actions/github-script@0.4.0
        with:
          github-token: ${{secrets.GITHUB_TOKEN}}
          script: "github.issues.removeLabel({...context.issue, name: github.event.label.name})"

  dismiss:
    name: Dismiss
    if: github.event.label.name == 'review_stop' || github.event.action == 'closed'
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Optional Dismiss
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)

          release_name=$(werf helm get-release)
          if werf helm get $release_name 2>/dev/null; then
            werf dismiss --with-namespace
          else
            echo "Release ${release_name} was not found"
          fi
        env:
          WERF_ENV: review-${{ github.event.number }}
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\staging_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Staging Deployment
on:
  repository_dispatch:
    types: [staging_deployment]
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Define environment url
        run: |
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}.kube.DOMAIN

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: staging
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\production_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Production Deployment
on:
  push:
    tags:
    - v*
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: production
          WERF_ENV_URL: https://www.company.org
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\docker_registry_cleanup.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Docker Registry Cleanup 
on:
  schedule:
    - cron:  '0 6 * * *'
  repository_dispatch:
    types: [cleanup]

jobs:
  cleanup:
    name: Cleanup
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Fetch all history for all tags and branches
        run: git fetch --prune --unshallow

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Cleanup
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf cleanup
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_REPO_GITHUB_TOKEN: ${{ secrets.WERF_GITHUB_TOKEN }}
```
{% endraw %}

</div>
</div>

</div>

<div id="complete_github_ci_4" class="tabs__content no_toc_section" markdown="1">

### Детали workflow
{:.no_toc}

> Подробнее про workflow можно почитать в отдельной [статье]({{ site.baseurl }}/documentation/reference/ci_cd_workflow_overview.html#4-branch-branch-branch)

* Выкат на review контур по стратегии [№2 Автоматически по имени ветки](#2-автоматически-по-имени-ветки).
* Выкат на staging и production контуры осуществляется по стратегии [№4 Branch, branch, branch!](#4-branch-branch-branch).
* [Очистка стадий](#очистка-образов). 

### Конфигурации
{:.no_toc}

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment
on:
  pull_request:
    types:
      - opened
      - reopened
      - synchronize
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    if: ${{ contains( github.head_ref, 'review' ) }}
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Define environment
        run: |
          pr_id=${{ github.event.number }}
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}-${pr_id}.kube.DOMAIN

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: review-${{ github.event.number }}
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\review_deployment_dismiss.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Review Deployment Dismiss
on:
  pull_request:
    types: [closed]
jobs:

  dismiss:
    name: Dismiss
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Optional Dismiss
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)

          release_name=$(werf helm get-release)
          if werf helm get $release_name 2>/dev/null; then
            werf dismiss --with-namespace
          else
            echo "Release ${release_name} was not found"
          fi
        env:
          WERF_ENV: review-${{ github.event.number }}
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\staging_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Staging Deployment
on:
  push:
    branches:
    - master
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Define environment url
        run: |
          github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
          echo ::set-env name=WERF_ENV_URL::http://${github_repository_id}.kube.DOMAIN

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: staging
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\production_deployment.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Production Deployment
on:
  push:
    branches:
    - production
jobs:

  build-publish-deploy:
    name: Build, Publish and Deploy
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Build and Publish
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf build-and-publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Deploy
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf deploy --set "global.ci_url=$WERF_ENV_URL"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_ENV: production
          WERF_ENV_URL: https://www.company.org
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.github\workflows\docker_registry_cleanup.yml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
name: Docker Registry Cleanup 
on:
  schedule:
    - cron:  '0 6 * * *'
  repository_dispatch:
    types: [cleanup]

jobs:
  cleanup:
    name: Cleanup
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Fetch all history for all tags and branches
        run: git fetch --prune --unshallow

      - name: Install multiwerf
        run: |
          # add ~/bin into PATH
          export PATH=$PATH:$HOME/bin
          echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

          # install multiwerf into ~/bin directory
          mkdir -p ~/bin
          cd ~/bin
          curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash

          echo "::add-path::$HOME/bin"

      - name: Create kube config
        run: |
          KUBECONFIG=$(mktemp -d)/config
          echo $BASE64_KUBECONFIG | base64 -d -w0 > $KUBECONFIG
          echo ::set-env name=KUBECONFIG::$KUBECONFIG
        env:
          BASE64_KUBECONFIG: ${{ secrets.BASE64_KUBECONFIG }}

      - name: Cleanup
        run: |
          . $(multiwerf use 1.1 alpha --as-file)
          . $(werf      ci-env github --as-file --verbose)
          werf cleanup
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WERF_REPO_GITHUB_TOKEN: ${{ secrets.WERF_GITHUB_TOKEN }}
```
{% endraw %}

</div>
</div>

</div>
