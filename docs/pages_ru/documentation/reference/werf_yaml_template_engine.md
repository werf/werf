---
title: Шаблонизатор werf.yaml
sidebar: documentation
permalink: documentation/reference/werf_yaml_template_engine.html
---

При чтении конфигурации `werf.yaml`, werf использует встроенный движок шаблонов Go ([text/template](https://golang.org/pkg/text/template)), а также расширяет набор функций с помощью [Sprig](#функции-sprig) и [функций werf](#функции-werf).

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
{{ tpl <STRING> <VALUES> }}
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

Функция `env` читает переменную окружения. Переменная окружения должна быть установлена, но значение может быть пустым.

__Синтаксис__:
{% raw %}
```yaml
{{ env <ENV_NAME> }}
```
{% endraw %}

> По умолчанию, использование функции `env` запрещено гитерминизмом (подробнее об этом в [статье]({{ "documentation/advanced/giterminism.html" | true_relative_url }}))

### Файлы проекта

#### .Files.Get

Функция `.Files.Get` получает содержимое определенного файла проекта.

__Синтаксис__:
{% raw %}
```yaml
{{ .Files.Get <FILE_PATH> }}
```
{% endraw %}

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
{{ .Files.Glob <GLOB> }}
```
{% endraw %}

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