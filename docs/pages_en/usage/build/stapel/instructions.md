---
title: Running assembly instructions
permalink: usage/build/stapel/instructions.html
directive_summary: shell_and_ansible
---

## What are user stages?

***User stage*** is a stage containing _assembly instructions_ from the config.

Currently, there are two kinds of assembly instructions: _shell_ and _ansible_. werf provides four user stages and executes them in the following order: _beforeInstall_, _install_, _beforeSetup_, and _setup_. You can create a specific Docker layer by executing assembly instructions in the corresponding stage.

## Using user stages

werf provides four _user stages_ where assembly instructions can be defined. werf does not impose any restrictions on assembly instructions. You can specify the same variety of instructions as for the `RUN` instruction in Dockerfile. At the same time, the categorization of assembly instructions is based on our experience with real-world applications. So, the following actions are enough for building the vast majority of applications:

- installing system packages;
- installing system dependencies;
- installing application dependencies;
- setting up system applications;
- setting up an application.

What is the best strategy for carrying them out? You might think the best way is to run them one by one, caching the interim results. On the other hand, it is better not to mix instructions for these actions because of different file dependencies. The _user stages pattern_ suggests the following strategy:

- use the _beforeInstall_ user stage for installing system packages;
- use the _install_ user stage to install system and application dependencies;
- use the _beforeSetup_ user stage to configure system parameters and install an application;
- use the _setup_ user stage to configure an application.

<div class="details">
<a href="javascript:void(0)" class="details__summary">How the Stapel stage assembly works</a>
<div class="details__content" markdown="1">

When building a stage, the stage instructions are supposed to run in a container based on the previous built stage or [base image]({{"usage/build/stapel/base.html#from-fromlatest" | true_relative_url }}). We will further refer to such a container as a **build container**.

Before running the _build container_, werf prepares a set of instructions. This set depends on the stage type and contains both werf service commands and user commands specified in the `werf.yaml` config file. The service commands may include, for example, adding files, applying patches, running Ansible jobs, etc.

The Stapel builder uses its own set of tools and libraries and does not depend on the base image in any way. When the _build container_ is started, werf mounts everything it needs from the special service image named `registry.werf.io/werf/stapel`.

The _build container_ [gets the socket to communicate with the SSH-agent on the host](#how-to-use-the-ssh-agent-in-build-instructions); [custom mounts]({{"usage/build/stapel/mounts.html" | true_relative_url }}) can also be used.



It is also worth noting that werf ignores some of the base image manifest parameters when building, replacing them with the following values:
- `--user=0:0`;
- `--workdir=/`;
- `--entrypoint=/.werf/stapel/embedded/bin/bash`.

So the start of the _build container_ of some arbitrary stage typically looks as follows:

```shell
docker run \
  --volume=/tmp/ssh-ln8yCMlFLZob/agent.17554:/.werf/tmp/ssh-auth-sock \
  --volumes-from=stapel_0.6.1 \
  --env=SSH_AUTH_SOCK=/.werf/tmp/ssh-auth-sock \
  --user=0:0 \
  --workdir=/ \
  --entrypoint=/.werf/stapel/embedded/bin/bash \
  sha256:d6e46aa2470df1d32034c6707c8041158b652f38d2a9ae3d7ad7e7532d22ebe0 \
  -ec eval $(echo c2V0IC14 | /.werf/stapel/embedded/bin/base64 --decode)
```

</div>
</div>

### beforeInstall

```yaml
shell:
  beforeInstall:
    - apt update -q
    - apt install -y curl mysql-client libmysqlclient-dev g++ build-essential libcurl4
  beforeInstallCacheVersion: "1"
```

This stage executes various instructions before installing an application. It is best suited for system applications that rarely change. At the same time, their installation process is very time-consuming. Also, at this stage, you can configure some system parameters that rarely change, such as a locale or a timezone, add groups and users, etc. For example, you can install language distributions and build tools like PHP and Composer, Java and Gradle, and so on.

Since these components rarely change, they will be cached by the _beforeInstall_ stage for an extended period.

`beforeInstallCacheVersion: <string>` — an optional directive to invalidate the build cache of a given stage in a deterministic way based on changes introduced by the Git commit.

### install

```yaml
shell:
  install:
    - bundle install
    - npm ci
  installCacheVersion: "1"
```

This stage is best suited for installing an application and its dependencies and performing some basic configuration.

This stage has access to the application source code in Git, so you can install application dependencies using build tools (e.g., Composer, Gradle, npm, etc.) that require a manifest file (e.g., pom.xml, Gruntfile) to work. A best practice is to make this stage dependent on changes in that manifest file.

`installCacheVersion: <string>` — an optional directive to invalidate the build cache of a given stage in a deterministic way based on changes introduced by the Git commit.

### beforeSetup

```yaml
shell:
  beforeSetup:
    - rake assets:precompile
  beforeSetupCacheVersion: "1"
```

This stage allows you to prepare your application to customize some parameters. It supports all kinds of compiling tasks: creating jars, creating executable files and dynamic libraries, creating web assets, uglification, and encryption. This stage is often made dependent on changes in the source code.

`beforeSetupCacheVersion: <string>` — an optional directive to invalidate the build cache of a given stage in a deterministic way based on changes introduced by the Git commit.

### setup

```yaml
shell:
  setup:
    - npm run build
  setupCacheVersion: "1"
```

This stage deals with the application settings. A typical set of actions includes copying some profiles into `/etc`, copying configuration files to already-known locations, creating a file containing the application version. These actions should not be time-consuming since they will likely be performed on every commit.

`setupCacheVersion: <string>` — an optional directive to invalidate the build cache of a given stage in a deterministic way based on changes introduced by the Git commit.

### Custom strategy

No limitations are imposed on assembly instructions. The suggested use of _user stages_ is only a recommendation based on our experience with real-world applications. You can use just one _user stage_, or you can design your own instruction grouping strategy to take advantage of caching and dependency changes in the Git repositories tailored to how your application is built.

## Syntax

There are two mutually exclusive top-level ***builder directives*** for assembly instructions: `shell` and `ansible`. You can build an image either via ***shell instructions*** or via their ***ansible counterparts***.

The _builder directive_ includes four directives that define assembly instructions for each _user stage_:

- `beforeInstall`;
- `install`;
- `beforeSetup`;
- `setup`.

Builder directives can also contain ***cacheVersion directives*** that, in essence, are user-defined parts of _user-stage digests_. The detailed information is available in the [CacheVersion](#dependency-on-the-cacheversion) section.

## Shell

Here is the example of the _user stage_ syntax featuring _shell assembly instructions_:

```yaml
shell:
  beforeInstall:
  - <bash_command 1>
  - <bash_command 2>
  # ...
  - <bash_command N>
  install:
  - bash command
  # ...
  beforeSetup:
  - bash command
  # ...
  setup:
  - bash command
  # ...
  cacheVersion: <version>
  beforeInstallCacheVersion: <version>
  installCacheVersion: <version>
  beforeSetupCacheVersion: <version>
  setupCacheVersion: <version>
```

_Shell assembly instructions_ are made up of arrays. Each array includes Bash commands for the corresponding _user stage_. Commands for each stage are executed as a single `RUN` instruction in Dockerfile. Thus, a single layer is created for each _user stage_.

werf provides distribution-agnostic Bash binary, so you do not need to add it to the [base image]({{ "usage/build/stapel/base.html" | true_relative_url }}).

```yaml
beforeInstall:
- apt-get update
- apt-get install -y build-essential g++ libcurl4
```

The `bash` binary is stored in a _Stapel volume_. You can find additional information about the concept in this [blog post [RU]](https://habr.com/company/flant/blog/352432/) (`dappdeps` has been renamed to `stapel`; still, the principle remains the same)

## Ansible

Here is the _user stage_ syntax featuring _ansible assembly instructions_:

```yaml
ansible:
  beforeInstall:
  - <ansible task 1>
  - <ansible task 2>
  # ...
  - <ansible task N>
  install:
  - ansible task
  # ...
  beforeSetup:
  - ansible task
  # ...
  setup:
  - ansible task
  # ...
  cacheVersion: <version>
  beforeInstallCacheVersion: <version>
  installCacheVersion: <version>
  beforeSetupCacheVersion: <version>
  setupCacheVersion: <version>
```

> **Note:** the ansible syntax is not available for the Buildah building backend.

### Ansible config and stage playbook

_Ansible assembly instructions_ for the _user stage_ are a set of Ansible tasks.

The generated `ansible.cfg` file contains the Ansible settings:
- use local transport (transport = local);
- werf's stdout_callback method for better logging (stdout_callback = werf);
- turn on the force_color mode (force_color = 1);
- use sudo for privilege escalation (to avoid using `become` in ansible tasks).

The generated `playbook.yml` file is a playbook with all tasks from the specific _user stage_. Below is an example of `werf.yaml` that includes all the tasks for the _install_ stage:

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
  # ...
```

`ansible` and `python` binaries/libraries are stored in a _Stapel volume_. You can find more information about this concept in [this blog post [RU]](https://habr.com/company/flant/blog/352432/) (`dappdeps` has been renamed to `stapel`; still, the principle remains the same).

### Supported modules

One of the ideas at the core of werf is idempotent builds. werf must generate the identical image every time as long as there are no changes. We achieve this by calculating a _digest_ for _stages_. However, Ansible modules are non-idempotent, meaning they produce different results even if the input parameters are the same. Consequently, werf cannot correctly calculate a _digest_ to determine whether _stages_ need to be rebuilt. This is why werf currently supports a limited list of modules:

- Command modules: command, shell, raw, script.
- Crypto modules: openssl_certificate, and other.
- Files modules: acl, archive, copy, stat, tempfile, and other.
- Net Tools Modules: get_url, slurp, uri.
- Packaging/Language modules: composer, gem, npm, pip, and other.
- Packaging/OS modules: apt, apk, yum, and other.
- System modules: user, group, getent, locale_gen, timezone, cron, and other.
- Utilities modules: assert, debug, set_fact, wait_for.

An attempt to do a _werf config_ with the module not in this list will result in an error and a failed build. Feel free to create an [issue](https://github.com/werf/werf/issues/new) if you think some module should be enabled.

### Copying files

[Git mappings]({{ "usage/build/stapel/git.html" | true_relative_url }}) are the preferred way of copying files into an image. werf cannot detect changes to the files referred to in the `copy` module. Currently, the only way to copy some external file into an image is to use the `.Files.Get` method of Go templates. This method returns the contents of the file as a string. Thus, the contents become a part of the _user stage digest_, and changes to the file cause the _user stage_ to be rebuilt.

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

werf will render the above snippet as a go template and transform it into the following `playbook.yml`:

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
            # ...
```

### Jinja templates

Ansible supports [Jinja templates](https://docs.ansible.com/ansible/2.5/user_guide/playbooks_templating.html) in playbooks. Unfortunately, Go and Jinja templates have similar delimiters: {% raw %}{{ and }}{% endraw %}. So you have to escape Jinja templates to use them. There are two possible ways to do that: you can either escape the {% raw %}{{{% endraw %} delimiters or the entire Jinja expression.

Let’s take a look at the example. Say, you have the following Ansible task:

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
In this case, the `{{item}}` Jinja expression must be escaped:
{% endraw %}

{% raw %}
```yaml
# Escape {{ only.
src: {{"{{"}} item }}
```
or
```yaml
# Escape the whole expression.
src: {{`{{item}}`}}
```
{% endraw %}

### Ansible complications

- Only raw and command modules support Live stdout output. Other modules display contents of stdout and stderr streams after execution, which results in output delays.
- The `apt` module causes the build process to hang in some Debian and Ubuntu versions. The derived images are affected as well ([issue #645](https://github.com/werf/werf/issues/645)).

## Environment variables of the build container

You can use service environment variables which are available in build container during the build. They can be used in your shell assembly instructions. Using them will not affect the build instructions and will not trigger stage rebuilds, even if these service environment variables change.

The following environment variables are available:
- `WERF_COMMIT_HASH`. An example of value: `cda9d17265d174c62424e8f7b5e5640bf749c565`.
- `WERF_COMMIT_TIME_HUMAN`. An example of value: `2022-01-24 17:26:19 +0300 +0300`.
- `WERF_COMMIT_TIME_UNIX`. An example of value: `1643034379`.

Usage example:
{% raw %}
```yaml
shell:
  install:
  - echo "Commands on the Install stage for $WERF_COMMIT_HASH"
```
{% endraw %}

In the example above the current commit hash will be inserted into the `echo ...` command, but this will happen in the very last moment — when the build instructions will be interpreted and executed by the shell. This way there will be no "install" stage rebuilds on every commit.

## User stage dependencies

werf features the ability to define the dependencies that will cause the _stage_ to be rebuilt. _Stages_ are built sequentially, and the _digest_ is calculated for each _stage_. _Digests_ have various dependencies. When those dependencies change, the _stage digest_ changes as well. As a result, werf rebuilds the affected _stage_ and all the subsequent _stages_.

You can use these dependencies to shape the rebuilding process of the _user stages_. The _digests_ of the _user stages_ (and, consequently, the rebuilding process) depend on:

- changes in the assembly instructions;
- changes of the _cacheVersion directives_;
- changes in the Git repository;
- changes in the files imported from the [artifacts]({{ "usage/build/stapel/imports.html#what-is-an-artifact" | true_relative_url }}).

The first three dependency types are described in more detail below.

## Dependency on changes in the assembly instructions

The _digest_ of the user stage depends on the rendered text of the assembly instructions. Changes in the assembly instructions for the _user stage_ lead to the rebuilding of this _stage_. Suppose you have the following _shell-based assembly instructions_:

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

During the first build of this image, instructions for all four user stages will be executed. There is no _git mapping_ in this _config_, so the assembly instructions will never be executed on subsequent builds, since _digests_ of user stages will be the same, and the build cache will remain valid.

Let us change the assembly instructions for the _install_ user stage:

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

The digest of the install stage has changed, so running `werf build` will result in executing the assembly instructions in the _install_ stage as well as the instructions defined in subsequent _stages_, i.e., _beforeSetup_ and _setup_.

The stage digest may also change due to the use of environment variables and Go templates, resulting in unforeseen rebuilds:

{% raw %}
```yaml
shell:
  beforeInstall:
  - echo "Commands on the Before Install stage for {{ env "CI_COMMIT_SHA” }}"
  install:
  - echo "Commands on the Install stage"
  # ...
```
{% endraw %}

In the example above, the _digest_ of the _beforeInstall_ stage will be calculated during the first build:

```shell
echo "Commands on the Before Install stage for 0a8463e2ed7e7f1aa015f55a8e8730752206311b"
```

The _digest_ of the _beforeInstall_ stage will change with each subsequent commit:

```shell
echo "Commands on the Before Install stage for 36e907f8b6a639bd99b4ea812dae7a290e84df27"
```

In other words, the contents of the assembly instructions will change with each subsequent commit because of the `CI_COMMIT_SHA` variable. So this configuration causes the _beforeInstall_ user stage to be rebuilt at each commit.

## Dependency on changes in the Git repo

<a class="google-drawings" href="{{ "images/configuration/assembly_instructions3.svg" | true_relative_url }}" data-featherlight="image">
    <img src="{{ "images/configuration/assembly_instructions3.svg" | true_relative_url }}" alt="Dependency on git repo changes">
</a>

The _git mapping reference_ states that there are _gitArchive_ and _gitLatestPatch_ stages. _gitArchive_ runs after the _beforeInstall_ user stage, and _gitLatestPatch_ runs after the _setup_ user stage if there are changes in the local Git repository. Thus, in order to run the assembly instructions on the latest source code version, you can initiate the rebuilding of the _beforeInstall_ stage (by changing _cacheVersion_ or its instructions).

The _install_, _beforeSetup_, and _setup_ user stages also depend on changes in the Git repository. In this case, a git patch is applied at the beginning of the _user stage_, and the assembly instructions are executed on the latest version of the source code.

> During the process of building an image, the source code is updated **only at one of the stages**; all subsequent stages are based on this stage and thus use the actualized files. The source files contained in the Git repository are added during the first build at the _gitArchive_ stage. All subsequent builds update the source files at _gitCache_, _gitLatestPatch_ stages, or at one of the following user stages: _install_, _beforeSetup_, _setup_.

<br />
<br />
This stage is depicted as of the _Calculation digest phase_
![git files actualized on specific stage]({{ "images/build/git_mapping_updated_on_stage.png" | true_relative_url }})

You can specify the dependency of the _user stage_ on Git repository changes using the `git.stageDependencies` parameter. It has the following syntax:

```yaml
git:
- ...
  stageDependencies:
    install:
    - <mask 1>
    # ...
    - <mask N>
    beforeSetup:
    - <mask>
    # ...
    setup:
    - <mask>
```

The `git.stageDependencies` parameter has 3 keys: `install`, `beforeSetup`, and `setup`. Each key defines an array of masks for a single user stage. The _user stage_ will be rebuilt if there are changes in the Git repository that match one of the masks defined for the _user stage_.

For each _user stage_, werf creates a list of matching files and calculates a checksum based on the attributes and contents of each file. This checksum is a part of the _stage digest_. Thus, the digest changes in response to any changes in the repository, such as getting new file attributes, changing file contents, adding or deleting a new matching file, etc.

The `git.stageDependencies` masks work jointly with the `git.includePaths` and `git.excludePaths` masks. Only files that match the `includePaths` filter and the `stageDependencies` masks are considered eligible. And vice versa: only files that do not match the `excludePaths` filter and the `stageDependencies` masks are considered suitable by werf.

The `stageDependencies` masks work in the same fashion as the `includePaths` and `excludePaths` filters. The mask defines a file and path template and may contain the following glob patterns:

- `*` — matches any file. This pattern includes `.` and excludes `/`;
- `**` — matches directories recursively or files expansively;
- `?` — matches any single character. It is equivalent to /.{1}/ in regexp;
- `[set]` — matches any character within the set. It behaves exactly like character sets in regexp, including set negation ([^a-z]);
- `\` — escapes the next metacharacter.

A mask starting with `*` is treated as an anchor name by the YAML parser. Thus, masks starting with `*` or `**` patterns at the beginning must be surrounded by quotation marks:

```yaml
# * at the beginning of mask, so use double quotation marks
- "*.rb"
# single quotation marks also work
- '**/*'
# no star at the beginning, no quotation marks are needed
- src/**/*.js
```

werf finds out whether the files have been changed in the Git repository by calculating checksums. It applies the following algorithm to the _user stage_ and to each mask:

- create a list of all files in the `add` path and apply the `excludePaths` and `includePaths` filters;
- compare the path of each file in the list to the mask using the glob patterns;
- if some directory matches a mask, then all the contents of that directory are considered to be matching recursively;
- calculate checksums of attributes and contents of all matching files.

These checksums are calculated at the beginning of the build process before any stage container is run.

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

The _git mapping configuration_ in the above `werf.yaml` instructs werf to transfer the contents of the `/src` directory of the local Git repository to the `/app` directory of the image. During the first build, files will be cached at the _gitArchive_ stage, and assembly instructions for _install_ and _beforeSetup_ will be executed. During the builds triggered by the subsequent commits which leave the contents of the `/src` directory unchanged, werf will not run the assembly instructions. Changes in the `/src` directory due to some commit will also result in changes in the checksums of the files matching the mask. This will cause werf to apply the git patch and rebuild any existing stages starting with _beforeSetup_, namely _beforeSetup_ and _setup_. The git patch will be applied once during the _beforeSetup_ stage.

## Dependency on the CacheVersion

There are situations when a user wants to rebuild all _user stages_ or just one of them. They can do so by changing `cacheVersion` or `<user stage name>CacheVersion` parameters.

The digest of the _install_ user stage depends on the value of the `installCacheVersion` parameter. To rebuild the _install_ user stage (and
subsequent stages), you have to change the value of the `installCacheVersion` parameter.

> Note that the `cacheVersion` and `beforeInstallCacheVersion` directives have the same effect. Changing them will cause the _beforeInstall_ stage and all subsequent stages to be rebuilt.

### Example: A universal image for multiple applications

An image containing shared system packages can be defined in a separate `werf.yaml` file. You can use the `cacheVersion` parameter for rebuilding this image to update the package versions.

```yaml
image: ~
from: ubuntu:latest
shell:
  beforeInstallCacheVersion: 2
  beforeInstall:
  - apt update
  - apt install ...
```

This image can be used as a base for multiple applications if images from _hub.docker.com_ do not suit your needs.

### Example of using external dependencies

You can use _CacheVersion directives_ jointly with the Go templates to define dependencies of a _user stage_ on files outside the Git tree.

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

The build script can be used to download the `some-library-latest.tar.gz` archive and then run the `werf build` command. Any changes to the file will trigger the rebuild of the _install user stage_ and all the subsequent stages.

## How to use the SSH agent in build instructions

Working with remote Git repositories relies on the `SSH_AUTH_SOCK` UNIX socket, which is mounted in all build containers. This way, the build instructions can use the SSH agent over the specified UNIX socket.

> **NOTE:** There is a restriction that only the `root` user inside the build container can access the UNIX socket defined by the `SSH_AUTH_SOCK` environment variable.

By default (if no parameters are specified), werf tries to use the SSH-agent running on the system by checking its availability via the `SSH_AUTH_SOCK` environment variable.

If no SSH agent is running on the system, werf tries to act as an SSH client, using the user's default SSH key (`~/.ssh/id_rsa|id_dsa`). If werf finds one of these files, it runs a temporary SSH agent and adds to it the keys it has found.

You have to specify the `--ssh-key PRIVATE_KEY_FILE_PATH` startup option to use specific SSH keys only (you can use this option more than once to specify multiple SSH keys). In this case, werf runs a temporary SSH agent and adds to it only the specified SSH keys.

### Temporary SSH agent

werf can start a temporary SSH agent to run some commands. Such an SSH agent is terminated when the corresponding werf command finishes its work.

If there is an SSH agent running on the system, the temporary SSH agent started by werf does not conflict with the one running on the system.
