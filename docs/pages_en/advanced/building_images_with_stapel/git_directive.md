---
title: Adding source code from git repositories
permalink: advanced/building_images_with_stapel/git_directive.html
directive_summary: git
---

## What is git mapping?

***Git mapping*** describes a file or a directory in the git repository that should be added to the image at the particular path. The repository may be a local one, hosted in the directory that contains the config, or a remote one. In the latter case, the configuration of the _git mapping_ includes a repository address and version (branch, tag, or commit hash).

werf adds files from the repository to the image either by fully transferring them via git archive or by applying patches between commits. The full transfer is used to add files initially. Subsequent builds apply patches to reflect changes in a git repository. You can learn more about the algorithm behind fully transferring and applying patches in the [More details: git_archive...](#more-details-gitarchive-gitcache-gitlatestpatch) section.

The configuration of _git mappings_ supports filtering of files, and you can use a set of _git mappings_ to create virtually any file structure in the image. Also, you can specify the owner and the group of files in the _git mapping_ configuration, without the need to run `chown`.

werf has support for submodules. If it detects that files specified in the _git mapping_ configuration are present in submodules, it would act accordingly in order to change files in submodules correctly.

> All submodules of the project are bound to a specific commit. Thus, all collaborators get the same content. werf **does not initialize or update submodules**. Instead, it merely uses these bound commits

Here is an example of a _git mapping_ configuration. It adds source files from a local repository (here, `/src` is the source, and `/app` is the destination directory), and imports remote phantomjs source files to `/src/phantomjs`:

```yaml
git:
- add: /src
  to: /app
- url: https://github.com/ariya/phantomjs
  add: /
  to: /src/phantomjs
```

## Syntax of a git mapping

The _git mapping_ configuration for a local repository has the following parameters:

- `add` — path to a directory or a file whose contents must be copied into the image. The path must be specified relative to the repository root, and it is absolute (i.e., starts with `/`). This parameter is optional, contents of the entire repository are transferred by default, i.e., an empty `add` is equivalent to `add: /`;
- `to` — the path in the image to copy the contents specified with `add`;
- `owner` — the name or uid of the owner of the files to be copied;
- `group` — the name or gid of the owner’s group;
- `excludePaths` — a set of masks to exclude files or directories during recursive copying. Paths in masks must be specified relative to add;
- `includePaths` — a set of masks to include files or directories during recursive copying. Paths in masks must be specified relative to add;
- `stageDependencies` — a set of masks to monitor for changes that lead to rebuilds of the user stages. This is reviewed in detail in the [Running assembly instructions]({{ "advanced/building_images_with_stapel/assembly_instructions.html" | true_relative_url }}) reference.

The configuration of a _git mapping_ for a remote repository has some additional parameters:
- `url` — address of the remote repository;
- `branch`, `tag`, `commit` — the name of a branch, tag, or a commit hash that will be used. If these parameters are omitted, the master branch is used instead.

> By default, the use of the `branch` directive is not allowed by giterminism (read more about it [here]({{ "advanced/giterminism.html" | true_relative_url }}))

## Using git mappings

### Copying directories

The `add` parameter defines a source path in a repository. Then all files in this directory are recursively retrieved and added to the image at the `to` path. If the parameter is not set, werf uses the default path ( `/` ) instead. In other words, the entire repository will be copied. For example:

```yaml
git:
- add: /
  to: /app
```

This basic _git mapping_ configuration adds entire contents of the repository to the `/app` directory in the image.

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn btn__example1 active" onclick="openTab(event, 'btn__example1', 'tab__example1', 'git-mapping-01-source')">Git repo structure</a>
  <a href="javascript:void(0)" class="tabs__btn btn__example1" onclick="openTab(event, 'btn__example1', 'tab__example1', 'git-mapping-01-dest')">Structure of a resulting image</a>
</div>
<div id="git-mapping-01-source" class="tabs__content tab__example1 active">
  <img src="{{ "images/build/git_mapping_01.png" | true_relative_url }}" alt="git repository files tree" />
</div>
<div id="git-mapping-01-dest" class="tabs__content tab__example1">
  <img src="{{ "images/build/git_mapping_02.png" | true_relative_url }}" alt="image files tree" />
</div>

You can specify multiple _git mappings_:

```yaml
git:
- add: /src
  to: /app/src
- add: /assets
  to: /static
```

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn btn__example2 active" onclick="openTab(event, 'btn__example2', 'tab__example2', 'git-mapping-02-source')">Git repo structure</a>
  <a href="javascript:void(0)" class="tabs__btn btn__example2" onclick="openTab(event, 'btn__example2', 'tab__example2', 'git-mapping-02-dest')">Structure of a resulting image</a>
</div>
<div id="git-mapping-02-source" class="tabs__content tab__example2 active">
  <img src="{{ "images/build/git_mapping_03.png" | true_relative_url }}" alt="git repository files tree" />
</div>
<div id="git-mapping-02-dest" class="tabs__content tab__example2">
  <img src="{{ "images/build/git_mapping_04.png" | true_relative_url }}" alt="image files tree" />
</div>

It should be noted, however, that the _git mapping_ doesn't specify a directory to be transferred (similarly to `cp -r /src /app`). Instead, the `add` parameter specifies the contents of a directory that will be transferred from the repository recursively. That is, if you need to copy the contents of the `/assets` directory to the `/app/assets` directory, then you have to specify the **assets** keyword twice in the configuration or use the `includePaths` [filter](#using-filters). For example:

```yaml
git:
- add: /assets
  to: /app/assets
```

or

```yaml
git:
- add: /
  to: /app
  includePaths: assets
```

### Changing the owner

The _git mapping_ configuration provides the `owner` and `group` parameters. These are the names or numerical ids of the owner and group common to all files and directories transferred to the image.

```yaml
git:
- add: /src/index.php
  to: /app/index.php
  owner: www-data
```

![index.php owned by www-data user and group]({{ "images/build/git_mapping_05.png" | true_relative_url }})

If the `group` parameter is omitted, then the group is set to the primary group of the user.

If the `owner` or `group` value is a string, then the specified user or group must exist in the system by the moment the transfer of files is complete. Otherwise, the build would end with an error.

```yaml
git:
- add: /src/index.php
  to: /app/index.php
  owner: wwwdata
```



### Using filters

`includePaths` and `excludePaths` parameters help werf to process the file list. These are the sets of masks that you can use to include and exclude files and directories to/from the list of files to transfer to the image. The `excludePaths` filter works as follows: masks are applied to each file found in the `add` path. If there is at least one match, then the file is ignored; if no matches are found, then the file gets added to the image. `includePaths` works the opposite way: if there is at least one match, then the file gets added to the image.

_Git mapping_ configuration can contain both filters. In this case, a file is added to the image if the path matches any of `includePaths` masks and not match all `excludePaths` masks.

For example:

```yaml
git:
- add: /src
  to: /app
  includePaths:
  - '**/*.php'
  - '**/*.js'
  excludePaths:
  - '**/*-dev.*'
  - '**/*-test.*'
```

This git mapping configuration adds `.php` and `.js` files from `/src` except for files with suffixes starting with `-dev.` or `-test.`.

> The step involving the addition of a `**/*` template is here for convenience: the most common use case of a _git mapping_ with filters is to configure recursive copying for the directory. The addition of `**/*` allows you to specify the directory name only; thus, its entire contents would match the filter

Masks have the following wildcards:

- `*` — matches any file. This pattern includes `.` and excludes `/`
- `**` — matches directories recursively or files expansively
- `?` — matches exactly one character. It is equivalent to /.{1}/ in regexp
- `[set]` — matches any character within the set. It behaves exactly like character sets in regexp, including the set negation ([^a-z])
- `\` — escapes the next metacharacter

Masks that start with `*` or `**` should be escaped with quotation marks in the `werf.yaml` file:
 - `"*.rb"` — with double quotation marks
- `'**/*'` — with single quotation marks

Examples of filters:

```yaml
add: /src
to: /app
includePaths:
# match all php files residing directly in /src
- '*.php'

# match recursively all php files in /src
# (also matches *.php because '.' is included in **)
- '**/*.php'

# match all files in /src/module1 recursively
# an example of the implicit addition of **/*
- module1
```

You can use the `includePaths` filter to copy a single file without renaming it:
```yaml
git:
- add: /src
  to: /app
  includePaths: index.php
```

### Overlapping of target paths

Those who prefer to add multiple _git mappings_ need to remember that overlapping paths defined in `to` may result in the inability to add files to the image. For example:

```yaml
git:
- add: /src
  to: /app
- add: /assets
  to: /app/assets
```

When processing a config, werf calculates possible overlaps among all _git mappings_ related to `includePaths` and `excludePaths` filters. If an overlap is detected, werf tries to resolve the conflict by adding `excludePaths` into the _git mapping_ implicitly. In all other cases, the build ends with an error. However, the implicit `excludePaths` filter can have undesirable side effects, so it is better to avoid conflicts caused by overlapping paths between configured git mappings.

Here is an implicit `excludePaths` example:

```yaml
git:
- add: /src
  to: /app
  excludePaths:  # werf add this filter to resolve a conflict
  - assets       # between paths /src/assets and /assets
- add: /assets
  to: /app/assets
```

## Working with remote repositories

werf can use remote repositories as file sources. For this, you have to specify the repository address via the `url` parameter in the _git mapping_ configuration. werf supports `https` and `git+ssh` protocols.

### https

Here is the syntax for the https protocol:

{% raw %}
```yaml
git:
- url: https://[USERNAME[:PASSWORD]@]repo_host/repo_path[.git/]
```
{% endraw %}

To access the repository over `https`, you may need to enter login and password.

Here is an example of using GitLab CI variables for getting a login and password:

{% raw %}
```yaml
git:
- url: https://{{ env "CI_REGISTRY_USER" }}:{{ env "CI_JOB_TOKEN" }}@registry.gitlab.company.name/common/helper-utils.git
```
{% endraw %}

In the above example, we use the [env](http://masterminds.github.io/sprig/os.html) method from the sprig library for accessing the environment variables.

### git, ssh

werf supports accessing the repository via the git protocol. Commonly, this protocol is secured with ssh: this feature is used by GitHub, Bitbucket, GitLab, Gogs, Gitolite, etc. Generally, the repository address will look as follows:

```yaml
git:
- url: git@gitlab.company.name:project_group/project.git
```

A good understanding of the process of werf searching for access keys is required to use the remote repositories over ssh (read more below).

#### Working with ssh keys

The ssh-agent provides keys for ssh connections. It is a daemon operating via a file socket. The path to the socket is stored in the environment variable `SSH_AUTH_SOCK`. werf mounts this file socket into all _assembly containers_ and sets the environment variable `SSH_AUTH_SOCK`, i.e., connection to remote git repositories is established using keys registered in the running ssh-agent.

werf applies the following algorithm for using the ssh-agent:

- If werf is started with the `--ssh-key` flag (there might be multiple flags):
  - A temporary ssh-agent starts and uses the defined keys; it is used for all git operations with remote repositories.
  - The already running ssh-agent is ignored in this case.
- No `--ssh-key` flag(s) is specified and ssh-agent is running:
  - werf uses the `SSH_AUTH_SOCK` environment variable; keys that are added to this agent are used for git operations.
- No `--ssh-key` flag(s) is specified and ssh-agent is not running:
  - If the `~/.ssh/id_rsa` file exists, werf runs the temporary ssh-agent with the key contained in the `~/.ssh/id_rsa` file.
- If none of the previous options is applicable, then the ssh-agent does not start. Thus, no keys for git operations are available and building images using remote _git mappings_ ends with an error.

## More details: gitArchive, gitCache, gitLatestPatch

Let us review the process of adding files to the resulting image in more detail. As it was stated earlier, the docker image contains multiple layers. To understand what layers werf create, let's consider the building actions based on three sample commits: `1`, `2` and `3`:

- Build of a commit No. 1. All files are added to a single layer depending on the configuration of the _git mappings_. This is done with the help of the git archive command. The resulting layer corresponds to the _gitArchive_ stage.
- Build of a commit No. 2. Another layer is added. In it, files are modified by applying a patch. This layer corresponds to the _gitLatestPatch_ stage.
- Build of a commit No. 3. Files have been added already, and werf applies patches in the _gitLatestPatch_ stage layer.

The build sequence for these commits may be represented as follows:

| | gitArchive | gitLatestPatch |
|---|:---:|:---:|
| Commit No. 1 is made, build at 10:00 |  files as in commit No. 1 | - |
| Commit No. 2 is made, build at 10:05 |  files as in commit No. 1 | files as in commit No. 2 |
| Commit No. 3 is made, build at 10:15 |  files as in commit No. 1 | files as in commit No. 3 |

With time, the number of commits grows, and the size of the patch between commit No. 1 and the current one may become quite large. It will further increase the size of the latest layer and the total size of stages. To prevent the uncontrolled growth of the latest layer, werf provides the additional intermediary stage — _gitCache_. When _gitLatestPatch_ diff becomes too big, much of its diff is merged with the _gitCache_ diff, thus reducing the _gitLatestPatch_ stage size.

### _git stages_ and rebasing

Each _git stage_ stores service labels containing SHA commits that this _stage_ was built up on.
werf will use them for creating patches when assembling the next _git stage_ (in a nutshell, it is a `git diff COMMIT_FROM_PREVIOUS_GIT_STAGE LATEST_COMMIT` for each described _git mapping_).
So, if some stage has a saved commit that is not in a git repository (e.g., after rebasing), then werf would rebuild that stage at the next build using the latest commits.
