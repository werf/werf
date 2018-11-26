---
title: Dappfile
sidebar: reference
permalink: reference/build/dappfile.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

## What is dappfile?

Dapp uses the configuration file called ***dappfile*** to build docker images. At present, dapp only support YAML syntax and uses one of the following files in the root folder of your project — `dappfile.yaml` or `dappfile.yml`. 

The dappfile is a collection of [YAML documents](http://yaml.org/spec/1.2/spec.html#id2800132) combined with delimiter `---`. Each YAML document contains instructions to build independent docker image. 

Thus, you can describe multiple images in one dappfile:

```yaml
YAML_DOC
---
YAML_DOC
---
YAML_DOC
```

Dapp represents YAML document as an internal object. Currently, dapp supports two object types, [_dimg_]({{ site.baseurl }}/reference/build/naming.html) and [_artifact_]({{ site.baseurl }}/reference/build/artifact.html). 

***Dimg*** is the named set of rules, an internal representation of user image. An ***artifact*** is a special dimg that is used by another _dimgs_ and _artifacts_ to isolate the build process and build tools resources (environments, software, data).  

## Organizing dappfile configuration

Part of the configuration can be moved in ***separate template files*** and then includes in __dappfile.yaml__. _Template files_ should live in the ***.dappfiles*** directory with **.tmpl** extension (any nesting is supported).

> **Tip:** templates can be generated or downloaded before running dapp. For example, for sharing common logic between projects.

Dapp parses all files in one environment, thus described [define](#include) of one _template file_ becomes available in other files, including _dappfile.yaml_.

<details markdown="1" open>
<summary><b>dappfile.yaml</b></summary>

{% raw %}
```yaml
{{ $_ := set . "RubyVersion" "2.3.4" }}
{{ $_ := set . "BaseImage" "alpine" }}

dimg: rails
from: {{ .BaseImage }}
ansible:
  beforeInstall:
  {{- include "(component) mysql client" . }}
  {{- include "(component) ruby" . }}
  install:
  {{- include "(component) Gemfile dependencies" . }}
```
{% endraw %}

</details>

<details markdown="1">
<summary><b>.dappfiles/ansible/components.tmpl</b></summary>

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
      ssh-keyscan fox.flant.com >> /root/.ssh/known_hosts
      ssh-keyscan gitlab.flant.ru >> /root/.ssh/known_hosts
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

</details>

> If there are templates with the same name dapp will be will use template defined in _dappfile.yaml_ or the latest described in _templates files_.

If need to use the whole _template file_, use template file path relative to _.dappfiles_ directory as a template name in [include](#include) function.

<details markdown="1" open>
<summary><b>dappfile.yaml</b></summary>

{% raw %}
```yaml
dimg: app
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

</details>

<details markdown="1">
<summary><b>.dappfiles/artifact/appserver.tmpl</b></summary>

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

</details>

<details markdown="1">
<summary><b>.dappfiles/artifact/storefront.tmpl</b></summary>

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

</details>

## Processing of dappfile

The following steps could describe the processing of a YAML configuration file:
1. Reading `dappfile.yaml` and extra templates from `.dappfiles` directory;
1. Executing Go templates;
1. Saving dump into `.dappfile.render.yaml` (that file will remain after build and will be available until next render);
1. Splitting rendered YAML file into separate YAML documents;
1. Validating each YAML document:
  * Validating YAML syntax (you could read YAML reference [here](http://yaml.org/refcard.html)).
  * Validating our syntax.
1. Generating a set of dimgs.

### Go templates

Go templates are available within YAML configuration. The following functions are supported:

* [Built-in Go template functions](https://golang.org/pkg/text/template/#hdr-Functions) and other language features. E.g. using common variable:
  
  {% raw %}
  ```yaml
  {{ $base_image := "golang:1.11-alpine" }}
  
  dimg: gogomonia
  from: {{ $base_image }}
  ---
  dimg: saml-authenticator
  from: {{ $base_image }}
  ```
  {% endraw %}

* [Sprig functions](http://masterminds.github.io/sprig/). E.g. using environment variable:
  
  {% raw %}
  ```yaml
  {{ $_ := env "SPECIFIC_ENV_HERE" | set . "GitBranch" }}

  dimg: ~
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

* `include` function with `define` for reusing configs:<a id="include" href="#include" class="anchorjs-link " aria-label="Anchor link for: include" data-anchorjs-icon=""></a>
  
  {% raw %}
  ```yaml
  dimg: app1
  from: alpine
  ansible:
    beforeInstall:
    {{- include "(component) ruby" . }}
  ---
  dimg: app2
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

* `.Files.Get` function for getting project file content:<a id="files-get" href="#files-get" class="anchorjs-link " aria-label="Anchor link for: .Files.Get" data-anchorjs-icon=""></a>

  {% raw %}
  ```yaml
  dimg: app
  from: alpine
  ansible:
    setup:
    - name: "Setup /etc/nginx/nginx.conf"
      copy:
        content: |
  {{ .Files.Get ".dappfiles/nginx.conf" | indent 8 }}
        dest: /etc/nginx/nginx.conf
  ```
  {% endraw %}
