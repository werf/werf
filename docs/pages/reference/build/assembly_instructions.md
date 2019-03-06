---
title: Running assembly instructions
sidebar: reference
permalink: reference/build/assembly_instructions.html
summary: |
  <a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vQcjW39mf0TUxI7yqNzKPq4_9ffzg2IsMxQxu1Uk1-M0V_Wq5HxZCQJ6x-iD-33u2LN25F1nbk_1Yx5/pub?w=2031&amp;h=144" data-featherlight="image">
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

  <a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRv56S-dpoTSzLC_24ifLqJHQoHdmJ30l1HuAS4dgqBgUzZdNQyA1balT-FwK16pBbbXqlLE3JznYDk/pub?w=1956&amp;h=648" data-featherlight="image">
    <img src="https://docs.google.com/drawings/d/e/2PACX-1vRv56S-dpoTSzLC_24ifLqJHQoHdmJ30l1HuAS4dgqBgUzZdNQyA1balT-FwK16pBbbXqlLE3JznYDk/pub?w=622&amp;h=206">
  </a>
---

## What is user stages?

***User stage*** is a [_stage_]({{ site.baseurl }}/reference/build/stages.html) with _assembly instructions_ from config.
Currently, there are two kinds of assembly instructions: _shell_ and _ansible_. Werf
defines 4 _user stages_ and executes them in this order: _beforeInstall_, _install_,
_beforeSetup_ and _setup_. Assembly instructions from one stage are executed to
create one docker layer.

## Motivation behind stages

### Opinionated build structure

_User stages pattern_ is based on analysis of real applications building
instructions. It turns out that group assembly instructions into 4 _user stages_
are enough for most applications. Instructions grouping decrease layers sizes
and speed up image building.

### Framework for a build process

_User stages pattern_ defines a structure for building process and thus set
boundaries for a developer. This is a high speed up over unstructured
instructions in Dockerfile because developer knows what kind of instructions
should be on each _stage_.

### Run assembly instruction on git changes

_User stage_ execution can depend on changes of files in a git repository. Werf
supports local and remote git repositories. User stage can be
dependent on changes in several repositories. Different changes in one
repository can cause a rebuild of different _user stages_.

### More build tools: shell, ansible, ...

_Shell_ is a familiar and well-known build tool. Ansible is a newer tool and it
needs some time for learning.

If you need prototype as soon as possible then _the shell_ is enough — it works
like a `RUN` instruction in Dockerfile. Ansible configuration is declarative and
mostly idempotent which is good for long-term maintenance.

Stage execution is isolated in code, so implementing support for another tool
is not so difficult.

## Usage of user stages

Werf provides 4 _user stages_ where assembly instructions can be defined. Assembly
instructions are not limited by werf. You can define whatever you define for `RUN`
instruction in Dockerfile. However, assembly instructions grouping arises from
experience with real-life applications. So the vast majority of application builds
need these actions:

- install system packages
- install system dependencies
- install application dependencies
- setup system applications
- setup application

What is the best strategy to execute them? First thought is to execute them one
by one to cache interim results. The other thought is not to mix instructions
for these actions because of different file dependencies. _User stages pattern_
suggests this strategy:

- use _beforeInstall_ user stage for installing system packages
- use _install_ user stage to install system dependencies and application dependencies
- use _beforeSetup_ user stage to setup system settings and install an application
- use _setup_ user stage to setup application settings

### beforeInstall

A stage that executes instructions before install an application. This stage
is for system applications that rarely changes but time consuming to install.
Also, long-lived system settings can be done here like setting locale, setting
time zone, adding groups and users, etc. Installation of enormous language
distributions and build tools like PHP and composer, java and gradle, etc. are good candidates to execute at this stage.

In practice, these components are rarely changes, and _beforeInstall_ stage caches
them for an extended period.

### install

A stage to install an application. This stage is for installation of
application dependencies and setup some standard settings.

Instructions on this stage have access to application source codes, so
application dependencies can be installed with build tools (like composer,
gradle, npm, etc.) that require some manifest file (i.e., pom.xml,
Gruntfile). Best practice is to make this stage dependent on changes in that
manifest file.

### beforeSetup

This stage is for prepare application before setup some settings. Every kind
of compilation can be done here: creating jars, creating executable files and
dynamic libraries, creating web assets, uglification and encryption. This stage
often makes to be dependent on changes in source codes.

### setup

This stage is to setup application settings. The usual actions here are copying
some profiles into `/etc`, copying configuration files into well-known
locations, creating a file with the application version. These actions should not be
time-consuming as they execute on every commit.

### custom strategy

Again, there are no limitations for assembly instructions. The previous
definitions of _user stages_ are just suggestions arise from experience with real
applications. You can even use only one _user stage_ or define your strategy
of grouping assembly instructions and get benefits from caching and git
dependencies.

## Syntax

There are two ***builder directives*** for assembly instructions on top level: `shell` and `ansible`. These builder directives are mutually exclusive. You can build an image with ***shell assembly instructions*** or with ***ansible assembly instructions***.

_Builder directive_ has four directives which define assembly instructions for each _user stage_:
- `beforeInstall`
- `install`
- `beforeSetup`
- `setup`

Builder directives also contain ***cacheVersion  directives*** that are user-defined parts of _user stages signatures_. More details in a [CacheVersion](#dependency-on-cacheversion-values) section.

## Shell

Syntax for _user stages_ with _shell assembly instructions_:

```yaml
shell:
  beforeInstall:
  - <bash_command 1>
  - <bash_command 2>
  ...
  - <bash_command N>
  install:
  - bash command
  ...
  beforeSetup:
  - bash command
  ...
  setup:
  - bash command
  ...
  cacheVersion: <version>
  beforeInstallCacheVersion: <version>
  installCacheVersion: <version>
  beforeSetupCacheVersion: <version>
  setupCacheVersion: <version>
```

_Shell assembly instructions_ are arrays of bash commands for _user stages_. Commands for one stage are executed as one `RUN` instruction in Dockerfile, and thus werf creates one layer for one _user stage_.

Werf provides distribution agnostic bash binary, so you need no bash binary in the [base image]({{ site.baseurl }}/reference/build/base_image.html). Commands for one stage are joined with `&&` and then encoded as base64. _User stage assembly container_ runs decoding and then executes joined commands. For example, _beforeInstall_ stage with `apt-get update` and `apt-get install` commands:

```yaml
beforeInstall:
- apt-get update
- apt-get install -y build-essential g++ libcurl4
```

These commands transform into this command for _user stage assembly container_:
```shell
bash -ec 'eval $(echo YXB0LWdldCB1cGRhdGUgJiYgYXB0LWdldCBpbnN0YWxsIC15IGJ1aWxkLWVzc2VudGlhbCBnKysgbGliY3VybDQK | base64 --decode)'
```

`bash` and `base64` binaries are stored in _werfdeps volume_. Details of _werfdeps volumes_ can be found in this [blog post [RU]](https://habr.com/company/flant/blog/352432/).

## Ansible

Syntax for _user stages_ with _ansible assembly instructions_:

```yaml
ansible:
  beforeInstall:
  - <ansible task 1>
  - <ansible task 2>
  ...
  - <ansible task N>
  install:
  - ansible task
  ...
  beforeSetup:
  - ansible task
  ...
  setup:
  - ansible task
  ...
  cacheVersion: <version>
  beforeInstallCacheVersion: <version>
  installCacheVersion: <version>
  beforeSetupCacheVersion: <version>
  setupCacheVersion: <version>
```

### Ansible config and stage playbook

_Ansible assembly instructions_ for _user stage_ is a set of ansible tasks. To run
this tasks with `ansible-playbook` command werf mounts this directory structure
into the _user stage assembly container_:

```bash
/.werf/ansible-workdir
├── ansible.cfg
├── hosts
└── playbook.yml
```

`ansible.cfg` contains settings for ansible:
- use local transport
- werf stdout_callback for better logging
- turn on force_color
- use sudo for privilege escalation (no need to use `become` in tasks)

`hosts` is an inventory file and contains the only localhost. Also, there are some
ansible_* settings, i.e., the path to python in werfdeps.

`playbook.yml` is a playbook with all tasks from one _user stage_. For example,
`werf.yaml` with _install_ stage like this:

```yaml
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
  ...
```

werf produces this `playbook.yml` for _install_ stage:
```yaml
- hosts: all
  gather_facts: 'no'
  tasks:
  - debug: msg='Start install'  \
  - file: path=/etc mode=0777   |
  - copy:                        > these lines are copied from werf.yaml
      src: /bin/sh              |
      dest: /bin/sh.orig        |
  - apk:                        |
      name: curl                |
      update_cache: yes         /
  ...
```

Werf plays the _user stage_ playbook in the _user stage assembly container_ with `playbook-ansible`
command:

```bash
$ export ANSIBLE_CONFIG="/.werf/ansible-workdir/ansible.cfg"
$ ansible-playbook /.werf/ansible-workdir/playbook.yml
```

`ansible` and `python` binaries and libraries are stored in _werfdeps/ansible volume_. Details of _werfdeps volumes_ can be found in this [blog post [RU]](https://habr.com/company/flant/blog/352432/).

### Supported modules

One of the ideas behind werf is idempotent builds. If nothing changed — werf
should create the same image. This task accomplished by _signature_ calculation
for _stages_. Ansible has non-idempotent modules — they are giving different
results if executed twice and werf cannot correctly calculate _signature_ to
rebuild _stages_. For now, there is a list of supported modules:

- [Command modules](https://docs.ansible.com/ansible/2.5/modules/list_of_commands_modules.html): command, shell, raw, script.

- [Crypto modules](https://docs.ansible.com/ansible/2.5/modules/list_of_crypto_modules.html): openssl_certificate, and other.

- [Files modules](https://docs.ansible.com/ansible/2.5/modules/list_of_files_modules.html): acl, archive, copy, stat, tempfile, and other.

- [Net Tools Modules](https://docs.ansible.com/ansible/2.5/modules/list_of_net_tools_modules.html): get_url, slurp, uri.

- [Packaging/Language modules](https://docs.ansible.com/ansible/2.5/modules/list_of_packaging_modules.html#language): composer, gem, npm, pip, and other.

- [Packaging/OS modules](https://docs.ansible.com/ansible/2.5/modules/list_of_packaging_modules.html#os): apt, apk, yum, and other.

- [System modules](https://docs.ansible.com/ansible/2.5/modules/list_of_system_modules.html): user, group, getent, locale_gen, timezone, cron, and other.

- [Utilities modules](https://docs.ansible.com/ansible/2.5/modules/list_of_utilities_modules.html): assert, debug, set_fact, wait_for.

_Werf config_ with the module not from this list gives an error and stops a build. Feel free to report an issue if some module should be enabled.

### Copy files

The preferred way of copying files into an image is [_git paths_]({{ site.baseurl }}/reference/build/git_directive.html). Werf cannot calculate changes of files referred in `copy` module. The only way to
copy some external file into an image, for now, is to use the go-templating method
`.Files.Get`. This method returns file content as a string. So content of the file becomes a part of _user stage signature_, and file changes lead to _user stage_
rebuild.

For example, copy `nginx.conf` into an image:

{% raw %}
```yaml
ansible:
  install:
  - copy:
      content: |
{{ .Files.Get '/conf/etc/nginx.conf' | indent 6}}
      dest: /etc/nginx
```
{% endraw %}

Werf renders that snippet as go template and then transforms it into this `playbook.yml`:

```yaml
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

### Jinja templating

Ansible supports [Jinja templating](https://docs.ansible.com/ansible/2.5/user_guide/playbooks_templating.html) of playbooks. However, Go templates and Jinja
templates has the same delimiters: {% raw %}`{{` and `}}`{% endraw %}. Jinja templates should be escaped
 to work. There are two possible variants: escape only {% raw %}`{{`{% endraw %} or escape
the whole Jinja expression.

For example, you have this ansible task:

{% raw %}
```yaml
- copy:
    src: {{item}}
    dest: /etc/nginx
    with_files:
    - /app/conf/etc/nginx.conf
    - /app/conf/etc/server.conf
```
{% endraw %}

{% raw %}
So, Jinja expression `{{item}}` should be escaped:
{% endraw %}

{% raw %}
```yaml
# escape {{ only
src: {{"{{"}} item }}
```
or
```yaml
# escape the whole expression
src: {{`{{item}}`}}
```
{% endraw %}

### Ansible problems

- Live stdout implemented for raw and command modules. Other modules display stdout and stderr content after execution.
- Excess logging into stderr may hang ansible task execution ([issue #784](https://github.com/flant/werf/issues/784)).
- `apt` module hangs build process on particular debian and ubuntu versions. This affects derived images as well ([issue #645](https://github.com/flant/werf/issues/645)).

## User stages dependencies

One of the werf features is an ability to define dependencies for _stage_ rebuild.
As described in [_stages_ reference]({{ site.baseurl }}/reference/build/stages.html), _stages_ are built one by one, and each _stage_ has
a calculated _stage signature_. _Signatures_ have various dependencies. When
dependencies are changed, the _stage signature_ is changed, and werf rebuild this _stage_ and
all following _stages_.

These dependencies can be used for defining rebuild for the
_user stages_. _User stages signatures_ and so rebuilding of the _user stages_
depends on:
- changes in assembly instructions
- changes of _cacheVersion directives_
- git repository changes
- changes in files that imports from an [artifacts]({{ site.baseurl }}/reference/build/artifact.html)

First three dependencies are described further.

## Dependency on assembly instructions changes

_User stage signature_ depends on rendered assembly instructions text. Changes in
assembly instructions for _user stage_ lead to the rebuilding of this _stage_. E.g., you
use the following _shell assembly instructions_:

```yaml
shell:
  beforeInstall:
  - echo "Commands on the Before Install stage"
  install:
  - echo "Commands on the Install stage"
  beforeSetup:
  - echo "Commands on the Before Setup stage"
  setup:
  - echo "Commands on the Setup stage"
```

First, build of this image execute all four _user stages_. There is no _git path_ in
this _config_, so next builds never execute assembly instructions because _user
stages signatures_ not changed and build cache remains valid.

Changing assembly instructions for _install_ user stage:

```yaml
shell:
  beforeInstall:
  - echo "Commands on the Before Install stage"
  install:
  - echo "Commands on the Install stage"
  - echo "Installing ..."
  beforeSetup:
  - echo "Commands on the Before Setup stage"
  setup:
  - echo "Commands on the Setup stage"
```

Now `werf build` executes _install assembly instructions_ and instructions from
following _stages_.

Go-templating and using environment variables can changes assembly instructions
and lead to unforeseen rebuilds. For example:

{% raw %}
```yaml
shell:
  beforeInstall:
  - echo "Commands on the Before Install stage for {{ env "CI_COMMIT_SHA” }}"
  install:
  - echo "Commands on the Install stage"
  ...
```
{% endraw %}

First build renders _beforeInstall command_ into:
```bash
echo "Commands on the Before Install stage for 0a8463e2ed7e7f1aa015f55a8e8730752206311b"
```

Build for the next commit renders _beforeInstall command_ into:

```bash
echo "Commands on the Before Install stage for 36e907f8b6a639bd99b4ea812dae7a290e84df27"
```

Using `CI_COMMIT_SHA` assembly instructions text changes every commit.
So this configuration rebuilds _beforeInstall_ user stage on every commit.

## Dependency on git repo changes

<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRv56S-dpoTSzLC_24ifLqJHQoHdmJ30l1HuAS4dgqBgUzZdNQyA1balT-FwK16pBbbXqlLE3JznYDk/pub?w=1956&amp;h=648" data-featherlight="image">
    <img src="https://docs.google.com/drawings/d/e/2PACX-1vRv56S-dpoTSzLC_24ifLqJHQoHdmJ30l1HuAS4dgqBgUzZdNQyA1balT-FwK16pBbbXqlLE3JznYDk/pub?w=622&amp;h=206">
  </a>

As stated in a _git path_ reference, there are _gitArchive_ and _gitLatestPatch_ stages. _gitArchive_ is executed after _beforeInstall_ user stage, and _gitLatestPatch_ is executed after _setup_ user stage if a local git repository has changes. So, to execute assembly instructions with the latest version of source codes, you may rebuild _gitArchive_ with [special commit]({{site.baseurl}}/reference/build/git_directive.html#rebuild-of-git_archive-stage) or rebuild _beforeInstall_ (change _cacheVersion_ or instructions for _beforeInstall_ stage).

_install_, _beforeSetup_ and _setup_ user stages are also dependant on git repository changes. A git patch is applied at the beginning of _user stage_ to execute assembly instructions with the latest version of source codes.

_User stage_ dependency on git repository changes is defined with `git.stageDependencies` parameter. Syntax is:

```yaml
git:
- ...
  stageDependencies:
    install:
    - <mask 1>
    ...
    - <mask N>
    beforeSetup:
    - <mask>
    ...
    setup:
    - <mask>
```

`git.stageDependencies` parameter has 3 keys: `install`, `beforeSetup` and `setup`. Each key defines an array of masks for one user stage. User stage is rebuilt if a git repository has changes in files that match with one of the masks defined for _user stage_.

For each _user stage_ werf creates a list of matched files and calculates a checksum over each file attributes and content. This checksum is a part of _stage signature_. So signature is changed with every change in a repository: getting new attributes for the file, changing file's content, adding a new matched file, deleting a matched file, etc.

`git.stageDependencies` masks work together with `git.includePaths` and `git.excludePaths` masks. werf considers only files matched with `includePaths` filter and `stageDependencies` masks. Likewise, werf considers only files not matched with `excludePaths` filter and matched with `stageDependencies` masks.

`stageDependencies` masks works like `includePaths` and `excludePaths` filters. Masks are matched with files paths with [fnmatch method](https://ruby-doc.org/core-2.2.0/File.html#method-c-fnmatch). Masks may contain the following patterns:

- `*` — matches any file. This pattern includes `.` and excludes `/`
- `**` — matches directories recursively or files expansively
- `?` — matches any one character. Equivalent to /.{1}/ in regexp
- `[set]` — matches any one character in the set. Behaves exactly like character sets in regexp, including set negation ([^a-z])
- `\` — escapes the next metacharacter

Mask that starts with `*` or `**` patterns should be escaped with quotes in `werf.yaml` file:

- `"*.rb"` — with double quotes
- `'**/*'` — with single quotes

Werf determines whether the files changes in the git repository with use of checksums. For _user stage_ and for each mask, the following algorithm is applied:

- werf creates a list of all files from `add` path and apply `excludePaths` and `includePaths` filters
- each file path from the list compared to the mask with the use of [fnmatch](https://ruby-doc.org/core-2.2.0/File.html#method-c-fnmatch) with FNM_PATHNAME and FNM_PERIOD flags (`.` is included in the `*`, however `/` is excluded);
  - if fnmatch returns true, then the file is matched;
- if mask matches a directory then this directory content is matched recursively;
- werf calculates checksum of attributes and content of all matched files;

These checksums are calculated in the beginning of the build process before any stage container is ran.

Example:

```yaml
---
image: app
git:
- add: /src
  to: /app
  stageDependencies:
    beforeSetup:
    - "*"
shell:
  install:
  - echo "install stage"
  beforeSetup:
  - echo "beforeSetup stage"
```

This `werf.yaml` has a git path configuration to transfer `/src` content from local git repository into `/app` directory in the image. During the first build, files are cached in _gitArchive_ stage and assembly instructions for _install_ and _beforeSetup_ are executed. The next builds of commits that have only changes outside of the `/src` do not execute assembly instructions. If a commit has changes inside `/src` directory, then checksums of matched files are changed, and werf rebuilds _beforeSetup_ stage with applying a git patch.

## Dependency on CacheVersion values

There are situations when a user wants to rebuild all or one of _user stages_. This
can be accomplished by changing `cacheVersion` or `<user stage name>CacheVersion` values.

Signature of the _install_ user stage depends on the value of the
`installCacheVersion` parameter. To rebuild the _install_ user stage (and
subsequent stages), you need to change the value of the `installCacheVersion` parameter.

> Note that `cacheVersion` and `beforeInstallCacheVersion` directives have the same effect. When these values are changed, then the _beforeInstall_ stage and subsequent stages rebuilt.

### Example: common image for multiple applications

You can define an image with common packages in separated `werf.yaml`. `cacheVersion` value can be used to rebuild this image to refresh packages versions.

```yaml
image: ~
from: ubuntu:latest
shell:
  beforeInstallCacheVersion: 2
  beforeInstall:
  - apt update
  - apt install ...
```

This image can be used as base image for multiple applications if images from hub.docker.com doesn't suite your needs.

### External dependency example

_CacheVersion directives_ can be used with [go templates]({{ site.baseurl }}/reference/config.html#go-templates) to define _user stage_ dependency on files, not in the git tree.

{% raw %}
```yaml
image: ~
from: ubuntu:latest
shell:
  installCacheVersion: {{.Files.Get "some-library-latest.tar.gz" | sha256sum}}
  install:
  - tar zxf some-library-latest.tar.gz
  - <build application>
```
{% endraw %}

Build script can be used to download `some-library-latest.tar.gz` archive and then execute `werf build` command. If the file is changed then werf rebuilds _install user stage_ and subsequent stages.
