---
title: YAML directives
sidebar: not_used
permalink: not_used/
---

# YAML config

Configuration is a collection of YAML documents (http://yaml.org/spec/1.2/spec.html#id2800132). These YAML documents are searched for in one of the following files:

* `REPO_ROOT/werf.yaml`
* `REPO_ROOT/werf.yaml`

YAML configuration file will precede ruby `REPO_ROOT/Werf config` if the case both files exist. `REPO_ROOT/werf.yaml` will precede `REPO_ROOT/werf.yaml` in case both files exist.

Processing of YAML configuration is done in two steps:

* Rendering go templates into `WORKDIR/.config.render.yml` or `WORKDIR/.config.render.yaml`.
* Processing the result file as a set of YAML documents.

## Go templates

Go templates are available within YAML config.

* Sprig functions supported: https://golang.org/pkg/text/template/, http://masterminds.github.io/sprig/.
* `env` sprig function also supported to access build-time environment variables (unlike helm, where `env` function is forbidden).

Werf firstly will render go templates into `WORKDIR/.config.render.yml` or `WORKDIR/.config.render.yaml`. That file will remain after build and will be available if some validation or build error occurs.

## Differences from Werf config

1. No equivalent for `dimg_group` and nested `dimg`s and `artifact`s.
2. No context inheritance because of 1. Use go-template functionality
   to define common parts.
3. Use `import` in `dimg` for copy artifact results instead of `export`
4. Each `artifact` must have a name

## Configuration

```
YAML_DOC
---
YAML_DOC
---
...
```

Each YAML_DOC is either `dimg` or `artifact` and contains all instructions to build docker image.

One of `dimg: NAME` or `artifact: NAME` is a required parameter for each document.

## Directives

### Basic

#### `dimg: NAME` (one of required)

If doc contains `dimg` key, then werf will treat this yaml-doc as dimg configuration.

`NAME` may be:

* Special value `~` to define unnamed dimg. This should be default choice for dimg name in config.
* Some string to define named dimg with the specified name. The name should be a valid docker image name.
* Array of strings to define multiple dimgs with the same configuration and different names. The names should be valid docker image names.

Conflicts with `artifact: NAME`.

#### `artifact: NAME` (one of required)

If doc contains `artifact` key, werf will treat this yaml-doc as artifact configuration.

`NAME` is a string, that defines the artifact name. An artifact may be referenced in `import` and `fromArtifact` directives by that name.

Conflicts with `dimg: NAME`.

#### `from: DOCKER_IMAGE` (required)

Specify docker image as a base image for current dimg or artifact.

#### `docker: DOCKER_PARAMS_MAP`

Define a map with docker image parameters. Supported parameters are:

* CMD (https://docs.docker.com/engine/reference/builder/#/cmd)
* ENV (https://docs.docker.com/engine/reference/builder/#/env)
* ENTRYPOINT (https://docs.docker.com/engine/reference/builder/#/entrypoint)
* EXPOSE (https://docs.docker.com/engine/reference/builder/#/expose)
* LABEL (https://docs.docker.com/engine/reference/builder/#/label)
* ONBUILD (https://docs.docker.com/engine/reference/builder/#/onbuild)
* USER (https://docs.docker.com/engine/reference/builder/#/user)
* VOLUME (https://docs.docker.com/engine/reference/builder/#/volume)
* WORKDIR (https://docs.docker.com/engine/reference/builder/#/workdir)

Example:

```
docker:
  WORKDIR: /app
  EXPOSE: '8080'
  USER: app
```

#### `import: IMPORTS_ARR`

`IMPORTS_ARR` is an array, where each element is a map:

```
artifact: ARTIFACT_NAME
add: SOURCE_DIRECTORY_TO_IMPORT
to: DESTINATION_DIRECTORY # optional
before: STAGE | after: STAGE
```

* `ARTIFACT_NAME` refers to the artifact to copy files from.
* `SOURCE_DIRECTORY_TO_IMPORT` specifies a directory/file path in the artifact that should be imported.
* `DESTINATION_DIRECTORY` sets the destination path in current image configuration. Optional the same as `SOURCE_DIRECTORY_TO_IMPORT` by default.
* `after: STAGE` or `before: STAGE` specifies the stage when to import this artifact. The stage should be one of the build stages of current image configuration. Allowed options are `install` or `setup`.

Example:

```
import:
- artifact: application-assets
  add: /app/public/assets
  after: install
- artifact: application-assets
  add: /vendor
  to: /app/vendor
  after: install
```

#### `git: GIT_ARR`

`GIT_ARR` is an array, where each element is a separate git-specification:

```
git:
- GIT_SPEC
- GIT_SPEC
...
```

Example of `GIT_SPEC` to add a local git repository:

```
git:
- add: /
  to: /app
  owner: app
  group: app
  excludePaths:
  - public/assets
  - vendor
  - .helm
  stageDependencies:
    install:
    - package.json
    - Bowerfile
    - Gemfile.lock
    - app/assets/*
```

Example of `GIT_SPEC` to add a remote git repository:

```
git:
- url: https://github.com/kr/beanstalkd.git
  add: /
  to: /build
```

Example of adding multiple local and remote git repositories:

```
git:
# remote repo GIT_SPEC
- url: https://github.com/kr/beanstalkd.git
  add: /
  to: /build
# remote repo GIT_SPEC
- url: https://github.com/kubernetes/helm
  add: /
  to: /helm
# local repo GIT_SPEC
- add: /
  to: /app
  excludePaths:
  - app2
# local repo GIT_SPEC
- add: /app2
  to: /app2
```

### Shell

Shell builder instructions are given as follows:

```
shell:
  STAGE:
  - INSTRUCTION
  - INSTRUCTION
  ...
  STAGE:
  - INSTRUCTION
  - INSTRUCTION
  ...
```

* `STAGE` is one of: `beforeInstall`, `install`, `beforeSetup`, `setup`.
* `INSTRUCTION` is a free form bash expression.

Example:

```
shell:
  beforeInstall:
    - useradd -d /app -u 7000 -s /bin/bash app
    - rm -rf /usr/share/doc/* /usr/share/man/*
    - apt-get update
    - apt-get -y install apt-transport-https git curl gettext-base locales tzdata
  setup:
    - locale-gen en_US.UTF-8
```

### Ansible

Ansible builder instructions are given as follows:

```
ansible:
  STAGE:
  - ANSIBLE_TASK
  - ANSIBLE_TASK
  ...
  STAGE:
  - ANSIBLE_TASK
  - ANSIBLE_TASK
  ...
```

* `STAGE` is one of: `beforeInstall`, `install`, `beforeSetup`, `setup`.
* Each `ANSIBLE_TASK` is an element from ansible `tasks` array, see https://docs.ansible.com/ansible/latest/playbooks_intro.html

List of supported ansible modules:

* Commands Modules (https://docs.ansible.com/ansible/latest/list_of_commands_modules.html)
    * command
    * shell
    * raw
    * script
* Files Modules (https://docs.ansible.com/ansible/latest/list_of_files_modules.html)
    * assemble
    * archive
    * unarchive
    * blockinfile
    * lineinfile
    * file
    * find
    * tempfile
    * copy
    * acl
    * xattr
    * ini_file
    * iso_extract
* Net Tools Modules (https://docs.ansible.com/ansible/latest/list_of_net_tools_modules.html)
    * get_url
    * slurp
    * uri
* Packaging Modules (https://docs.ansible.com/ansible/latest/list_of_packaging_modules.html)
    * apk
    * apt
    * apt_key
    * apt_repository
    * yum
    * yum_repository
* System Modules (https://docs.ansible.com/ansible/latest/list_of_system_modules.html)
    * user
    * group
    * getent
    * locale_gen
* Utilities Modules (https://docs.ansible.com/ansible/latest/list_of_utilities_modules.html)
    * assert
    * debug
    * set_fact
    * wait_for
* Crypto Modules (https://docs.ansible.com/ansible/latest/list_of_crypto_modules.html)
    * openssl_certificate
    * openssl_csr
    * openssl_privatekey
    * openssl_publickey

Example:

```
ansible:
  beforeInstall:
  - name: "Create non-root main application user"
    user:
      name: app
      comment: "Non-root main application user"
      uid: 7000
      shell: /bin/bash
      home: /app
  - name: Disable docs and man files installation in dpkg
    copy:
      content: |
        path-exclude=/usr/share/man/*
        path-exclude=/usr/share/doc/*
      dest: /etc/dpkg/dpkg.cfg.d/01_nodoc
  install:
  - name: "Precompile assets"
    shell: |
      set -e
      export RAILS_ENV=production
      source /etc/profile.d/rvm.sh
      cd /app
      bundle exec rake assets:precompile
    args:
      executable: /bin/bash
```
