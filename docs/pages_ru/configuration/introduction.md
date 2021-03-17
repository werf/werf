---
title: Общие сведения
sidebar: documentation
permalink: configuration/introduction.html
author: Alexey Igrychev <alexey.igrychev@flant.com>, Timofey Kirillov <timofey.kirillov@flant.com>
---

## Что такое конфигурация werf?

Для использования werf в приложении должна быть описана конфигурация, которая включает в себя:

1. Определение мета-информации проекта. Например, имени проекта, которое будет впоследствии влиять на результат сборки, деплоя и другие команды.
2. Определение списка образов проекта и инструкций их сборки.

Конфигурация werf хранится в YAML-файле `werf.yaml` в корневой папке проекта (приложения), и представляет собой набор секций конфигурации -- частей YAML-файла, разделенных [тремя дефисами](http://yaml.org/spec/1.2/spec.html#id2800132):

```yaml
CONFIG_SECTION
---
CONFIG_SECTION
---
CONFIG_SECTION
```

Каждая секция конфигурации может быть одного из трех типов:

1. Секция для описания мета-информации проекта, далее — *секция мета-информации*.
2. Секция описания инструкций сборки образа, далее — *секция образа*.
3. Секция описания инструкций сборки артефакта, далее — *секция артефакта*.

В будущем, возможно, количество типов увеличится.

### Секция мета-информации

```yaml
project: PROJECT_NAME
configVersion: CONFIG_VERSION
OTHER_FIELDS
---
```

Секция мета-информации, **обязательная** секция конфигурации, содержащая ключи `project: PROJECT_NAME` и `configVersion: CONFIG_VERSION`. 
В каждом файле конфигурации `werf.yaml` должна быть только одна секция мета-информации.

Директивы `deploy` и `cleanup` вынесены в отдельные статьи, [развертывание в Kubernetes]({{ site.baseurl }}/configuration/deploy_into_kubernetes.html) и [политики очистки]({{ site.baseurl }}/configuration/cleanup.html).

#### Имя проекта

Ключ `project` определяет уникальное имя проекта вашего приложения. 
Имя проекта влияет на имена образов в _stages storage_, namespace в Kubernetes, имя Helm-релиза и зависящие от него имена (смотри подробнее про [развертывание в Kubernetes]({{ site.baseurl }}/configuration/deploy_into_kubernetes.html)).
Ключ `project` — единственное обязательное поле секции мета-информации.

Имя проекта должно быть уникальным в пределах группы проектов, собираемых на одном сборочном узле и развертываемых на один и тот же кластер Kubernetes (например, уникальным в пределах всех групп одного GitLab).

Имя проекта должно быть не более 50 символов, содержать только строчные буквы латинского алфавита, цифры и знак дефиса.

**ВНИМАНИЕ**. Никогда не меняйте имя проекта в процессе работы, если вы не осознаете всех последствий.

Смена имени проекта приводит к следующим проблемам:
1. Инвалидация сборочного кэша. Все образы должны быть собраны повторно, а старые удалены из локального хранилища или Docker registry вручную.
2. Создание совершенно нового Helm-релиза. Смена имени проекта и повторное развертывание приложения приведет к созданию еще одного экземпляра, если вы уже развернули ваше приложение в кластере Kubernetes.

werf не поддерживает изменение имени проекта и все возникающие проблемы должны быть решены вручную.

#### Версии конфигурации

Директива `configVersion` определяет формат файла `werf.yaml`. 
В настоящее время, это всегда — `1`.

### Секция образа

В каждой секции образа содержатся инструкции, описывающие правила сборки одного независимого образа. 
В одном файле конфигурации `werf.yaml` может быть произвольное количество секций образов.

Секция образа определяется по наличию в секции конфигурации ключа `image: IMAGE_NAME`, где `IMAGE_NAME` — короткое имя Docker-образа. 
Это имя должно быть уникальным в пределах одного файла конфигурации `werf.yaml`.

```yaml
image: IMAGE_NAME_1
OTHER_FIELDS
---
image: IMAGE_NAME_2
OTHER_FIELDS
---
...
---
image: IMAGE_NAME_N
OTHER_FIELDS
```

### Секция артефакта

Секция артефакта также содержит инструкции, описывающие правила сборки одного независимого образа артефактов. 
Образ артефакта — это вспомогательный образ, предназначенный для изолирования сборочного окружения, инструментов и данных, от конечного образа приложения (подробнее [здесь]({{ site.baseurl }}/configuration/stapel_artifact.html)).
В одном файле конфигурации `werf.yaml` может быть произвольное количество секций артефактов.

Секция артефакта определяется по наличию в секции конфигурации ключа `artifact: IMAGE_NAME`, где `IMAGE_NAME` — произвольное имя, на которое можно ссылаться из других секций конфигурации. 
Это имя должно быть уникальным в пределах одного файла конфигурации `werf.yaml`.

### Пример минимальной конфигурации

```yaml
project: my-project
configVersion: 1
```

## Описание конфигурации в нескольких файлах 

### При использовании директории шаблонов

Часть конфигурации может быть вынесена в ***отдельные файлы шаблонов*** и затем включаться в __werf.yaml__. 
_Файлы шаблонов_ должны размещаться в папке ***.werf*** и иметь расширение **.tmpl** (поддерживается вложенность).

> **Совет:** вы можете скачать или сгенерировать шаблоны до запуска werf. 
> Например, это может быть удобно при вынесении общей логики нескольких проектов в общие шаблоны

werf обрабатывает все файлы в одном окружении, поэтому [define](#include) в одном шаблоне, доступен в любом другом, включая сам _werf.yaml_.

<div class="details active">
<a href="javascript:void(0)" class="details__summary">werf.yaml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
{{ $_ := set . "RubyVersion" "2.3.4" }}
{{ $_ := set . "BaseImage" "alpine" }}

project: my-project
configVersion: 1
---

image: rails
from: {{ .BaseImage }}
ansible:
  beforeInstall:
  {{- include "(component) mysql client" . }}
  {{- include "(component) ruby" . }}
  install:
  {{- include "(component) Gemfile dependencies" . }}
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.werf/ansible/components.tmpl</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
{{- define "(component) Gemfile dependencies" }}
  - file:
      path: /root/.ssh
      state: directory
      owner: root
      group: root
      recurse: yes
  - name: "Setup ssh known_hosts used in Gemfile"
    shell: |
      set -e
      ssh-keyscan github.com >> /root/.ssh/known_hosts
      ssh-keyscan mygitlab.myorg.com >> /root/.ssh/known_hosts
    args:
      executable: /bin/bash
  - name: "Install Gemfile dependencies with bundler"
    shell: |
      set -e
      source /etc/profile.d/rvm.sh
      cd /app
      bundle install --without development test --path vendor/bundle
    args:
      executable: /bin/bash
{{- end }}

{{- define "(component) mysql client" }}
  - name: "Install mysql client"
    apt:
      name: "{{`{{ item }}`}}"
      update_cache: yes
    with_items:
    - libmysqlclient-dev
    - mysql-client
    - g++
{{- end }}

{{- define "(component) ruby" }}
  - command: gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3
  - get_url:
      url: https://raw.githubusercontent.com/rvm/rvm/master/binscripts/rvm-installer
      dest: /tmp/rvm-installer
  - name: "Install rvm"
    command: bash -e /tmp/rvm-installer
  - name: "Install ruby {{ .RubyVersion }}"
    raw: bash -lec {{`{{ item | quote }}`}}
    with_items:
    - rvm install {{ .RubyVersion }}
    - rvm use --default {{ .RubyVersion }}
    - gem install bundler --no-ri --no-rdoc
    - rvm cleanup all
{{- end }}
```
{% endraw %}

</div>
</div>

> При существовании нескольких шаблонов с одинаковым именем, werf будет использовать шаблон определенный в _werf.yaml_, либо последний описанный в _файлах шаблонов_

Если нужно обратиться непосредственно к _файлу шаблона_, например, в функции [include](#include), то можно использовать путь к соответствующему файлу относительно папки _.werf_.

<div class="details active">
<a href="javascript:void(0)" class="details__summary">werf.yaml</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
project: my-project
configVersion: 1
---
image: app
from: java:8-jdk-alpine
shell:
  beforeInstall:
  - mkdir /app
  - adduser -Dh /home/gordon gordon
import:
- artifact: appserver
  add: '/usr/src/atsea/target/AtSea-0.0.1-SNAPSHOT.jar'
  to: '/app/AtSea-0.0.1-SNAPSHOT.jar'
  after: install
- artifact: storefront
  add: /usr/src/atsea/app/react-app/build
  to: /static
  after: install
docker:
  ENTRYPOINT: ["java", "-jar", "/app/AtSea-0.0.1-SNAPSHOT.jar"]
  CMD: ["--spring.profiles.active=postgres"]
---
{{ include "artifact/appserver.tmpl" . }}
---
{{ include "artifact/storefront.tmpl" . }}
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.werf/artifact/appserver.tmpl</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
artifact: appserver
from: maven:latest
git:
- add: '/app'
  to: '/usr/src/atsea'
shell:
  install:
  - cd /usr/src/atsea
  - mvn -B -f pom.xml -s /usr/share/maven/ref/settings-docker.xml dependency:resolve
  - mvn -B -s /usr/share/maven/ref/settings-docker.xml package -DskipTests
```
{% endraw %}

</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">.werf/artifact/storefront.tmpl</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
artifact: storefront
from: node:latest
git:
- add: /app/react-app
  to: /usr/src/atsea/app/react-app
shell:
  install:
  - cd /usr/src/atsea/app/react-app
  - npm install
  - npm run build
```
{% endraw %}

</div>
</div>

### При использовании функции tpl

Функция `tpl` позволяет обрабатывать строки, как Go шаблоны в `werf.yaml`, передавая их в переменных окружения либо в файлах проекта. Таким образом, шаблоны могут располагаться в произвольном месте в проекте и добавляться в `werf.yaml`.

Стоит отметить, что шаблоны выполняются в общем контексте, поэтому в них доступны все шаблоны и значения из `werf.yaml` и [директории шаблонов](#при-использовании-директории-шаблонов). 

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'tpl_werf_yaml')">werf.yaml</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'tpl_backend')">backend/werf-partial.yaml</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'tpl_frontend')">frontend/werf-partial.yaml</a>
</div>

<div id="tpl_werf_yaml" class="tabs__content active" markdown="1">

{% raw %}
```yaml
{{ $_ := set . "BaseImage" "node:14.3" }}

project: app
configVersion: 1
---

{{ range $path, $content := .Files.Glob "**/werf-partial.yaml" }}
{{ tpl $content $ }}
{{ end }}

{{- define "common install commands" }}
- npm install
- npm run build
{{- end }}
```
{% endraw %}

</div>

<div id="tpl_backend" class="tabs__content" markdown="1">

{% raw %}
```yaml
image: backend
from: {{ .BaseImage }}
git:
- add: /backend
  to: /app/backend
shell:
  install:
  - cd /app/backend
{{- include "common install commands" . | indent 2 }}
```
{% endraw %}

</div>

<div id="tpl_frontend" class="tabs__content" markdown="1">

{% raw %}
```yaml
image: frontend
from: {{ .BaseImage }}
git:
- add: /frontend
  to: /app/frontend
shell:
  install:
  - cd /app/frontend
{{- include "common install commands" . | indent 2 }}
```
{% endraw %}

</div>

## Этапы обработки конфигурации

Следующие шаги описывают как происходит обработка конфигурации:
1. Чтение файла `werf.yaml` и файлов шаблонов из папки `.werf`.
2. Выполнение Go-шаблонов.
3. Сохранение отрендеренного файла, получившегося после применения Go-шаблонов, в `.werf.render.yaml` (файл остаётся после выполнения команды и удаляется автоматически с помощью процедуры GC).
4. Разбиение отрендеренного YAML-файла на секции (они отделяются друг от друга тремя дефисами, смотри подробнее [спецификацию](https://yaml.org/spec/1.2/spec.html#id2800132)).
5. Валидация каждой секции конфигурации:
   * Валидация YAML-синтаксиса (смотрите подробнее [здесь](http://yaml.org/refcard.html)).
   * Валидация werf-синтаксиса.
6. Генерация описанных в файле конфигураций образов.

### Шаблоны Go

В конфигурации возможно использование шаблонов Go (Go templates). Поддерживаются следующие функции:

* [Встроенные функции шаблонов Go](https://golang.org/pkg/text/template/#hdr-Functions) и некоторые особенности языка. Например, использование переменных:<a id="go-templates" href="#go-templates" class="anchorjs-link " aria-label="Anchor link for: go templates" data-anchorjs-icon=""></a>

  {% raw %}
  ```yaml
  {{ $base_image := "golang:1.11-alpine" }}

  project: my-project
  configVersion: 1
  ---

  image: gogomonia
  from: {{ $base_image }}
  ---
  image: saml-authenticator
  from: {{ $base_image }}
  ```
  {% endraw %}

* [Sprig-функции](http://masterminds.github.io/sprig/). Например, использование переменных окружения:<a id="sprig-functions" href="#sprig-functions" class="anchorjs-link " aria-label="Anchor link for: sprig functions" data-anchorjs-icon=""></a>

  {% raw %}
  ```yaml
  project: my-project
  configVersion: 1
  ---

  {{ $_ := env "SPECIFIC_ENV_HERE" | set . "GitBranch" }}

  image: ~
  from: alpine
  git:
  - url: https://github.com/company/project1.git
    branch: {{ .GitBranch }}
    add: /
    to: /app/project1
  - url: https://github.com/company/project2.git
    branch: {{ .GitBranch }}
    add: /
    to: /app/project2
  ```
  {% endraw %}

* Функция `include` и конструкция `define` для переиспользования частей конфигурации:<a id="include" href="#include" class="anchorjs-link " aria-label="Anchor link for: include" data-anchorjs-icon=""></a>

  {% raw %}
  ```yaml
  project: my-project
  configVersion: 1
  ---

  image: app1
  from: alpine
  ansible:
    beforeInstall:
    {{- include "(component) ruby" . }}
  ---
  image: app2
  from: alpine
  ansible:
    beforeInstall:
    {{- include "(component) ruby" . }}

  {{- define "(component) ruby" }}
    - command: gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3
    - get_url:
        url: https://raw.githubusercontent.com/rvm/rvm/master/binscripts/rvm-installer
        dest: /tmp/rvm-installer
    - name: "Install rvm"
      command: bash -e /tmp/rvm-installer
    - name: "Install ruby 2.3.4"
      raw: bash -lec {{`{{ item | quote }}`}}
      with_items:
      - rvm install 2.3.4
      - rvm use --default 2.3.4
      - gem install bundler --no-ri --no-rdoc
      - rvm cleanup all
  {{- end }}
  ```
  {% endraw %}

* Функция `tpl` позволяет обрабатывать строки, как Go шаблоны в `werf.yaml`, передавая их в переменных окружения либо в файлах проекта: [раннее описанный пример с файлами проекта](#при-использовании-функции-tpl).<a id="tpl" href="#tpl" class="anchorjs-link " aria-label="Anchor link for: tpl" data-anchorjs-icon=""></a>

* Функции `.Files.Get` и `.Files.Glob` для работы с файлами проекта:<a id="files-get" href="#files-get" class="anchorjs-link " aria-label="Anchor link for: .Files.Get" data-anchorjs-icon=""></a>

  <div class="tabs">
    <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'ansible')">Ansible</a>
    <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'shell')">Shell</a>
  </div>

  <div id="ansible" class="tabs__content active" markdown="1">
   
  **.Files.Get**

  {% raw %}
  ```yaml
  project: my-project
  configVersion: 1
  ---

  image: app
  from: alpine
  ansible:
   setup:
   - name: "Setup /etc/nginx/nginx.conf"
     copy:
       content: |
  {{ .Files.Get ".werf/nginx.conf" | indent 8 }}
       dest: /etc/nginx/nginx.conf
  ```
  {% endraw %}

  **.Files.Glob**

  {% raw %}
    > Функция поддерживает [shell pattern matching](https://www.gnu.org/software/findutils/manual/html_node/find_html/Shell-Pattern-Matching.html) + `**`. Результаты вызова функции можно объединить со [sprig функцией `merge`](https://github.com/Masterminds/sprig/blob/master/docs/dicts.md#merge-mustmerge) (к примеру, `{{ $filesDict := merge (.Files.Glob "*/*.txt") (.Files.Glob "app/**/*.txt") }}`)

  {% endraw %}
  
  {% raw %}
  ```yaml
  project: my-project
  configVersion: 1
  ---

  image: app
  from: alpine
  ansible:
   install:
   - raw: mkdir /app
   setup:
  {{ range $path, $content := .Files.Glob ".werf/files/*" }}
   - name: "Setup /app/{{ base $path }}"
     copy:
       content: |
  {{ $content | indent 8 }}
       dest: /app/{{ base $path }}
  {{ end }}
  ```
  {% endraw %}
  </div>

  <div id="shell" class="tabs__content" markdown="1">

  **.Files.Get**

  {% raw %}
  ```yaml
  project: my-project
  configVersion: 1
  ---

  image: app
  from: alpine
  shell:
    setup:
    - |
      head -c -1 <<'EOF' > /etc/nginx/nginx.conf
   {{ .Files.Get ".werf/nginx.conf" | indent 4 }}
      EOF
  ```
  {% endraw %}

  **.Files.Glob**

  {% raw %}
  > Функция поддерживает [shell pattern matching](https://www.gnu.org/software/findutils/manual/html_node/find_html/Shell-Pattern-Matching.html) + `**`. Результаты вызова функции можно объединить со [sprig функцией `merge`](https://github.com/Masterminds/sprig/blob/master/docs/dicts.md#merge-mustmerge) (к примеру, `{{ $filesDict := merge (.Files.Glob "*/*.txt") (.Files.Glob "app/**/*.txt") }}`)

  {% endraw %}

  {% raw %}
  ```yaml
  project: my-project
  configVersion: 1
  ---

  image: app
  from: alpine
  shell:
   install: mkdir /app
   setup:
  {{ range $path, $content := .Files.Glob ".werf/files/*" }}
   - |
     head -c -1 <<EOF > /app/{{ base $path }}
  {{ $content | indent 4 }}
     EOF
  {{ end }}
  ```
  {% endraw %}

  </div>
