---
title: Organizing configuration
sidebar: documentation
permalink: documentation/advanced/configuration/organizing_configuration.html
---

## With templates dir

Part of the configuration can be moved in ***separate template files*** and then included into __werf.yaml__. _Template files_ should live in the ***.werf*** directory with **.tmpl** extension (any nesting is supported).

> **Tip:** templates can be generated or downloaded before running werf. For example, for sharing common logic between projects

werf parses all files in one environment, thus described _define_ of one _template file_ becomes available in other files, including _werf.yaml_.

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

> If there are templates with the same name werf will use template defined in _werf.yaml_ or the latest described in _templates files_

If need to use the whole _template file_, use template file path relative to _.werf_ directory as a template name in _include_ function.

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

## With tpl function

The `tpl` function allows the user to evaluate strings as Go templates inside a template. Thus, werf partials can be located anywhere in the project and be included in `werf.yaml`.

It is worth noting that these files can use anything defined in `werf.yaml` and templates in `.werf` ([templates dir](#with-templates-dir)).

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
