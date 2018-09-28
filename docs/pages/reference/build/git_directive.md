---
title: Adding source code from git repositories
sidebar: reference
permalink: reference/build/git_directive.html
---

## Subject matter

* Первая проблема, с которой предстоит столкнуться: как добиться минимальной задержки между коммитом и получением готового образа для дальнейшего тестирования/деплоя. Большая часть коммитов в репозиторий приложения относится к обновлению кода самого самого приложения. В этом случае сборка нового образа должна представлять собой не более, чем применение патча к файлам предыдущего образа.
* Вторая проблема: если после сделанных изменений в исходном коде приложения и сборки образа эти изменения были отменены (например, через git revert) в следующем коммите — должен сработать кэш сборщика образов. Т.е. собираемые сборщиком образа должны кэшироваться по содержимому изменений в файлах git репозитория, а не по факту наличия этих изменений.
* Третья проблема: добавлять файлы из git репозитория в образ путем копирования всего дерева исходников необходимо выполнять редко, основным способом добавления изменения должно служить наложение патчей.

Требуется собрать образ с зашитым исходным кодом приложения из git.

## Our decision

Для добавления кода в собираемый образ, предусмотрена директива `git`, с помощью которой можно добавить код из локального или удаленного репозитория (включая сабмодули) в dimg.

The `git` directive in `dappfile.yml` adds source code from a Git repository to an dimg. It supports local and remote repositories, including submodules.

> Note: to work with remote git repositories, dapp requires libssh2 and/or libssl installed.
> Refer to [Dapp Installation]({{ site.baseurl }}/how_to/installation.html#install-dependencies) for details.

При первой сборке образа с указанием директивы `git`, в него добавляется содержимое git репозитория (стадия `g_a_archive`) согласно соответствующих инструкций. При последующих сборках образа, изменения в git репозитории добавляются отдельным docker-слоем, который содержит git-патч (git patch apply). Содержимое таких docker-слоев с патчами также кешируется, что еще более повышает скорость сборки. В случае отмены сделанных изменений в исходном коде приложения (например, через git revert), при сборке будет накладываться патч с отменой изменений, будет использоваться слой из кеша.

When dapp first builds an image from a dappfile with `git` directive, it adds source code from a Git repository to the image.
This happens on the `g_a_archive` stage. (See [build stages]({{ site.baseurl }}/reference/stages/git_stages.html) for more details.)

On each subsequent build dapp does not create a new image with a full copy of the source code.
Instead, it generates a git patch (with `git patch apply`) and applies it as an image layer.
Dapp caches these image layers to boost build speed.
If changes in source code are undone, for example with `git revert`, dapp detects that and reuses a cached layer.

### g_a_archive
### g_a_post_setup_patch

Сигнатура стадии git artifact post setup patch зависит от размера патчей git-артефактов и будет пересобрана, если их сумма превысит лимит (10 MB).

### g_a_latest_patch


## Syntax


## Introduction

### Optimizing build speed


### Building artifacts

> TODO: Переместить вперёд. Слишком рано для этой инфы, даже про `stageDependencies` ещё не было.

Система кэширования dapp принимает решение о необходимости повторной сборки стадии или использовании кэша, на основании вычисления [сигнатуры стадии]({{ site.baseurl }}/not_used/stages_diagram.html), которая не зависит напрямую от состояния git репозитория. Т.о., если не указать это явно (см далее про директиву `stageDependencies`), то изменения только кода в git репозитории не приведут к повторной сборке пользовательской стадии (`before_install`, `install`, `before_setup`, `setup`, `build_artifact`). Чтобы явно определить зависимость от файлов и папок, при изменении которых сборщику необходимо выполнить принудительную сборку определенных пользовательских стадий, в директиве `git` предусмотрена директива `stageDependencies` (`stage_dependencies` для Ruby синтаксиса).

> Q: а когда сигнатура всё-таки меняется? От чего зависит? По stages_diagram не очень понятно.

Dapp decides whether to rebuild a stage or use a cached result by calculating the [stage signature]({{ site.baseurl }}/not_used/stages_diagram.html).
This signature has no direct connection to the state of the git repository.
Changing the code in the git repository does not result in rebuilding user stages
(`before_install`, `install`, `before_setup`, `setup`, and `build_artifact`).
An exception is when user stages are dependent on files, listed in `stageDependencies` directive.
This will be explained further.

Правильная установка зависимостей - важное условие построения эффективного процесса сборки!

Количество указаний директивы `git` в описании образа не ограничено, но нужно стремиться к их уменьшению, путем правильного использования `includePaths` и `excludePaths`.

Correctly installing dependencies is crucial to making an effective build process.


> TODO это дублируется в списке дальше

### General Features 

> Общие особенности использования

* описание сборки образа (`dimg` или `artifact`) может содержать любое количество git-директив;
* изменение кода в git репозитории который используется при сборке **образа приложения**, накладывается патчем и не ведет к пересборке какой-либо пользовательской стадии (если нет явного указания зависимости через `stageDependencies`);
* изменение кода в git репозитории который используется при сборке **образа артефакта**, не ведет к пересборке образа артефакта и не накладывается патчем (если нет явного указания зависимости через `stageDependencies`);
* для пересборки пользовательской стадии в зависимости от изменений в git репозитории нужно описывать зависимости с использованием `stageDependencies`;

* при использовании git submodule-й, логика не меняется - инструкции описываются так же, как в случае с директориями;
* для исключения избыточного копирования кода в образ, в директиве `git` предусмотрены параметры `includePaths` и `excludePaths`;
* важно помнить, что код добавленный с помощью директивы `git`, еще не доступен на пользовательской стадии `before install` (см. [подробней]({{ site.baseurl }}/not_used/stages_arhitecture.html) про стадии сборки).

* Destination paths, defined with `to`, should be different between artifact images in one dappfile. 
* Description of an image (either `dimg` or `artifact`) may have any number of git directives.

> TODO: уточнить и написать просто.

* When building an application image, changed code in the git repository is added as an image layer with a patch.
* When building an artifact image, changing code in git repository does not.

* With git submodule the logic is the same.


## Using `git` Directive

> YAML синтаксис (dappfile.yml)

В dappfile.yml директива `git` применяется следующим образом:

The `git` directive adds source code from git repositories.
It is an array of one or more elements, each specifying a single repository:

```yaml
git:
- GIT_SPEC
- GIT_SPEC
  ...
```

, где `GIT_SPEC` - один или несколько массивов описаний директив добавления кода следующего вида:
- для работы с локальным репозиторием

In this example all code from the local repository will be added to the `/app` directory:

```yaml
git:
- to: /app

```

Minimal specification for a remote repository:

```yaml
git:
- url: https://github.com/kr/beanstalkd.git
  add: /
  to: /build
```

Описание директив:

## Basic Directives

### `to`

```yaml
to: <to_absolute_path>
```

Sets the *destination path* — an absolute path in the described image to copy files and directories to.
Destination paths should not overlap within an image.

`to` is a required directive.

* `to: <to_absolute_path>` -  определяет путь назначения, при копировании файлов из репозитория, где `<to_absolute_path>` - абсолютный путь, в который будут копироваться ресурсы.
* пути добавления не должны пересекаться между артефактами;

### `add`


```yaml
add: <add_absolute_path>
```

Sets the *source path* — an absolute path in the source repository.
Dapp will copy from this path recursively: with all subdirectories and files.

`add` defaults to `/` and can be omitted.

If you need to copy selected directories or files from a repository, consider using `includePaths` and `excludePaths` directives.

* `add: <add_absolute_path>` - определяет путь - источник репозитория, где `<add_absolute_path>` - путь относительно репозитория, из которого будут копироваться ресурсы.

## Advanced Directives

### `owner`

* `owner: <owner>` - определяет пользователя владельца, который будет установлен ресурсам после их копирования.

```yaml
owner: <owner>
```

### `group`

* `group: <group` - определяет группу владельца, которая будет установлена ресурсам после их копирования.

```yaml
group: <group>
```

### `include_paths` and `exclude_paths`

```yaml
includePaths:
- <path> 
- <path> 
...
excludePaths:
- <path> 
- <path> 
...
```

The number of repository specifications in a `git` directive is not limited.
Indeed, `includePaths` and `excludePaths` can help reduce this number.

For example, this code:

```yaml
git:
- add: /a
  to: /app/a
- add: /b 
  to: /app/b
```

can be reduced with `includePaths`:

```yaml
git:
- add: /
  to: /app
  includePaths:
  - a
  - b
```


* `include_paths: <relative_path_or_mask>` - определяет относительные пути или маски ресурсов которые и только которые будут скопированы.
* `exclude_paths: <relative_path_or_mask>` - определяет относительные пути или маски ресурсов которые необходимо игнорировать при копировании.


### `stageDependencies`

Declares that a user stage

* `stageDependencies: ` - определяет зависимость пользовательской стадии (`install`, `beforeSetup`, `setup` - для любого типа образов, `buildArtifact` - только для сборки образа артефактов) от файлов и папок, при изменении которых необходимо выполнить принудительную сборку пользовательской стадии. Файлы и папки определяются относительным путем или маской. Учитывается как содержимое так и имена файлов/папок.

### Using Wildcards in Paths

Wildcards can be used in `includePaths`, `excludePaths` and `stageDependencies` directives.

Strictly speaking, wildcards in dapp work like [fnmatch](http://man7.org/linux/man-pages/man3/fnmatch.3.html)
with `FNM_PATHNAME` and `FNM_PERIOD` flags.
If that does not explain it, here are the details.

* `*` is any number of characters.
* `?` is any single character.
* `[abc]` is one of the given characters.
* `**` means any number of nested directories.
* Wildcard paths are case-sensitive.

---

* Directories will be ignored. TODO: wut?
* TODO: about adding `**/*` to any path.


Правила указания масок:

* поддерживаются glob-паттерны
* пути в <glob> указываются относительные
* директории игнорируются
* маски чувствительны к регистру.

Here are some examples:

```yaml
git:
  includePaths:
  - /app/*.py
  - /otherapp/*.py
  excludePaths:
  - /**/test/*.py
  stageDependencies:
    install:
    - *.py
```


## Working with Remote Repositories

### `url`

```yaml
url: <git_repo_url>
```

`url` sets the URL of a remote repository to copy files from.
It is the only required directive for copying from remote repositories.

Dapp can clone repositories over SSH or HTTPS:

```yaml
git:
- url: https://github.com/example/https.git
  to: /example-https
- url: git@github.com:example/ssh.git
  to: /example-ssh
```
> Note: working with remote repositories over HTTPS or SSL requires [installing libssl or libssh2]({{ site.baseurl }}/how_to/installation.html#install-dependencies) respectively.

To make SSH connections, dapp will use ssh-agents in the following order, most to least preferred:

* Run a temporary ssh-agent with the key, provided with `--ssh-key <path_to_key>` parameter in the command line.
* Use the system ssh-agent if it is available and running.
* Run a temporary ssh-agent with key at `~/.ssh/id_rsa`

---

* `url: <git_repo_url>` - определяет внешний git репозиторий, где `<git_repo_url>` - ssh или https адрес репозитория (в случае использования ssh адреса, ключ `--ssh-key` dapp позволяет указать ssh-ключ для доступа к репозиторию).
* поддерживается два типа git-директив, local и remote, для использования локального и удаленного репозитория соответственно;

### `branch`

```yaml
branch: <branch_name>
```

Sets the branch name in a remote repository.

`branch` is optional and defaults to `master`.

* `branch: <branch_name>` - определяет используемую ветку внешнего git репозитория, необязательный параметр (по умолчанию - master).

### `commit`

```yaml
commit: <commit>
```

Sets the commit in a remote repository

`commit` is optional and default to the last commit in selected `branch`.

* `commit: <commit>` - определяет используемый коммит внешнего git репозитория, необязательный параметр.

### `as`

* `as: <custom_name>` - назначает данному описанию git артефакта имя. Используется, например, в helm шаблонах для получения и передачи через переменные окружения в образ id коммита (обратиться можно через `.Values.global.dapp.dimg.DIMG_NAME.git.CUSTOM_NAME.commit_id` для именованного образа и `.Values.global.dapp.dimg.git.CUSTOM_NAME.commit_id` для безымянного образа).

```yaml
as: <custom_name>
```


## Code examples

```yaml
git:
- as: <custom_name>
  add: <add_absolute_path>
  to: <to_absolute_path>
  owner: <owner>
  group: <group>
  includePaths:
  - <relative_path_or_mask>
  excludePaths:
  - <relative_path_or_mask>
  stageDependencies:
    install:
    - <relative_path_or_mask>
    beforeSetup:
    - <relative_path_or_mask>
    setup:
    - <relative_path_or_mask>
```

* для работы с удаленным репозиторием

And one for a remote repository:

```yaml
git:
- url: <git_repo_url>
  branch: <branch_name>
  commit: <commit>
  as: <custom_name>
  add: <add_absolute_path>
  to: <to_absolute_path>
  owner: <owner>
  group: <group>
  includePaths:
  - <relative_path_or_mask>
  excludePaths:
  - <relative_path_or_mask>
  stageDependencies:
    install:
    - <relative_path_or_mask>
    beforeSetup:
    - <relative_path_or_mask>
    setup:
    - <relative_path_or_mask>
    build_artifact:
    - <relative_path_or_mask>
```


### Adding all source code from a remote repository

```
git:
- url: https://github.com/kr/beanstalkd.git
  add: /
  to: /build
```


### Adding a single file from a local repository

Пример добавления файла `/folder/file` из локального репозитория в папку `/destfolder` собираемого образа, с определением зависимости пересборки пользовательской стадии setup при изменении файла `/folder/file` в репозитории:

```
git:
- add: /folder/file
  to: /destfolder
  includePaths:
  - file
  stageDependencies:
    setup:
    - file
```

### Adding several directories and setting permissions

Как в предыдущем примере, только добавляется вся папка `/folder`, и зависимость определяется на изменение любого файла в исходной папке.

```
git:
- add: /folder/
  to: /destfolder
  stageDependencies:
    setup:
    - "*"
```

### Building an application image

Пример сборки приложения на nodeJS. Код приложения находится в корне локального репозитория.

```
dimg: testimage
from: node:9.11-alpine
git:
  - add: /
    to: /app
    stageDependencies:
      install:
        - package.json
        - bower.json
      beforeSetup:
        - app
        - gulpfile.js
shell:
  beforeInstall:
  - apk update
  - apk add git
  - npm install --global bower
  - npm install --global gulp
  install:
  - cd /app
  - npm install
  - bower install --allow-root
  beforeSetup:
  - cd /app
  - gulp build
docker:
  WORKDIR: "/app"
  CMD: ["gulp", "connect"]
```
