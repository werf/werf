---
title: Ansible builder
sidebar: not_used
permalink: not_used/ansible_builder.html
---

You can use ansible to build your containers.

## dappfile placing

dappfile may be put at different locations:

* `REPO_ROOT/dappfile.yml`
* `REPO_ROOT/dappfile.yaml`
* `REPO_ROOT/.dappfiles/*.yml`
* `REPO_ROOT/.dappfiles/*.yaml`

Dapp will read files from directory `.dappfiles` in the alphabetical order.
`dappfile.yml` will preceed any files from `.dappfiles`

## dappfile yaml syntax 

Processing of Yaml configuration consists of several steps:

1. Find and read all available files
2. Render each file as go-template with sprig functions (https://golang.org/pkg/text/template/, http://masterminds.github.io/sprig/)
3. Split each file by `---` marker into array of yaml documents
4. Parse each document in array as yaml description of dimg or artifact

Parsing rules:
1. Each yaml document must have `dimg` or `artifact` directive with name.
2. Each document must have `from` directive.
3. Whole configuration must have at least one dimg.
4. Whole configuration must have builder instructions of one type: `shell` or `ansible`.

Example of minimal configuration:

```
# dappfile.yml
dimg: app
from: alpine:latest
```

Support for ansible builder divide to this parts:

1. dappdeps-ansible docker image with python and ansible
2. Dapp::Dimg::Builder::Ansible ruby module. This code converts array of ansible tasks to ansible-playbook
   and verify checksums
3. support for `ansible` directive in yaml configuration

### differences from Chef-style dappfile

1. No equivalent for `dimg_group` and nested `dimg`s and `artifact`s.
2. No context inheritance because of 1. Use go-template functionality
   to define common parts.
3. Use `import` in `dimg` for copy artifact results instead of `export`
4. Each `artifact` must have a name

### dappfile config

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

### ansible config and stage playbook

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

### checksums

Dapp calculates checksum for each stage before build. Stage is considered to be rebuild if checksum
changed. The simplest checksum is a hash over text of stage configuration. More interesting is checksum of
files involved into build process. You can place ansible config files
everywhere in repository tree. But Ansible has rich syntax for modules and dapp should parse
ansible syntax to get all pathes from `src`, `with_files`, etc and implement logic for lookup plugins
to mount that files into stage container. That is very difficult to implement. That's why we come to 2 approaches:

The first iteration of Ansible builder will implement only text checksum. To calculate checksum for files use go template
function .Files.Get and `content` attribute of modules.

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

Next iteration will implement .Files.Path function. Builder can collect all calls to this function
and generate files structure to mount as volume.
 
{% raw %}
```
> dappfile.yml

ansible:
  install:
  - copy:
      src: {{ .Files.Path '/conf/etc/nginx.conf' }}
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
        src: /.dapp/ansible-files-install/conf--etc--nginx.conf
        dest: /etc/nginx

```

### modules

Initial ansible builder will support only some modules:

* Command
* Shell
* Copy
* Debug
* packaging category

Other ansible modules are available, but they may be not stable.

### jinja templating

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

### templates

No templates for first iteration. Templates can be supported when .Files.Path will be implemented.


### ansible problems

1. No live stdout. In general we need to implement our stdout callback and connection plugin.
  stdout callbacks has no pre_* methods for display information about executed task. There are only post_* methods
  for display information about finished task.
2. `-c local` may be an overload because of zipping modules. There must be a way to directly start modules.
  Ansible-solo command with direct modules calls would be a great improvement for building images.
