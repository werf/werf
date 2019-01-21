---
title: Config
sidebar: reference
permalink: reference/config.html
author: Alexey Igrychev <alexey.igrychev@flant.com>, Timofey Kirillov <timofey.kirillov@flant.com>
---

## What is werf config?

Application should be configured to use Werf. This configuration includes:

1. Definition of project meta information such as project name, which will affect build, deploy and other commands.
2. Definition of the images to be built.

Werf uses YAML configuration file `werf.yaml` placed in the root folder of your application. The config is a collection of [YAML documents](http://yaml.org/spec/1.2/spec.html#id2800132) combined with delimiter `---`:

```yaml
YAML_DOC
---
YAML_DOC
---
YAML_DOC
```

Each YAML document has a type. There are currently 3 types of documents:

1. Document to describe project meta information, which will be referred to as *meta configuration doc*.
2. Document to describe image build instructions, which will be referred to as *image configuration doc*.
3. Document to describe artifact build instructions, which will be referred to as *artifact configuraton doc*.

More document types can be added in the future.

### Meta configuration doc

```
project: PROJECT_NAME
OTHER_FIELDS
---
```

Yaml doc with the key `project: PROJECT_NAME` is the meta configuration doc. This is required doc. There should be only one meta configuration doc in the single `werf.yaml` configuration.

#### Project name

`project` defines unique project name of your application. Project name affects build cache image names, Kubernetes Namespace, Helm Release name and other derived names (see [deploy to Kubernetes for detailed description]({{ site.baseurl }}/reference/deploy/deploy_to_kubernetes.html)). This is single required field of meta configuration.

Project name must be maximum 50 chars, only lowercase alphabetic chars, digits and dashes are allowed.

**WARNING**. You should never change project name, once it has been set up, unless you know what you are doing.

Changing project name leads to issues:
1. Invalidation of build cache. New images must be built. Old images must be cleaned up from local host and docker registry manually.
2. Creation of completely new Helm Release. So if you already had deployed your application, then changed project name and deployed it again, there will be created another instance of the same application.

Werf cannot automatically resolve project name change. Described issues must be resolved manually.

### Image configuration doc

Each image configuration doc defines instructions to build one independent docker image. There may be multiple image cofiguration docs defined in the same `werf.yaml` config to build multiple images.

Yaml doc with the key `dimg: IMAGE_NAME` is the image configuration doc. `dimg` defines short name of the docker image to be built. This name must be unique in single `werf.yaml` config.

```
dimg: IMAGE_NAME_1
OTHER_FIELDS
---
dimg: IMAGE_NAME_2
OTHER_FIELDS
---
...
---
dimg: IMAGE_NAME_N
OTHER_FIELDS
```

### Artifact configuration doc

Artifact configuration doc also defines instructions to build one independent artifact docker image. Arifact is a secondary image aimed to isolate a build process and build tools resources (environments, software, data, see [artifacts article for the details]({{ site.baseurl }}/reference/build/artifact.html)). There may be multiple artifact configuration docs for multiple artifacts defined in the same `werf.yaml` config.

Yaml doc with the key `artifact: IMAGE_NAME` is the artifact configuration doc. `artifact` defines short name of the artifact to be referred to in another docs. This name must be unique in single `werf.yaml` config.

### Minimal config example

Currently Werf requires to define meta configuration doc and at least one image configuration doc. Image configuration docs will be fully optional soon.

Example of minimal werf config:

```yaml
project: my-project
---
dimg: ~
from: alpine:latest
```

## Organizing configuration

Part of the configuration can be moved in ***separate template files*** and then included into __werf.yaml__. _Template files_ should live in the ***.werf*** directory with **.tmpl** extension (any nesting is supported).

> **Tip:** templates can be generated or downloaded before running werf. For example, for sharing common logic between projects.

Werf parses all files in one environment, thus described [define](#include) of one _template file_ becomes available in other files, including _werf.yaml_.

<details markdown="1" open>
<summary><b>werf.yaml</b></summary>

{% raw %}
```yaml
{{ $_ := set . "RubyVersion" "2.3.4" }}
{{ $_ := set . "BaseImage" "alpine" }}

project: my-project
---

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
<summary><b>.werf/ansible/components.tmpl</b></summary>

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

</details>

> If there are templates with the same name werf will use template defined in _werf.yaml_ or the latest described in _templates files_.

If need to use the whole _template file_, use template file path relative to _.werf_ directory as a template name in [include](#include) function.

<details markdown="1" open>
<summary><b>werf.yaml</b></summary>

{% raw %}
```yaml
project: my-project
---
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
<summary><b>.werf/artifact/appserver.tmpl</b></summary>

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
<summary><b>.werf/artifact/storefront.tmpl</b></summary>

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

## Processing of config

The following steps could describe the processing of a YAML configuration file:
1. Reading `werf.yaml` and extra templates from `.werf` directory;
1. Executing Go templates;
1. Saving dump into `.werf.render.yaml` (that file will remain after build and will be available until next render);
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

  project: my-project
  ---

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
  project: my-project
  ---

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
  project: my-project
  ---

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
  project: my-project
  ---

  dimg: app
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
