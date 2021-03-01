---
title: werf.yaml template engine
sidebar: documentation
permalink: documentation/reference/werf_yaml_template_engine.html
---

Reading _the werf configuration_, werf uses a built-in Go template engine ([text/template](https://golang.org/pkg/text/template)) and expands the function set with [Sprig](#sprig-functions) and [werf functions](#werf-functions).

## Built-in Go template features

To work effectively, we recommend you looking at all the features, or at least the following sections:

- [Actions, the core of the engine](https://golang.org/pkg/text/template/#hdr-Actions).
- [Variables](https://golang.org/pkg/text/template/#hdr-Variables).
- [Text and spaces](https://golang.org/pkg/text/template/#hdr-Text_and_spaces).
- [Functions](https://golang.org/pkg/text/template/#hdr-Functions).

## Sprig functions

The Sprig library provides over 70 template functions:

- [String Functions](http://masterminds.github.io/sprig/strings.html).
- [Integer Math Functions](http://masterminds.github.io/sprig/math.html).
- [Encoding Functions](http://masterminds.github.io/sprig/encoding.html).
- [Dictionaries and Dict Functions](http://masterminds.github.io/sprig/dicts.html).
- [Path and Filepath Functions](http://masterminds.github.io/sprig/paths.html).
- [Others](http://masterminds.github.io/sprig/).

Among all functions, werf does not support the `expandenv` function and has its own implementation for the [env](#env) function.

## werf functions

### templating

#### include

The `include` function brings in another template, and then pass the results to other template functions.

__Syntax__:
{% raw %}
```yaml
{{ include "<TEMPLATE_NAME>" <VALUES> }}
```
{% endraw %}

##### Example: how to use a common configuration

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

The `tpl` function allows evaluating string as a template inside a template.

__Syntax__:
{% raw %}
```yaml
{{ tpl <STRING> <VALUES> }}
```
{% endraw %}

- `<STRING>` — the content of a project file, an environment variable value, or an arbitrary string.
- `<VALUES>` — the template values. If you use the current context, `.`, all templates and values (including those described in [the template directory](#template-directory) files) can be used in the template.

##### Example: how to use project files as the werf configuration partials

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

### environment variables

#### env

The `env` function reads an environment variable. The environment variable must be set but the value can be empty.

__Syntax__:
{% raw %}
```yaml
{{ env <ENV_NAME> }}
```
{% endraw %}

> By default, the use of the `env` function is not allowed by giterminism (read more about it [here]({{ "documentation/advanced/giterminism.html" | true_relative_url }}))

### project files

#### .Files.Get

The function `.Files.Get` gets a certain project file content.

__Syntax__:
{% raw %}
```yaml
{{ .Files.Get <FILE_PATH> }}
```
{% endraw %}

##### Example: how to add a certain file to stapel image without git directive (shell)

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

##### Example: how to add a certain file to stapel image without git directive (ansible)

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

The function `.Files.Glob` allows getting project files with a glob and working with their content. 

The function supports [shell pattern matching](https://www.gnu.org/software/findutils/manual/html_node/find_html/Shell-Pattern-Matching.html) and `**`. The function results can be merged with the [merge](https://github.com/Masterminds/sprig/blob/master/docs/dicts.md#merge-mustmerge) sprig function (e.g., {% raw %}`{{ $filesDict := merge (.Files.Glob "glob1") (.Files.Glob "glob2")`{% endraw %}).

__Syntax__:
{% raw %}
```yaml
{{ .Files.Glob <GLOB> }}
```
{% endraw %}

##### Example: how to add files by a glob to stapel image without git directive (shell)

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

##### Example: how to add files by a glob to stapel image without git directive (ansible)

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

### others

#### required

The `required` function declares a particular values entry as required for template rendering. If the value is empty, the template rendering will fail with a user submitted error message.

__Syntax__:
{% raw %}
```yaml
value: {{ required "<ERROR_MSG>" <VALUE> }}
```
{% endraw %}

#### fromYaml

The `fromYaml` functions decodes a YAML document into a structure.

__Syntax__:
{% raw %}
```yaml
value: {{ fromYaml "<STRING>" }}
```
{% endraw %}

##### Example: how to read YAML file and then use a value

{% raw %}
```yaml
{{- $values := .Files.Get "werf_values.yaml" | fromYaml -}} # or fromYaml (.Files.Get "werf_values.yaml")
from: {{- $values.image.from }}
```
{% endraw %}