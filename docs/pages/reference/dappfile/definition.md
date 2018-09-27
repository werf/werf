---
title: What is dappfile?
sidebar: reference
permalink: reference/dappfile/definition.html
---

Dapp uses the configuration file called dappfile to build docker images. At present, dapp only support YAML syntax and uses one of the following files in the root folder of your project — `dappfile.yaml` or `dappfile.yml`. 

The dappfile is a collection of [YAML documents](http://yaml.org/spec/1.2/spec.html#id2800132) combined with splitter `---`. Each YAML document contains instructions to build independent docker image. 

Thus, you can describe multiple images in one dappfile:

```yaml
YAML_DOC
---
YAML_DOC
---
YAML_DOC
```

## Processing of dappfile

Processing of YAML configuration file could be described by the following steps:
1. Reading dappfile;
1. Executing Go templates;
1. Saving dump into `.dappfile.render.yaml` (that file will remain after build and will be available until next render);
1. Splitting rendered YAML file into separate YAML documents;
1. Validating each YAML document:
  * Validating YAML syntax (you could read YAML reference [here](http://yaml.org/refcard.html)).
  * Validating our syntax.

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
