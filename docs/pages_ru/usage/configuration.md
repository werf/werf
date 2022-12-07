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

# Гитерминизм

werf вводит такое понятие как _гитерминизм_, которое происходит от совмещения слов `git` и `determinism`, что можно понимать как режим "детерминированный гитом".

Для обеспечения надежности и гарантии воспроизводимости, werf читает конфигурационные и файлы сборочного контекста из текущего коммита репозитория проекта, а также исключает внешние зависимости. Таким образом, werf стремится к легко воспроизводимым конфигурациям и позволяет разработчикам в Windows, macOS и Linux использовать образы, только что собранные в CI-системе, просто переключившись на желаемый коммит.

werf не позволяет работать с незакоммиченными и неотслеживаемыми файлами. Если файлы вам не требуются, то их следует явно исключить с помощью файла `.gitignore` (включая глобальный) или `.helmignore`.

> Мы настоятельно рекомендуем следовать этому подходу, но при необходимости вы можете явно ослабить ограничения гитерминизма, а также включить функциональность, требующую осмысленного использования, с [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }})

При отладке и разработке, изменение файлов проекта может доставлять неудобства за счёт необходимости создания промежуточных коммитов. Мы работаем над режимом разработки, чтобы упростить этот процесс и в то же время оставить всю логику работы неизменной.
В текущих версиях, режим разработки (активируется опцией `--dev`) позволяет работать с состоянием worktree git-репозитория проекта, с отслеживаемыми (tracked) и неотслеживаемыми (untracked) файлами. werf игнорирует изменения с учётом правил, описанных в `.gitignore`, а также правил, заданных пользователем опцией `--dev-ignore=<glob>` (может использоваться несколько раз).

### Конфигурация werf

#### CLI

##### Опция --use-custom-tag

Использование алиасов тегов с неизменяемыми значениями (например, `%image%-master`) делает предыдущие выкаты невоспроизводимыми и требует указания политики `imagePullPolicy: Always` для каждого образа при конфигурации контейнеров приложения в helm-чарте.

Для активации опции `--use-custom-tag` необходимо использовать [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }}), но мы рекомендуем еще раз подумать о возможных последствиях.

#### Функции Go-шаблонизатора

##### env

Использование функции [env]({{ "reference/werf_yaml_template_engine.html#env-1" | true_relative_url }}) усложняет совместное использование и воспроизводимость конфигурации в заданиях CI и среди разработчиков, поскольку значение переменной среды влияет на окончательный дайджест собираемых образов и значение должно быть идентичным на всех этапах CI-пайплайна и во время локальной разработки при воспроизведении.

Для активации функции `env` необходимо использовать [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }}), но мы рекомендуем еще раз подумать о возможных последствиях.

#### Dockerfile-образ

##### contextAddFiles

Использование директивы [contextAddFiles]({{ "reference/werf_yaml.html#contextaddfiles" | true_relative_url }}) усложняет совместное использование и воспроизводимость конфигурации в заданиях CI и среди разработчиков, поскольку данные файла влияют на окончательный дайджест собираемых образов и должны быть идентичными на всех этапах CI-пайплайна и во время локальной разработки при воспроизведении.

Для активации директивы `contextAddFiles` необходимо использовать [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }}), но мы рекомендуем еще раз подумать о возможных последствиях.

#### Stapel-образ

##### fromLatest

При использовании [fromLatest]({{ "advanced/building_images_with_stapel/base_image.html#from-fromlatest" | true_relative_url }}), werf начинает учитывать реальный дайджест _базового образа_ в дайджесте собираемого образа. Таким образом, использование этой директивы может нарушить воспроизводимость предыдущих сборок. Изменение базового образа в container registry делает непригодными для использования все ранее собранные образы.

* Предыдущие задания CI-пайплайна (например, converge) не могут быть выполнены повторно без пересборки образа после изменения базового образа в container registry.
* Изменение базового образа в container registry приводит к неожиданным падениям CI-пайплайна. Например, изменение происходит после успешной сборки, и следующие задания не могут выполниться из-за изменения дайджеста конечных образов вместе с дайджестом базового образа.

В качестве альтернативы мы рекомендуем использовать неизменяемый тег или периодически изменять значение [fromCacheVersion]({{ "advanced/building_images_with_stapel/base_image.html#fromcacheversion" | true_relative_url }}), чтобы гарантировать управляемый и предсказуемый жизненный цикл приложения.

Для активации директивы `fromLatest` необходимо использовать [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }}), но мы рекомендуем еще раз подумать о возможных последствиях.

##### git

###### branch

Использование ветки (по умолчанию ветка `master`) при работе с [произвольными git-репозиториями]({{ "advanced/building_images_with_stapel/git_directive.html#работа-с-удаленными-репозиториями" | true_relative_url }}) может нарушить воспроизводимость предыдущих сборок. werf использует историю git-репозитория для расчета дайджеста собираемого образа. Таким образом, новый коммит в ветке делает непригодным для использования все ранее собранные образы.

* Существующие задания CI-пайплайна (например, converge) не будут выполняться и потребуют пересборки образа, если HEAD ветки был изменён.
* Неконтролируемые коммиты в ветку могут приводить к сбоям CI-пайплайна без видимых причин. Например, изменения могут произойти после успешного завершения процесса сборки. В этом случае связанные задания CI-пайплайна завершатся с ошибкой из-за изменений в дайджестах собираемых образов вместе с HEAD связанной ветки.

В качестве альтернативы мы рекомендуем использовать неизменяемую ссылку, тег или коммит, чтобы гарантировать управляемый и предсказуемый жизненный цикл приложения.

Для активации директивы `branch` необходимо использовать [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }}), но мы рекомендуем еще раз подумать о возможных последствиях.

##### mount

###### build_dir

Использование монтирования [build_dir]({{ "advanced/building_images_with_stapel/mount_directive.html" | true_relative_url }}) может приводить к непредсказуемому поведению при параллельном использовании и потенциально повлиять на воспроизводимость и надежность.

Для активации директивы `build_dir` необходимо использовать [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }}), но мы рекомендуем еще раз подумать о возможных последствиях.

###### fromPath

Использование монтирования [fromPath]({{ "advanced/building_images_with_stapel/mount_directive.html" | true_relative_url }}) может приводить к непредсказуемому поведению при параллельном использовании и потенциально повлиять на воспроизводимость и надежность. Данные в директории монтирования не влияют на окончательный дайджест собираемого образа, что может привести к невалидным образам, а также трудно отслеживаемым проблемам.

Для активации директивы `fromPath` необходимо использовать [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }}), но мы рекомендуем еще раз подумать о возможных последствиях.

## Сборочный контекст

Контекст сборки Dockerfile-образа — это файлы `context` (подробнее о директиве [context]({{ "reference/werf_yaml.html" | true_relative_url }})) текущего коммита репозитория проекта.

Контекст сборки Stapel-образа — это все файлы, добавляемые с помощью директивы [git]({{ "advanced/building_images_with_stapel/git_directive.html" | true_relative_url }}), из текущего коммита репозитория проекта.
