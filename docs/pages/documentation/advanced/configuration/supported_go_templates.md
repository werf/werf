---
title: Supported Go templates
sidebar: documentation
permalink: documentation/advanced/configuration/supported_go_templates.html
---

## Processing of config

The following steps could describe the processing of a YAML configuration file:
1. Reading `werf.yaml` and extra templates from `.werf` directory.
2. Executing Go templates.
3. Saving dump into `.werf.render.yaml` (this file remains after the command execution and will be removed automatically with GC procedure).
4. Splitting rendered YAML file into separate config sections (part of YAML stream separated by three hyphens, https://yaml.org/spec/1.2/spec.html#id2800132).
5. Validating each config section:
   * Validating YAML syntax (you could read YAML reference [here](http://yaml.org/refcard.html)).
   * Validating werf syntax.

## Go templates

Go templates are available within YAML configuration. The following functions are supported:

* [Built-in Go template functions](https://golang.org/pkg/text/template/#hdr-Functions) and other language features. E.g. using common variable:<a id="go-templates" href="#go-templates" class="anchorjs-link " aria-label="Anchor link for: go templates" data-anchorjs-icon=""></a>

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

* [Sprig functions](http://masterminds.github.io/sprig/). E.g. using environment variable:<a id="sprig-functions" href="#sprig-functions" class="anchorjs-link " aria-label="Anchor link for: sprig functions" data-anchorjs-icon=""></a>

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

* `include` function with `define` for reusing configs:<a id="include" href="#include" class="anchorjs-link " aria-label="Anchor link for: include" data-anchorjs-icon=""></a>

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

* `tpl` function to evaluate strings (either content of environment variable or project file) as Go templates inside a template: [example with project files]({{ "documentation/advanced/configuration/organizing_configuration.html#with-tpl-function" | true_relative_url }}).<a id="tpl" href="#tpl" class="anchorjs-link " aria-label="Anchor link for: tpl" data-anchorjs-icon=""></a>

* `.Files.Get` and `.Files.Glob` functions to work with project files:<a id="files-get" href="#files-get" class="anchorjs-link " aria-label="Anchor link for: .Files.Get and .Files.Glob" data-anchorjs-icon=""></a>

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
    > The function supports [shell pattern matching](https://www.gnu.org/software/findutils/manual/html_node/find_html/Shell-Pattern-Matching.html) + `**`. Results can be merged with [`merge` sprig function](https://github.com/Masterminds/sprig/blob/master/docs/dicts.md#merge-mustmerge) (e.g `{{ $filesDict := merge (.Files.Glob "*/*.txt") (.Files.Glob "app/**/*.txt") }}`)

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
  > The function supports [shell pattern matching](https://www.gnu.org/software/findutils/manual/html_node/find_html/Shell-Pattern-Matching.html) + `**`. Results can be merged with [`merge` sprig function](https://github.com/Masterminds/sprig/blob/master/docs/dicts.md#merge-mustmerge) (e.g `{{ $filesDict := merge (.Files.Glob "*/*.txt") (.Files.Glob "app/**/*.txt") }}`)

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
