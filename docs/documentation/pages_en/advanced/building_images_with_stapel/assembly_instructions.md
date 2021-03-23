---
title: Running assembly instructions
permalink: advanced/building_images_with_stapel/assembly_instructions.html
directive_summary: shell_and_ansible
---

## What are user stages?

***User stage*** is a [_stage_]({{ "internals/stages_and_storage.html" | true_relative_url }}) containing _assembly instructions_ from the config.
Currently, there are two kinds of assembly instructions: _shell_ and _ansible_. werf provides four user stages and executes them in the following order: _beforeInstall_, _install_, _beforeSetup_, and _setup_. You can create the specific docker layer by executing assembly instructions contained within the respective stage.

## Motivation behind stages

### Opinionated build structure

The _user stages pattern_ is based on analysis of instructions for building real-life applications. It turns out that the categorization of assembly instructions into four _user stages_ is perfectly enough for most applications. Grouping decreases the size of layers and speeds up the image building process.

### Framework for a build process

The _user stages pattern_ defines the structure of a building process, and thus, sets boundaries for a developer. It distinguishes favorably from Dockerfile’s unstructured instructions since the developer has a good understanding of what instructions have to be included in each _stage_.

### Run an assembly instruction on git changes

The execution of a _user stage_ can depend on changes of files in a git repository. werf supports local and remote git repositories. The user stage can be dependent on changes in several repositories. Various changes in the repository can lead to a rebuild of different _user stages_.

### More build tools: shell, ansible, ...

_Shell_ is a familiar and well-known build tool. Ansible is newer, and you have to spend some time studying it.

If you need a prototype as soon as possible, then _the shell_ is your choice — it works like a `RUN` instruction in Dockerfile. The configuration of Ansible is declarative, mostly idempotent, and well suited for long-term maintenance.

The stage execution is isolated in code, so implementing support for another tool is not so difficult.

## Using user stages

werf provides four _user stages_ where assembly instructions can be defined. werf does not impose any restrictions on assembly instructions. You can specify the same variety of instructions as for the `RUN` instruction in Dockerfile. At the same time, the categorization of assembly instructions is based on our experience with real applications. So, the following actions are enough for building the vast majority of applications:

- install system packages
- install system dependencies
- install application dependencies
- setup system applications
- setup application

What is the best strategy to execute them? You might think the best way is to execute them one by one while caching interim results. On the other side, it is better not to mix instructions for these actions because of different file dependencies. The _user stages pattern_ suggests the following strategy:

- use the _beforeInstall_ user stage for installing system packages
- use the _install_ user stage to install system and application dependencies
- use the _beforeSetup_ user stage to configure system parameters and install an application
- use the _setup_ user stage to configure an application

### beforeInstall

This stage executes various instructions before installing an application. It is best suited for system applications that rarely change. At the same time, their installation process is very time-consuming. Also, at this stage, you can configure some system parameters that rarely change, such as setting a locale or a timezone, adding groups and users, etc. For example, you can install language distributions and build tools like PHP and composer, java and gradle, and so on, at this stage.

In practice, all these components rarely change, and the _beforeInstall_ stage caches
them for an extended period.

### install

This stage is for installing an application. It is best suited for installing application dependencies and configuring some basic settings.

Instructions on this stage have access to application source codes, so you can install application dependencies using build tools (like composer, gradle, npm, etc.) that require a manifest file (e.g., pom.xml, Gruntfile) to work. A best practice is to make this stage dependent on changes in that manifest file.

### beforeSetup

At this stage, you can prepare an application for tuning some parameters. It supports all kinds of compiling tasks: creating jars, creating executable files and dynamic libraries, creating web assets, uglification, and encryption. This stage is often made dependent on changes in the source code.

### setup

This stage is intended for configuring application settings. The corresponding set of actions includes copying some profiles into `/etc`, copying configuration files to already-known locations, creating a file containing the application version. These actions should not be time-consuming since they will likely be executed on every commit.

### custom strategy

Once again, no limitations are imposed on assembly instructions. The previous definitions of _user stages_ are just suggestions arising from our experience with real-world applications. You can even use merely a single _user stage_ or define your strategy for grouping assembly instructions and benefit from caching and git dependencies.

## Syntax

There are two top-level ***builder directives*** for assembly instructions that are mutually exclusive: `shell` and `ansible`. You can build an image either via ***shell instructions*** or via their ***ansible counterparts***.

The _builder directive_ includes four directives that define assembly instructions for each _user stage_:
- `beforeInstall`
- `install`
- `beforeSetup`
- `setup`

Builder directives can also contain ***cacheVersion directives*** that, in essence, are user-defined parts of _user-stage digests_. The detailed information is available in the [CacheVersion](#dependency-on-the-cacheversion-value) section.

## Shell

Here is the syntax for _user stages_ containing _shell assembly instructions_:

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

_Shell assembly instructions_ are made up of arrays. Each array consists of bash commands for the related _user stage_. Commands for each stage are executed as a single `RUN` instruction in Dockerfile. Thus, werf creates one layer for each _user stage_.

werf provides distribution-agnostic bash binary, so you do not need a bash binary in the [base image]({{ "advanced/building_images_with_stapel/base_image.html" | true_relative_url }}).

```yaml
beforeInstall:
- apt-get update
- apt-get install -y build-essential g++ libcurl4
```

werf performs _user stage_ commands as follows:
- generate the temporary script on host machine

    ```shell
    #!/.werf/stapel/embedded/bin/bash -e

    apt-get update
    apt-get install -y build-essential g++ libcurl4
    ```

- mount it to the corresponding _user stage assembly container_ as `/.werf/shell/script.sh`, and
- run the script.

> The `bash` binary is stored in a _stapel volume_. You can find additional information about the concept in the [blog post [RU]](https://habr.com/company/flant/blog/352432/) (`dappdeps` was later renamed to `stapel`; nevertheless, the principle is the same)

## Ansible

Here is the syntax for _user stages_ containing _ansible assembly instructions_:

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
this tasks via `ansible-playbook` command, werf mounts the directory structure presented below into the _user stage assembly container_:

```shell
/.werf/ansible-workdir
├── ansible.cfg
├── hosts
└── playbook.yml
```

`ansible.cfg` contains settings for ansible:
- use local transport (transport = local)
- werf's stdout_callback method for better logging (stdout_callback = werf)
- turn on the force_color mode (force_color = 1)
- use sudo for privilege escalation (to avoid using `become` in ansible tasks)

`hosts` is an inventory file that contains the localhost as well as some ansible_* settings, e.g., the path to python in stapel.

`playbook.yml` is a playbook with all tasks from the specific _user stage_. Here is an example of `werf.yaml` that includes the _install_ stage:

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

werf would generate the following `playbook.yml` for the _install_ stage:
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

werf plays the playbook of the _user stage_ in the _user stage assembly container_ via the `playbook-ansible`
command:

```shell
$ export ANSIBLE_CONFIG="/.werf/ansible-workdir/ansible.cfg"
$ ansible-playbook /.werf/ansible-workdir/playbook.yml
```

`ansible` and `python` binaries/libraries are stored in a _stapel volume_. You can find more information about this concept in [this blog post [RU]](https://habr.com/company/flant/blog/352432/) (`dappdeps` was later renamed to `stapel`; nevertheless, the principle is the same).

### Supported modules

One of the ideas at the core of werf is idempotent builds. werf must generate the very same image every time if there are no changes. We solve this task by calculating a _digest_ for _stages_. However, ansible’s modules are non-idempotent, meaning they produce different results even if the input parameters are the same. Thus, werf is unable to correctly calculate a _digest_ in order to determine the need to rebuild _stages_. Because of that, werf currently supports a limited list of modules:

- [Command modules](https://docs.ansible.com/ansible/2.5/modules/list_of_commands_modules.html): command, shell, raw, script.
- [Crypto modules](https://docs.ansible.com/ansible/2.5/modules/list_of_crypto_modules.html): openssl_certificate, and other.
- [Files modules](https://docs.ansible.com/ansible/2.5/modules/list_of_files_modules.html): acl, archive, copy, stat, tempfile, and other.
- [Net Tools Modules](https://docs.ansible.com/ansible/2.5/modules/list_of_net_tools_modules.html): get_url, slurp, uri.
- [Packaging/Language modules](https://docs.ansible.com/ansible/2.5/modules/list_of_packaging_modules.html#language): composer, gem, npm, pip, and other.
- [Packaging/OS modules](https://docs.ansible.com/ansible/2.5/modules/list_of_packaging_modules.html#os): apt, apk, yum, and other.
- [System modules](https://docs.ansible.com/ansible/2.5/modules/list_of_system_modules.html): user, group, getent, locale_gen, timezone, cron, and other.
- [Utilities modules](https://docs.ansible.com/ansible/2.5/modules/list_of_utilities_modules.html): assert, debug, set_fact, wait_for.

An attempt to do a _werf config_ with the module not in this list will lead to an error, and a failed build. Feel free to report an [issue](https://github.com/werf/werf/issues/new) if some module should be enabled.

### Copying files

[Git mappings]({{ "advanced/building_images_with_stapel/git_directive.html" | true_relative_url }}) are the preferred way of copying files into an image. werf cannot detect changes to the files referred to in the `copy` module. Currently, the only way to copy some external file into an image involves using the `.Files.Get` method of Go templates. This method returns the contents of the file as a string. Thus, the contents become a part of the _user stage digest_, and file changes lead to the rebuild of the _user stage_.

Here is an example of copying `nginx.conf` into an image:

{% raw %}
```yaml
ansible:
  install:
  - copy:
      content: |
{{ .Files.Get "/conf/etc/nginx.conf" | indent 8 }}
      dest: /etc/nginx/nginx.conf
```
{% endraw %}

werf renders that snippet as a go template and then transforms it into the following `playbook.yml`:

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

### Jinja templates

Ansible supports [Jinja templates](https://docs.ansible.com/ansible/2.5/user_guide/playbooks_templating.html) in playbooks. However, Go templates and Jinja templates have the same delimiters: {% raw %}{{ and }}{% endraw %}. Thus, you have to escape Jinja templates to use them. There are two possible solutions: you can escape {% raw %}{{{% endraw %} delimiters only or escape the whole Jinja expression.

Let’s take a look at the example. Say, you have the following ansible task:

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
In this case, the Jinja expression `{{item}}` must be escaped:
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

### Ansible complications

- Only raw and command modules support Live stdout output. Other modules display contents of stdout and stderr streams after execution.
- The `apt` module hangs the build process in some debian and ubuntu versions. The derived images are affected as well ([issue #645](https://github.com/werf/werf/issues/645)).

## Dependencies of user stages

werf features the ability to define dependencies for rebuilding the _stage_. As described in the [_stages_ reference]({{ "internals/stages_and_storage.html" | true_relative_url }}), _stages_ are built one by one, and the _digest_ is calculated for each _stage_. _Digests_ have various dependencies. When dependencies change, the _stage digest_ changes as well. As a result, werf rebuilds this _stage_ and all the subsequent _stages_.

You can use these dependencies to shape the rebuilding process of _user stages_. _Digests_ of user stages  (and, therefore, the rebuilding process) depend on:

- changes in assembly instructions
- changes of _cacheVersion directives_
- changes in the git repository
- changes in files being imported from [artifacts]({{ "advanced/building_images_with_stapel/artifacts.html" | true_relative_url }})

The first three dependencies are described below in more detail.

## Dependency on changes in assembly instructions

The _digest_ of the user stage depends on the rendered text of assembly instructions. Changes in assembly instructions for the _user stage_ lead to the rebuilding of this _stage_. Say, you have the following _shell-based assembly instructions_:

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

On the first build of this image, instructions for all four user stages will be executed. There is no _git mapping_ in this _config_, so assembly instructions will never be executed on subsequent builds since _digests_ of user stages will be the same, and the build cache will remain valid.

Let us change assembly instructions for the _install_ user stage:

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

The digest of the install stage has changed, so `werf build` executes assembly instructions in the _install_ stage and instructions defined in subsequent _stages_, i.e., _beforeSetup_ and _setup_.

The stage digest may also change due to the use of environment variables and Go templates and that can lead to unforeseen rebuilds. For example:

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

The first build will calculate the _digest_ of the _beforeInstall_ stage:

```shell
echo "Commands on the Before Install stage for 0a8463e2ed7e7f1aa015f55a8e8730752206311b"
```

The _digest_ of the _beforeInstall_ stage will change with each subsequent commit:

```shell
echo "Commands on the Before Install stage for 36e907f8b6a639bd99b4ea812dae7a290e84df27"
```

In other words, the contents of assembly instructions will change with each subsequent commit because of the `CI_COMMIT_SHA` variable. Thus, such a configuration leads to the rebuild of the _beforeInstall_ user stage on every commit.

## Dependency on changes in the git repo

<a class="google-drawings" href="{{ "images/configuration/assembly_instructions3.png" | true_relative_url }}" data-featherlight="image">
    <img src="{{ "images/configuration/assembly_instructions3_preview.png" | true_relative_url }}" alt="Dependency on git repo changes">
</a>

The _git mapping reference_ states that there are _gitArchive_ and _gitLatestPatch_ stages. _gitArchive_ runs after the _beforeInstall_ user stage, and _gitLatestPatch_ runs after the _setup_ user stage if there are changes in the local git repository. Thus, in order to execute assembly instructions using the latest version of the source code, you can initiate the rebuilding of the _beforeInstall_ stage (by changing _cacheVersion_ or its instructions).

_install_, _beforeSetup_, and _setup_ user stages also depend on changes in the git repository. In this case, a git patch is applied at the beginning of the _user stage_, and assembly instructions are executed using the latest version of the source code.

> During the process of building an image, the source code is updated **only at one of the stages**; all subsequent stages are based on this stage and thus use the actualized files. The source files contained in the git repository are added with the first build during the _gitArchive_ stage. All subsequent builds update source files during _gitCache_, _gitLatestPatch_ stages, or during one of the following user stages: _install_, _beforeSetup_, _setup_.

<br />
<br />
This stage is pictured on the _Calculation digest phase_
![git files actualized on specific stage]({{ "images/build/git_mapping_updated_on_stage.png" | true_relative_url }})

You can specify the dependency of the _user stage_ on changes in the git repository via the `git.stageDependencies` parameter. It has the following syntax:

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

The `git.stageDependencies` parameter has 3 keys: `install`, `beforeSetup` and `setup`. Each key defines an array of masks for a single user stage. The _user stage_ will be rebuilt if there are changes in the git repository that match one of the masks defined for the _user stage_.

For each _user stage_, werf creates a list of matching files and calculates a checksum over attributes and contents of each file. This checksum is a part of the _stage digest_. Thus, the digest changes in response to any changes in the repository, such as getting new attributes for the file, changing its contents, adding new matching file, deleting a matching file, etc.

`git.stageDependencies` masks work jointly with `git.includePaths` and `git.excludePaths` masks. Only files that match the `includePaths` filter and `stageDependencies` masks are considered suitable. Similarly, only files that do not match the `excludePaths` filter and `stageDependencies` masks are considered suitable by werf.

`stageDependencies` masks work similarly to `includePaths` and `excludePaths` filters. The mask defines a template for files and paths and may contain the following glob patterns:

- `*` — matches any file. This pattern includes `.` and excludes `/`
- `**` — matches directories recursively or files expansively
- `?` — matches any single character. It is equivalent to /.{1}/ in regexp
- `[set]` — matches any character within the set. It behaves exactly like character sets in regexp, including set negation ([^a-z])
- `\` — escapes the next metacharacter

Mask that starts with `*` is treated as an anchor name by the yaml parser. Thus, masks starting with `*` or `**` patterns at the beginning must be surrounded by quotation marks:

```yaml
# * at the beginning of mask, so use double quotation marks
- "*.rb"
# single quotation marks also work
- '**/*'
# no star at the beginning, no quotation marks are needed
- src/**/*.js
```

werf finds out whether files have been changed in the git repository by calculating checksums. It applies the following algorithm for the _user stage_ and for each mask:

- create a list of all files at the `add` path and apply the `excludePaths` and `includePaths` filters:
- compare path of each file in the list to the mask using of glob patterns;
- if some directory matches a mask, then all contents of this directory are considered matching recursively;
- calculate the checksum of attributes and contents of all matching files.

These checksums are calculated at the beginning of the build process before any stage container is being run.

Example:

```yaml
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
  setup:
  - echo "setup stage"
```

The _git mapping configuration_ in the above `werf.yaml` requires werf to transfer the contents of the `/src` directory of the local git repository to the `/app` directory of the image. During the first build, files are cached at the _gitArchive_ stage, and assembly instructions for _install_ and _beforeSetup_ are executed. During the builds triggered by the subsequent commits that do not change he contents of the `/src` directory, werf does not execute assembly instructions. If there were changes in the `/src` directory because of some commit, then checksums of files matching the mask would change. As a result, werf would apply the git patch and rebuild all the existing stages beginning with _beforeSetup_, namely _beforeSetup_ and _setup_. The git patch will be applied once during the _beforeSetup_ stage.

## Dependency on the CacheVersion value

There are situations when a user wants to rebuild all or just one _user stage_. This
can be accomplished by changing `cacheVersion` or `<user stage name>CacheVersion` values.

The digest of the _install_ user stage depends on the value of the
`installCacheVersion` parameter. To rebuild the _install_ user stage (and
subsequent stages), you need to change the value of the `installCacheVersion` parameter.

> Note that `cacheVersion` and `beforeInstallCacheVersion` directives have the same effect. Changing them triggers the rebuild of the _beforeInstall_ stage and all subsequent stages.

### Example. The universal image for multiple applications

An image containing shared system packages can be defined in a separate `werf.yaml` file. You can use the `cacheVersion` value for rebuilding this image to refresh packages’ versions.

```yaml
image: ~
from: ubuntu:latest
shell:
  beforeInstallCacheVersion: 2
  beforeInstall:
  - apt update
  - apt install ...
```

You can use this image as a base for multiple applications if images from _hub.docker.com_ do not quite suit your needs.

### Example of using external dependencies

You can use _CacheVersion directives_ jointly with [go templates]({{ "reference/werf_yaml.html#image-section" | true_relative_url }}) to define dependency of the _user stage_ on files outside the git tree.

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

The build script can be used to download `some-library-latest.tar.gz` archive and then execute the `werf build` command. Any changes to the file trigger the rebuild of the _install user stage_ and all the subsequent stages.
