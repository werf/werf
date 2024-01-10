---
title: werf.yaml template engine
permalink: reference/werf_yaml_template_engine.html
---

Reading _the werf configuration_, werf uses a built-in Go template engine ([text/template](https://golang.org/pkg/text/template)) and expands the function set with [Sprig](#sprig-functions) and [werf functions](#werf-functions).

When organizing the configuration, it can be split into separate files in [template directory](#template-directory).

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

### various environments

#### .Env

The `.Env` variable allows organizing configuration for several environments (testing, production, staging, and so on) and switching between them by the `--env=<environment_name>` option.

> In helm templates, there is the `.Values.werf.env` variable that can be used the same way

### current commit information

#### .Commit.Hash

{% raw %}`{{ .Commit.Hash }}`{% endraw %} provides current commit SHA. It's best to avoid using `.Commit.Hash` if possible, since it might trigger many unneeded stage rebuilds.

#### .Commit.Date

{% raw %}`{{ .Commit.Date.Human }}`{% endraw %} provides commit date in Human form.

{% raw %}`{{ .Commit.Date.Unix }}`{% endraw %} provides commit date in Unix epoch form.

##### Example: rebuild whole image every month

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
{{ tpl "<STRING>" <VALUES> }}
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

The `env` function reads an environment variable.

__Syntax__:
{% raw %}
```yaml
{{ env "<ENV_NAME>" }}
{{ env "<ENV_NAME>" "default_value" }}
```
{% endraw %}

> By default, the use of the `env` function is not allowed by giterminism (read more about it [here]({{ "usage/project_configuration/giterminism.html" | true_relative_url }}))

### project files

#### .Files.Get

The function `.Files.Get` gets a certain project file content.

__Syntax__:
{% raw %}
```yaml
{{ .Files.Get "<FILE_PATH>" }}
```
{% endraw %}

> By default, the use of files that have non-committed changes is not allowed by giterminism (read more about it [here]({{ "usage/project_configuration/giterminism.html" | true_relative_url }}))

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
{{ .Files.Glob "<GLOB>" }}
```
{% endraw %}

> By default, the use of files that have non-committed changes is not allowed by giterminism (read more about it [here]({{ "usage/project_configuration/giterminism.html" | true_relative_url }}))

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

## Template directory

_Template files_ can be stored in a reserved directory (`.werf` by default) with the extension `.tmpl` (arbitrary nesting `.werf/**/*.tmpl` is supported).

_Template files_ and the werf configuration file define a common context:

- _Template file_ is a complete template and can be used with the [include](#include) function by the relative path ({% raw %}`{{ include "directory/partial.tmpl" . }}`{% endraw %}).
- The template defined with the `define` function in one _template file_ is available in any other, including the werf configuration file.

### Example: how to use _template files_

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

### Example: how to use templates defined in a _template file_

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
