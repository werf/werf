---
title: Running assembly instructions
sidebar: reference
permalink: reference/build/assembly_instructions.html
summary: |
  <a href="https://docs.google.com/drawings/d/e/2PACX-1vQcjW39mf0TUxI7yqNzKPq4_9ffzg2IsMxQxu1Uk1-M0V_Wq5HxZCQJ6x-iD-33u2LN25F1nbk_1Yx5/pub?w=2031&amp;h=144" data-featherlight="image">
      <img src="https://docs.google.com/drawings/d/e/2PACX-1vQcjW39mf0TUxI7yqNzKPq4_9ffzg2IsMxQxu1Uk1-M0V_Wq5HxZCQJ6x-iD-33u2LN25F1nbk_1Yx5/pub?w=1016&amp;h=72">
  </a>
  
  <div class="tab">
    <button class="tablinks active" onclick="openTab(event, 'shell')">Shell</button>
    <button class="tablinks" onclick="openTab(event, 'ansible')">Ansible</button>
  </div>
  
  <div id="shell" class="tabcontent active">
    <div class="language-yaml highlighter-rouge"><pre class="highlight"><code><span class="na">shell</span><span class="pi">:</span>
    <span class="na">beforeInstall</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;bash command&gt;</span>
    <span class="na">install</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;bash command&gt;</span>
    <span class="na">beforeSetup</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;bash command&gt;</span>
    <span class="na">setup</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;bash command&gt;</span>
    <span class="na">cacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">beforeInstallCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">installCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">beforeSetupCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">setupCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span></code></pre>
    </div>
  </div>
  
  <div id="ansible" class="tabcontent">
    <div class="language-yaml highlighter-rouge"><pre class="highlight"><code><span class="na">ansible</span><span class="pi">:</span>
    <span class="na">beforeInstall</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;task&gt;</span>
    <span class="na">install</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;task&gt;</span>
    <span class="na">beforeSetup</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;task&gt;</span>
    <span class="na">setup</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;task&gt;</span>
    <span class="na">cacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">beforeInstallCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">installCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">beforeSetupCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">setupCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span></code></pre>
      </div>
  </div>
  
  <br/>
  <b>Running assembly instructions with git</b>
  
  <a href="https://docs.google.com/drawings/d/e/2PACX-1vTiTbGDHqQZWglQaixUBwR_bQLfrNi_TmeZg9RDScJqBSZ1Sh_WXwrVFkn36K0P8zJIJQvK-ZEIBI9a/pub?w=2033&amp;h=433" data-featherlight="image">
    <img src="https://docs.google.com/drawings/d/e/2PACX-1vTiTbGDHqQZWglQaixUBwR_bQLfrNi_TmeZg9RDScJqBSZ1Sh_WXwrVFkn36K0P8zJIJQvK-ZEIBI9a/pub?w=1016&amp;h=216">
    </a>
---

## Проблематика

TODO

Стадии `before_install`, `install`, `before_setup`, `setup` предлагаются для того, чтобы в них прописывать свои инструкции. Это не все стадии, полный список можно увидеть в man-е или [в диаграме]({{ site.baseurl }}/not_used/stages_diagram.html).

![Диаграмма пользовательских стадий сборки]({{ site.baseurl }}/images/build/stages_01.png "Диаграмма пользовательских стадий сборки")

### Пользовательская стадия
Пользовательская стадия — это стадия, инструкции для сборки которой задаются пользователем dapp.

Инструкции задаются через dappfile или chef-рецепты — зависит от используемого сборщика: shell сборщик или [chef сборщик]({{ site.baseurl }}/ruby/chef.html).

### before_install

На этой стадии производятся долгоживущие настройки ОС, устанавливаются пакеты, версия которых меняется не часто. Например, устанавливается нужная локаль и часовой пояс, выполняются apt-get update, устанавливается nginx, php, расширения для php (версии которых не будут меняться очень часто).

По практике коммиты с изменениями команд или версий ПО на этой стадии составляют менее 1%  от всех коммитов в репозитории. При этом эта стадия самая затратная по времени.

Подсказка: before_install — в эту стадию добавлять редко изменяющиеся или тяжелые пакеты и долгоживущие настройки ОС.

Мнемоника: _перед установкой_ приложения нужно настроить ОС, установить системные пакеты

### install

В этот момент лучше всего ставить прикладные зависимости приложения. Например, composer, нужные расширения для php. Можно добавить общие настройки php (отключение ошибок в браузер, включение логов и прочее, что не часто изменяется).

Коммиты с изменениями в этих вещах составляют примерно 5% от общего числа коммитов. Эта стадия может являться менее затратной по времени, чем before_install.

Подсказка: install — стадия для прикладных зависимостей и их общих настроек.

Мнемоника: _установка_ всего что нужно для приложения.

### before_setup

Основные действия по сборке самого приложения производятся на этой стадии. Компилирование исходных текстов, компилирование ассетов, копирование исходных текстов в особую папку или сжатие исходных текстов — всё это тут.

Здесь выполняются действия, которые нужно произвести, чтобы приложение запустилось после изменения исходных текстов.

Подсказка: before_setup — стадия с действиями над исходными текстами.

Мнемоника: _перед настрокой_ приложения нужно установить само приложение.

### setup

Эта стадия выделена для конфигурации приложения, это может быть копирование файлов в /etc, создание программного модуля с версией приложения. В основном это лёгкие, быстрые действия, которые нужно выполнять при сборке для каждого коммита, либо примерно для 2% коммитов, в которых изменяются конфигурационные файлы приложения.

Стадия выполняется быстро, поэтому выполняется после сборки приложения. Возможны два сценария: либо стадия выполняется каждый коммит, либо изменились исходные тексты и будет перевыполнены стадии before_setup и setup.

Подсказка: setup — стадия для конфигурации приложения и для действий на каждый коммит.

Мнемоника: _настройка_ параметров приложения

## Shell

Shell assembly allows executing commands on any user stage. Each described command will be executed with bash in the time of build, so you can use any bash commands to prepare image.

```yaml
shell:
  beforeInstall:
    - apt update
  install:
    - apt install -qy python
  beforeSetup:
    - mkdir -p /app
  setup:
    - echo 'print "Hello world"' > /app/index.py
```

On any user stage you can use any number of commands as you need.

Executing commands with the shell assembly is like using RUN command in a Dockerfile, except dapp will create a new docker layer on each stage.

### Shell assembly syntax

{% raw %}
```yaml
shell:
  beforeInstall:
  - <bash_command>
  install:
  - <bash_command>
  beforeSetup:
  - <bash_command>
  setup:
  - <bash_command>
  cacheVersion: <version>
  beforeInstallCacheVersion: <version>
  installCacheVersion: <version>
  beforeSetupCacheVersion: <version>
  setupCacheVersion: <version>
```
{% endraw %}

## Ansible

`ansible` directive is similar to `shell`. It has 4 keys: `beforeInstall`, `install`,
`beforeSetup`, `setup` for each available stage. Stage description is an array
of ansible tasks:

```
dimg: app
ansible:
  install:
  - debug: msg='Start install'
  - file: path=/etc mode=0777
  - copy:
      src: /bin/sh
      dest: /bin/sh.orig
  - apk:
      name: curl
      update_cache: yes
  - command: ls -la /bin
```

### Ansible config and stage playbook

Each stage description array is converted to a playbook:


```
- hosts: all
  gather_facts: no
  tasks:
  - debug: msg='Start install'  -.
  - file: path=/etc mode=0777    |
  - copy:                        |> copy from ansible:
      src: /bin/sh               |              install:
      dest: /bin/sh.orig        -'              - debug: ...
  ...
```

This playbook is stored into `/.dapp/ansible-playbook-STAGE/playbook.yml` in stage container and thus available
in introspect mode for debugging purposes.

Default settings for ansible is not suited for dapp, so there are config and inventory files in `/.dapp/ansible-playbook-STAGE`:

`/.dapp/ansible-playbook-STAGE/ansible.cfg`

```
[defaults]
inventory = /.dapp/ansible-playbook-STAGE/hosts
transport = local
; do not generate retry files in ro volumes
retry_files_enabled = False
; more verbose stdout like ad-hoc ansible command
stdout_callback = minimal
```

`/.dapp/ansible-playbook-STAGE/hosts`

```
localhost ansible_python_interpreter=/.dappdeps/ansible/...
```

After generation of this files dapp plays the stage playbook like this:
```
ANSIBLE_CONFIG=/.dapp/ansible-playbook-STAGE/ansible.cfg ansible-playbook /.dapp/ansible-playbook-STAGE/playbook.yml
```

Notes:

1. stdout_callback set to _minimal_ because of more verbosity.
2. Ansible has no live stdout. This can be a show stopper for long lasting commands. Quite console is bad for build.
3. `inventory` and `ansible_python_interpreter` — this can be in dappdeps image
4. It would be great to create ansible-solo command for local plays with builtin config and inventory
5. `gather_facts` can be enabled with modules like `setup`, `set_fact`, etc.

### Checksums

Dapp calculates checksum for each stage before build. Stage is considered to be rebuild if checksum
changed. The simplest checksum is a hash over text of stage configuration. More interesting is checksum of
files involved into build process. You can place ansible config files
everywhere in repository tree. But Ansible has rich syntax for modules and dapp should parse
ansible syntax to get all pathes from `src`, `with_files`, etc and implement logic for lookup plugins
to mount that files into stage container. That is very difficult to implement. That's why we come to 2 approaches:

The first iteration of Ansible builder will implement only text checksum. To calculate checksum for files use go template function .Files.Get and `content` attribute of modules.

{% raw %}
```
> dappfile.yml

ansible:
  install:
  - copy:
      content: {{ .Files.Get '/conf/etc/nginx.conf'}}
      dest: /etc/nginx
```
{% endraw %}

```
> resulting playbook.yml

- hosts: all
  gather_facts: no
  tasks:   
    install:
    - copy:
        content: |
          http {
            sendfile on;
            tcp_nopush on;
            tcp_nodelay on;
            ...
```

.Files.Get input is path to file in repository. Function returns string with file content.

### Modules

Initial ansible builder will support only some modules:

* Command
* Shell
* Copy
* Debug
* packaging category

Other ansible modules are available, but they may be not stable.

### Jinja templating

{% raw %}
Go template and jinja has equal delimeters: `{{` and `}}`.
{% endraw %}

First iteration will support only go style escaping:

{% raw %}
```
> dappfile.yml

git:
- add: '/'
  to: '/app'
ansible:
  install:
  - copy:
      src: {{"{{"}} item {{"}}"}} OR src: {{`{{item}}`}}
      dest: /etc/nginx
      with_files:
      - /app/conf/etc/nginx.conf
      - /app/conf/etc/server.conf
```
{% endraw %}

### Templates

No templates for first iteration. Templates can be supported when .Files.Path will be implemented.

### Ansible problems

1. No live stdout. In general we need to implement our stdout callback and connection plugin.
  stdout callbacks has no pre_* methods for display information about executed task. There are only post_* methods
  for display information about finished task.
2. `-c local` may be an overload because of zipping modules. There must be a way to directly start modules.
  Ansible-solo command with direct modules calls would be a great improvement for building images.

### Ansible assembly syntax

{% raw %}
```yaml
ansible:
  beforeInstall:
  - <task>
  install:
  - <task>
  beforeSetup:
  - <task>
  setup:
  - <task>
  cacheVersion: <version>
  beforeInstallCacheVersion: <version>
  installCacheVersion: <version>
  beforeSetupCacheVersion: <version>
  setupCacheVersion: <version>
```   
{% endraw %}

## Forced build (CacheVersion)

If you need to forcibly rebuild some or all of user stages (in case they are not actual) you can use `<stage>cacheVersion` directives. Using any of these directives, give a possibility to force the rebuild of the appropriate stage and all of the subsequent stages when values in used directives have changed.

E.g., you use the following dappfile:

```
dimg: ~
from: alpine:latest

shell:
  installCacheVersion: 1
  beforeInstall:
  - echo "Commands on the Before Install stage"
  install:
  - echo "Commands on the Install stage"
  beforeSetup:
  - echo "Commands on the Before Setup stage"
  setup:
  - echo "Commands on the Setup stage"
```

When you've built this image with `dapp dimg build` command, run it again - dapp will use its build cache and will not rebuild any stages described in the dappfile. But if you'll change `installCacheVersion` value and will run again `dapp dimg build` - the install stage will be forcibly rebuilt in spite of stage commands have no changes.

You can use all of the `<stage>cacheVersion` directives when describing an image in dappfile. E.g.:

{% raw %}
```
dimg: ~
from: alpine:latest

shell:
  cacheVersion: 1
  beforeInstallCacheVersion: 1
  installCacheVersion: 1
  beforeSetupCacheVersion: 1
  setupCacheVersion: 1
  beforeInstall:
  - echo "Commands on the Before Install stage"
  install:
  - echo "Commands on the Install stage"
  beforeSetup:
  - echo "Commands on the Before Setup stage"
  setup:
  - echo "Commands on the Setup stage"
```
{% endraw %}

Pay attention that `cacheVersion` and `beforeInstallCacheVersion` directives have the same effect - when their values have changed, then `beforeInstall` stage and subsequent stages will be rebuilt.

When may you need to forcibly rebuild stage? When may the stage be not actual?

The simple case is, e.g., when you make `apt update` on the `beforeInstall` stage, and you wish that once a month `beforeInstall` stage be rebuilt and APT cache be updated. To achieve this, you can use the following dappfile, using GO templates:

{% raw %}
```
dimg: ~
from: ubuntu:latest
shell:
  beforeInstallCacheVersion: {{ now | date "2006-01" }}
  beforeInstall:
  - apt update
```
{% endraw %}

> Why is it "2006-01" argument in the date function? You can read about time format [here](https://golang.org/pkg/time/#pkg-constants).

## Использование совместно с добавлением исходников

Проблема возникает, когда подготовка образа включает в себя, например, установку внешних зависимостей gem'ов в образ на основе Gemfile и Gemfile.lock из git-репозитория. В таком случае необходимо пересобирать стадию, на которой происходит установка этих зависимостей, если поменялся Gemfile или Gemfile.lock.

* Существуют стадии, в формировании [cигнатур](#сигнатура-стадии) которых используется сигнатура последующей стадии, вдобавок к зависимостям самой стадии. Такие стадии всегда будут пересобираться вместе с зависимой стадией.
  * git artifact pre install patch зависит от install.
  * git artifact post install patch зависит от before setup.
  * git artifact pre setup patch зависит от setup.
  * git artifact artifact patch зависит от build artifact.
