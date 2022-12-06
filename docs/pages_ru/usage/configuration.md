---
title: Конфигурация 
permalink: usage/configuration.html
---

## Конфигурация

Конфигурация приложения может включать следующие файлы проекта:

- Конфигурация werf (по умолчанию `werf.yaml`), которая описывает все образы, настройки выката и очистки.
- Шаблоны конфигурации werf (`.werf/**/*.tmpl`).
- Файлы, которые используются с функциями Go-шаблона [.Files.Get]({{ "reference/werf_yaml_template_engine.html#filesget" | true_relative_url }}) и [.Files.Glob]({{ "reference/werf_yaml_template_engine.html#filesglob" | true_relative_url }}).
- Файлы helm-чарта (по умолчанию `.helm`).

> Все файлы конфигурации должны находиться в директории проекта. Симлинки поддерживаются, но ссылка должна указывать на файл в репозитории проекта git

По умолчанию werf запрещает использование определенного набора директив и функций Go-шаблона в конфигурации werf, которые могут потенциально приводить к внешним зависимостям.

## Имя проекта

Имя проекта должно быть уникальным в пределах группы проектов, собираемых на одном сборочном узле и развертываемых на один и тот же кластер Kubernetes (например, уникальным в пределах всех групп одного GitLab).

Имя проекта должно быть не более 50 символов, содержать только строчные буквы латинского алфавита, цифры и знак дефиса.

### Последствия смены имени проекта

**ВНИМАНИЕ**. Никогда не меняйте имя проекта в процессе работы, если вы не осознаете всех последствий.

Смена имени проекта приводит к следующим проблемам:

1. Инвалидация сборочного кэша. Все образы должны быть собраны повторно, а старые удалены из локального хранилища или container registry вручную.
2. Создание совершенно нового Helm-релиза. Смена имени проекта и повторное развертывание приложения приведет к созданию еще одного экземпляра, если вы уже развернули ваше приложение в кластере Kubernetes.

werf не поддерживает изменение имени проекта и все возникающие проблемы должны быть разрешены вручную.

## Директория шаблонов

_Файлы шаблонов_ могут храниться в зарезервированной директории (по умолчанию `.werf`) с расширением `.tmpl` (поддерживается произвольная вложенность `.werf/**/*.tmpl`).

Все _файлы шаблонов_ и основная конфигурация определяют общий контекст:

- _Файл шаблонов_ является полноценным шаблоном и может использоваться функцией [include](#include) по относительному пути ({% raw %}`{{ include "directory/partial.tmpl" . }}`{% endraw %}).
- Шаблон определённый с помощью `define` в одном _файле шаблонов_ доступен в любом другом, включая основную конфигурацию.

### Пример: использование файлов шаблонов

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

### Пример: использование шаблонов, описанных в файле шаблонов

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
