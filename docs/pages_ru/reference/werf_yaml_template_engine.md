---
title: Шаблонизатор werf.yaml
permalink: reference/werf_yaml_template_engine.html
---

При чтении конфигурации `werf.yaml`, werf использует встроенный движок шаблонов Go ([text/template](https://golang.org/pkg/text/template)), а также расширяет набор функций с помощью [Sprig](#функции-sprig) и [функций werf](#функции-werf).

При организации конфигурации она может быть разбита на отдельные файлы в [директории шаблонов](#директория-шаблонов).

## Встроенные возможности шаблонизатора Go

Для эффективной работы мы рекомендуем изучить все возможности шаблонизатора или хотя бы ознакомиться со следующими разделами:

- [Actions, основные возможности шаблонизатора](https://golang.org/pkg/text/template/#hdr-Actions).
- [Переменные](https://golang.org/pkg/text/template/#hdr-Variables).
- [Работа с текстом и пробелами](https://golang.org/pkg/text/template/#hdr-Text_and_spaces).
- [Функции](https://golang.org/pkg/text/template/#hdr-Functions).

## Функции Sprig

Библиотека Sprig предоставляет более 70 функций:

- [Строковые функции](http://masterminds.github.io/sprig/strings.html).
- [Математические функции](http://masterminds.github.io/sprig/math.html).
- [Функции кодирования](http://masterminds.github.io/sprig/encoding.html).
- [Работа с данными в формате ключ/значение (dict)](http://masterminds.github.io/sprig/dicts.html).
- [Работа с файловыми путями](http://masterminds.github.io/sprig/paths.html).
- [Остальное](http://masterminds.github.io/sprig/).

werf не поддерживает функцию `expandenv` и имеет свою собственную реализацию для функции [env](#env).

## Функции werf

### Различные окружения

#### .Env

Переменная `.Env` позволяет организовывать конфигурацию для нескольких сред (testing, staging, production и т.п.) и переключаться между ними с помощью опции `--env=<environment_name>`.

> В helm шаблонах таким же образом может использоваться переменная `.Values.werf.env`.

### Информация о текущем коммите

#### .Commit.Hash

{% raw %}`{{ .Commit.Hash }}`{% endraw %} возвращает SHA текущего коммита. Если есть возможность, лучше избегать использование `.Commit.Hash`, т. к. это может приводить к ненужным пересборкам слоев.

#### .Commit.Date

{% raw %}`{{ .Commit.Date.Human }}`{% endraw %} возвращает дату текущего коммита в человекопонятной форме.

{% raw %}`{{ .Commit.Date.Unix }}`{% endraw %} возвращает дату текущего коммита в формате Unix epoch.

##### Пример: пересобирать образ целиком каждый месяц

{% raw %}
```yaml
image: app
from: ubuntu
git:
- add: /
  to: /app
  stageDependencies:
    install:
    - "*"
fromCacheVersion: {{ div .Commit.Date.Unix (mul 60 60 24 30) }}
shell:
  beforeInstall:
  - apt-get update
  - apt-get install -y nodejs npm
  install:
  - npm ci
```
{% endraw %}

### Шаблонизация

#### include

Функция `include` позволяет переиспользовать общие части конфигурации, а также разбивать конфигурацию на несколько файлов.

__Синтаксис__:
{% raw %}
```yaml
{{ include "<TEMPLATE_NAME>" <VALUES> }}
```
{% endraw %}

##### Пример: использование общих кусков конфигурации

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

#### tpl

Функция `tpl` позволяет обрабатывать строки как Go-шаблоны.

__Синтаксис__:
{% raw %}
```yaml
{{ tpl "<STRING>" <VALUES> }}
```
{% endraw %}

- `<STRING>` — контент файла проекта, значение переменной окружения либо произвольная строка.
- `<VALUES>` — значения шаблона. При использовании текущего контекста, `.`, в шаблоне можно использовать все шаблоны и значения (в том числе, описанные в файлах директории шаблонов).

##### Пример: использование файлов проекта в качестве шаблонов werf.yaml

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

### Переменные окружения

#### env

Функция `env` читает переменную окружения.

__Синтаксис__:
{% raw %}
```yaml
{{ env "<ENV_NAME>" }}
{{ env "<ENV_NAME>" "default_value" }}
```
{% endraw %}

> По умолчанию, использование функции `env` запрещено гитерминизмом (подробнее об этом в [статье]({{ "usage/project_configuration/giterminism.html" | true_relative_url }}))

### Файлы проекта

#### .Files.Get

Функция `.Files.Get` получает содержимое определенного файла проекта.

__Синтаксис__:
{% raw %}
```yaml
{{ .Files.Get "<FILE_PATH>" }}
```
{% endraw %}

> По умолчанию, использование файлов, которые имеют незакоммиченные изменения, запрещено гитерминизмом (подробнее об этом в [статье]({{ "usage/project_configuration/giterminism.html" | true_relative_url }}))

##### Пример: использование определённого файла в stapel-образе без использования директивы git (shell)

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

##### Пример: использование определённого файла в stapel-образе без использования директивы git (ansible)

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

#### .Files.Glob

Функция `.Files.Glob` позволяет работать с файлами проекта и их содержимым по глобу. 

Функция поддерживает [shell pattern matching](https://www.gnu.org/software/findutils/manual/html_node/find_html/Shell-Pattern-Matching.html) и `**`. Результаты вызова функции можно объединить, используя sprig-функцию [merge](https://github.com/Masterminds/sprig/blob/master/docs/dicts.md#merge-mustmerge) (к примеру, {% raw %}`{{ $filesDict := merge (.Files.Glob "glob1") (.Files.Glob "glob2") }}`{% endraw %})

__Синтаксис__:
{% raw %}
```yaml
{{ .Files.Glob "<GLOB>" }}
```
{% endraw %}

> По умолчанию, использование файлов, которые имеют незакоммиченные изменения, запрещено гитерминизмом (подробнее об этом в [статье]({{ "usage/project_configuration/giterminism.html" | true_relative_url }}))

##### Пример: использование файлов по глобу в stapel-образе без использования директивы git (shell)

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
{{ range $path, $content := .Files.Glob "modules/*/images/*/{Dockerfile,werf.inc.yaml}" }}
  - |
    head -c -1 <<EOF > /app/{{ base $path }}
{{ $content | indent 4 }}
    EOF
{{ end }}
```
{% endraw %}

##### Пример: использование файлов по глобу в stapel-образе без использования директивы git (ansible)

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
{{ range $path, $content := .Files.Glob "modules/*/images/*/{Dockerfile,werf.inc.yaml}" }}
  - name: "Setup /app/{{ base $path }}"
    copy:
      content: |
{{ $content | indent 8 }}
      dest: /app/{{ base $path }}
{{ end }}
```
{% endraw %}

### Другие

#### required

Функция `required` проверяет наличие значения у определённой переменной. Если значение пусто, то рендеринг шаблона завершится ошибкой с заданным пользователем сообщением.

__Синтаксис__:
{% raw %}
```yaml
value: {{ required "<ERROR_MSG>" <VALUE> }}
```
{% endraw %}

#### fromYaml

Функция `fromYaml` декодирует YAML-документ в структуру.

__Синтаксис__:
{% raw %}
```yaml
value: {{ fromYaml "<STRING>" }}
```
{% endraw %}

##### Пример: чтение YAML-файла и использование полученного значения

{% raw %}
```yaml
{{- $values := .Files.Get "werf_values.yaml" | fromYaml -}} # или fromYaml (.Files.Get "werf_values.yaml")
from: {{- $values.image.from }}
```
{% endraw %}

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
