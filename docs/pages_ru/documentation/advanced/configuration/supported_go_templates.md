---
title: Поддержка Go шаблонов
sidebar: documentation
permalink: documentation/advanced/configuration/supported_go_templates.html
---

## Этапы обработки конфигурации

Следующие шаги описывают как происходит обработка конфигурации:
1. Чтение файла `werf.yaml` и файлов шаблонов из папки `.werf`.
2. Выполнение Go-шаблонов.
3. Сохранение отрендеренного файла, получившегося после применения Go-шаблонов, в `.werf.render.yaml` (файл остаётся после выполнения команды и удаляется автоматически с помощью процедуры GC).
4. Разбиение отрендеренного YAML-файла на секции (они отделяются друг от друга тремя дефисами, смотри подробнее [спецификацию](https://yaml.org/spec/1.2/spec.html#id2800132)).
5. Валидация каждой секции конфигурации:
   * Валидация YAML-синтаксиса (смотрите подробнее [здесь](http://yaml.org/refcard.html)).
   * Валидация werf-синтаксиса.
6. Генерация описанных в файле конфигураций образов.

### Шаблоны Go

В конфигурации возможно использование шаблонов Go (Go templates). Поддерживаются следующие функции:

* [Встроенные функции шаблонов Go](https://golang.org/pkg/text/template/#hdr-Functions) и некоторые особенности языка. Например, использование переменных:<a id="go-templates" href="#go-templates" class="anchorjs-link " aria-label="Anchor link for: go templates" data-anchorjs-icon=""></a>

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

* [Sprig-функции](http://masterminds.github.io/sprig/). Например, использование переменных окружения:<a id="sprig-functions" href="#sprig-functions" class="anchorjs-link " aria-label="Anchor link for: sprig functions" data-anchorjs-icon=""></a>

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

* Функция `include` и конструкция `define` для переиспользования частей конфигурации:<a id="include" href="#include" class="anchorjs-link " aria-label="Anchor link for: include" data-anchorjs-icon=""></a>

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

* Функция `tpl` позволяет обрабатывать строки, как Go шаблоны в `werf.yaml`, передавая их в переменных окружения либо в файлах проекта: [раннее описанный пример с файлами проекта]({{ "documentation/advanced/configuration/organizing_configuration.html#использование-функции-tpl" | true_relative_url }}).<a id="tpl" href="#tpl" class="anchorjs-link " aria-label="Anchor link for: tpl" data-anchorjs-icon=""></a>

* Функции `.Files.Get` и `.Files.Glob` для работы с файлами проекта:<a id="files-get" href="#files-get" class="anchorjs-link " aria-label="Anchor link for: .Files.Get" data-anchorjs-icon=""></a>

  <div class="tabs">
    <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'shell_tab')">Shell</a>
    <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'ansible_tab')">Ansible</a>
  </div>

  <div id="ansible_tab" class="tabs__content" markdown="1">
   
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
    > Функция поддерживает [shell pattern matching](https://www.gnu.org/software/findutils/manual/html_node/find_html/Shell-Pattern-Matching.html) + `**`. Результаты вызова функции можно объединить со [sprig функцией `merge`](https://github.com/Masterminds/sprig/blob/master/docs/dicts.md#merge-mustmerge) (к примеру, `{{ $filesDict := merge (.Files.Glob "*/*.txt") (.Files.Glob "app/**/*.txt") }}`)

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

  <div id="shell_tab" class="tabs__content active" markdown="1">

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
  > Функция поддерживает [shell pattern matching](https://www.gnu.org/software/findutils/manual/html_node/find_html/Shell-Pattern-Matching.html) + `**`. Результаты вызова функции можно объединить со [sprig функцией `merge`](https://github.com/Masterminds/sprig/blob/master/docs/dicts.md#merge-mustmerge) (к примеру, `{{ $filesDict := merge (.Files.Glob "*/*.txt") (.Files.Glob "app/**/*.txt") }}`)

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
