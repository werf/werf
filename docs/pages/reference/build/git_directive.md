---
title: Adding source code from git repositories
sidebar: reference
permalink: reference/build/git_directive.html
summary: |
  <a href="https://docs.google.com/drawings/d/e/2PACX-1vSJax_QDnGEkIwiNndib4Yu4sKxX4sMntjb33YfOYXvb3E3wqLRxII09gvNyuxDw5gYoAoLG24gmM_I/pub?w=2031&amp;h=144" data-featherlight="image">
      <img src="https://docs.google.com/drawings/d/e/2PACX-1vSJax_QDnGEkIwiNndib4Yu4sKxX4sMntjb33YfOYXvb3E3wqLRxII09gvNyuxDw5gYoAoLG24gmM_I/pub?w=1016&amp;h=72">
  </a>
      
  <div class="tab">
    <button class="tablinks active" onclick="openTab(event, 'local')">Local</button>
    <button class="tablinks" onclick="openTab(event, 'remote')">Remote</button>
  </div>
  
  <div id="local" class="tabcontent active">
    <div class="language-yaml highlighter-rouge"><pre class="highlight"><code><span class="s">git</span><span class="pi">:</span>
  <span class="pi">-</span> <span class="s">as</span><span class="pi">:</span> <span class="s">&lt;custom name&gt;</span>
    <span class="s">add</span><span class="pi">:</span> <span class="s">&lt;absolute path&gt;</span>
    <span class="s">to</span><span class="pi">:</span> <span class="s">&lt;absolute path&gt;</span>
    <span class="s">owner</span><span class="pi">:</span> <span class="s">&lt;owner&gt;</span>
    <span class="s">group</span><span class="pi">:</span> <span class="s">&lt;group&gt;</span>
    <span class="s">includePaths</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;relative path or glob&gt;</span>
    <span class="s">excludePaths</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;relative path or glob&gt;</span>
    <span class="s">stageDependencies</span><span class="pi">:</span>
      <span class="s">install</span><span class="pi">:</span>
      <span class="pi">-</span> <span class="s">&lt;relative path or glob&gt;</span>
      <span class="s">beforeSetup</span><span class="pi">:</span>
      <span class="pi">-</span> <span class="s">&lt;relative path or glob&gt;</span>
      <span class="s">setup</span><span class="pi">:</span>
      <span class="pi">-</span> <span class="s">&lt;relative path or glob&gt;</span></code></pre>
    </div>
  </div>
  
  <div id="remote" class="tabcontent">
    <div class="language-yaml highlighter-rouge"><pre class="highlight"><code><span class="s">git</span><span class="pi">:</span>
  <span class="pi">-</span> <span class="s">url</span><span class="pi">:</span> <span class="s">&lt;git repo url&gt;</span>
    <span class="s">branch</span><span class="pi">:</span> <span class="s">&lt;branch name&gt;</span>
    <span class="s">commit</span><span class="pi">:</span> <span class="s">&lt;commit&gt;</span>
    <span class="s">tag</span><span class="pi">:</span> <span class="s">&lt;tag&gt;</span>
    <span class="s">as</span><span class="pi">:</span> <span class="s">&lt;custom name&gt;</span>
    <span class="s">add</span><span class="pi">:</span> <span class="s">&lt;absolute path&gt;</span>
    <span class="s">to</span><span class="pi">:</span> <span class="s">&lt;absolute path&gt;</span>
    <span class="s">owner</span><span class="pi">:</span> <span class="s">&lt;owner&gt;</span>
    <span class="s">group</span><span class="pi">:</span> <span class="s">&lt;group&gt;</span>
    <span class="s">includePaths</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;relative path or glob&gt;</span>
    <span class="s">excludePaths</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;relative path or glob&gt;</span>
    <span class="s">stageDependencies</span><span class="pi">:</span>
      <span class="s">install</span><span class="pi">:</span>
      <span class="pi">-</span> <span class="s">&lt;relative path or glob&gt;</span>
      <span class="s">beforeSetup</span><span class="pi">:</span>
      <span class="pi">-</span> <span class="s">&lt;relative path or glob&gt;</span>
      <span class="s">setup</span><span class="pi">:</span>
      <span class="pi">-</span> <span class="s">&lt;relative path or glob&gt;</span></code></pre>
      </div>
  </div>
---

## What is git path? 

_Git path_ describes a file or directory from the git repository that should be added to the image by a certain path. The repository may be a local one, hosted in the directory that contains the dappfile, or a remote one, and in this case, the configuration of the git path contains the repository address and the version (branch, tag or commit hash).

dapp adds the files from the repository to the image by using full transfer of files with git archive or by applying patches betweeb commits.
The full transfer is used for initial adding of files. The subsequent builds use applying patches to reflect changes in git repository. The algorithm behind the full transfer and applying patches is reviewed the "[More details: g_a_archive, g_a_post_setup_patch, g_a_latest_patch](#more-details-g_a_archive-g_a_post_setup_patch-g_a_latest_patch)" section.

The configuration of the git path supports filtering files, and you can use the set of git paths to create virtually any resulting file structure in the final image. In addition, you can specify the owner and the group of files in the git path configuration — no subsequent `chown` required.

dapp has support for submodules. dapp detects if files specified with git path configuration are contained in submodules and does the very best it could to correctly handle the changes of files in submodules.

An example of a git path configuration for adding source files from local repository from the /src into the /app directory, and remote phantomjs source files to /src/phantomjs:

```
dimg: backend-app
git:
- add: /src
  to: /app
- url: https://github.com/ariya/phantomjs
  add: /
  to: /src/phantomjs
```

## Motivation for git paths

Main idea is to bring git history into build process.

### Patching instead of copying

Most commits in the real application repository relate to updating the code of the application itself. In this case, if compilation is not required, assembling a new image shall be nothing more than applying patches to the files in the previous image.

### Remote repositories

Building an application image may depend on source files in other repositories. dapp provides ability of adding files from remote repositories into the final image. dapp can detect changes in local repositories and in remote repositories.

## Syntax of a git path

The git path configuration for a local repository has the following parameters:

- `add` — the path to a directory or file whose contents must be copied to the image. The path is specified relative to the repository root, and the path is absolute (i.e. it must start with `/`). This parameter is optional, the content of the entire repository is transferred by default, i.e. an emtpy `add` is equal to `add: /`;
- `to` — the path in the image, where the content specified with `add` will be copied;
- `owner` — the name or uid of the owner of the copied files;
- `group` — the name or gid of the group of the owner;
- `excludePaths` — a set of masks to ignore the files or directories during recursive copying. Paths in masks are specified relative to add;
- `includePaths` — a set of masks to include the files or directories during recursive copying. Paths in masks are specified relative to add;
- `stageDependencies` — a set of masks to detect changes that lead to the user stages rebuilds. This is reviewed in detail in the [User stages and assembly instructions]({{ site.baseurl }}/reference/build/assembly_instructions.html) reference.

The git path configuration for a remote repository has a number of additional parameters:
- `url` — remote repository address;
- `branch`, `tag`, `commit` — name of branch, tag or commit hash that will be used. If these parameters are not specified, the master branch is used;
- `as` — defines an alias to simplify the retrieval of remote repository-related information in helm templates. Details are available in the [Deployment to kubernetes]({{ site.baseurl }}/reference/deploy/deploy_to_kubernetes.html) reference.

## Uses of git paths

### Copying of directories

The `add` parameter specifies a path in a repository, from which all files must be recursively retrieved and added to the image with the `to` path; if the parameter is not specified, then the default path — `/` is used, i.e. the entire repository is transferred.
For example:

```yaml
---
dimg: frontend
git:
- add: /
  to: /app
```

This is the simple git path configuration that adds the entire content from the repository to the /app directory in the image.

If the repository contains the following structure:

image1

Then the final image will contain this structure:

image2

Multiple git paths may be specified:

```yaml
---
dimg: frontend
git:
- add: /src
  to: /app/src
- add: /assets
  to: /static
```

If the repository contains the following structure:

image3

Then the final image will contain this structure:

image4

It should be noted, that git path configuration doesn't specify a directory to be transferred like `cp -r /src /app`. `add` parameter specifies a directory content that will be recursively transferred from the repository. That is, if the `/assets` directory needs to be transferred to the `/app/assets` directory, then the name 'assets' should be written twice, or `includePaths` [filter](#using-filters) can be used.


```yaml
- add: /assets
  to: /app/assets
or
- add: /
  to: /app
  includePaths: assets
```

_Note: dapp has no convention for trailing / that is available in rsync, i.e. `add: /src` and `add: /src/` are the same._


### Copying of file

Copying the content, not the specified directory, from `add` path also applies to files. To transfer the file to the image, you must specify its name twice — once in `add`, and again in `to`. This provides an ability to rename the file:

```yaml
---
dimg: frontend
git:
- add: /config/prod.yaml
  to: /app/conf/production.yaml
```

### Changing an owner

The git path configuration provides parameters `owner` and `group`. These are the names or numerical ids of the owner and group used for all files and directories transferred to the image.

```yaml
---
dimg: frontend
from: ubuntu:18.04
git:
- add: /src/index.php
  to: /app/index.php
  owner: www-data
```

image_here

If only the `owner` parameter is specified, the group for files will be the same as the primary group of the specified user.

If `owner` or `group` value is a string, then the specified user or group must be added to the system by the time of the full transfer of files, otherwise build will end with an error.

```yaml
---
dimg: frontend
from: ubuntu:18.04
git:
- add: /src/index.php
  to: /app/index.php
  owner: wwwdata
```

image_here


### Using filters

`includePaths` and `excludePaths` parameters are used when processing the file list. These are the sets of masks that can be used to include and exclude files and directories from/to the list of files that will be transferred to the image. Simply stated, the `excludePaths` filter works as follows: masks are applied to each file found in `add` path. If at least one mask matches, then the file is ignored; if no matches are found, then the file gets added to the image. `includePaths` works the opposite way: if at least one mask is a match, the file gets added to the image.

Git path configuration can contain both filters. In this case file is added to the image if path match with one of `includePaths` masks and not match with all `excludePaths` masks.

For example:

```yaml
dimg: frontend
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

This is the git path configuration that adds `.php` and `.js` files from `/src` except files with `-dev` or `-test` suffixes.

To determine whether the file matches the mask, the following algorithm is applied:
- the path in `add` is concatenated with the mask;
- an absolute file path inside the repository is taken;
- two paths are compared with the use of [fnmatch](https://ruby-doc.org/core-2.2.0/File.html#method-c-fnmatch) with FNM_PATHNAME and FNM_PERIOD flags (`.` is included in the `*`, however `/` is excluded);
- if fnmatch returns true, then the file is matched, and the algorithm is ended;
- the path in `add` is concatenated with the mask and with an additional pattern `**/*` ;
- an absolute file path inside the repository is taken;
- two paths are compared with the use of fnmatch with FNM_PATHNAME and FNM_DOTMATCH flags (`.` is included in the `*`, however `/` is excluded);
- if fnmatch returns true, then the file is matched; if false, the file does not match;

Note: the second step with adding `**/*` template is for convenience: the most frequent use case of a git path with filters is to configure recursive copying for the directory. Adding `**/*` makes enough to specify the directory name only, and its entire content will match the filter.

Masks may contain the following patterns:

- `*` — matches any file. This pattern include `.` and exclude `/`
- `**` — matches directories recursively or files expansively
- `?` — matches any one character. Equivalent to /.{1}/ in regexp
- `[set]` — matches any one character in set. Behaves exactly like character sets in regexp, including set negation ([^a-z])
- `\` — escapes the next metacharacter

Mask that starts with `*` or `**` patterns should be escaped with quotes in dappfile.yaml file:
 - `- "*.rb"` — with double quotes
- `- '**/*'` — with single quotes

Examples of filters:

```yaml
add: /src
to: /app
includePaths:
# match all php files residing directly in /src
- '*.php' 

# matches recursively all php files from /src
# (also matches *.php because '.' is included in **)
- '**/*.php'

# matches all php files from /src/module1, that reside
# directly in config, view or controller subdirectories
- module1/**/{config,view,contoller}/*.php

# matches all files from /src/module1 recursively
# an example of implicit adding of **/*
- module1
```

`includePaths` filter can be used to copy one file without renaming:
```yaml
git:
- add: /src
  to: /app
  includePaths: index.php
```


### Note: target paths intersecting

If multiple git paths are added, you should remember that intersecting paths defined in `to` may result in the inability to add files to the image. For example:

```yaml
---
dimg: frontend
git:
- add: /src
  to: /app
- add: /assets
  to: /app
```

When processing a dappfile, dapp calculates the possible intersections among all git paths with respect to includePaths and excludePaths filters. If an intersection is detected, then dapp can resolve simple conflicts with implicitly adding `excludePaths` into the git path. In other cases the build will end with an error. However, implicit `excludePaths` filter can have undesirable effects, so try to avoid conflicts of intersecting paths between configured git paths.

Implicit excludePaths example:

```yaml
dimg: frontend
git:
- add: /src
  to: /app
  excludePaths:  # dapp add this filter to resolve a conflict
  - assets       # between paths /src/assets and /assets
- add: /assets
  to: /app/assets
```

## Working with remote repositories

dapp may use remote repositories as file sources. For this purpose, the git path configuration contains an `url` parameter where you should specify the repository address. dapp supports `https` and `git+ssh` protocols.

### https

The syntax for https protocol is:

{% raw %}
```yaml
git:
- url: https://[USERNAME[:PASSWORD]@]repo_host/repo_path[.git/]
```
{% endraw %}

`https` access may require login and password.

For example, login and password from gitlab ci variables:

{% raw %}
```yaml
git:
- add: https://{{ env "CI_REGISTRY_USER" }}:{{ env "CI_JOB_TOKEN" }}@registry.gitlab.company.name/common/helper-utils.git
```
{% endraw %}

In this example, the [env method](http://masterminds.github.io/sprig/os.html) from the sprig library is used to access the environment variables.


### git, ssh

dapp supports access to the repository via the git protocol. Access via this protocol is typically protected using ssh tools: this feature is used by github, bitbucket, gitlab, gogs, gitolite, etc. Most often the repository address looks as follows:

```yaml
git:
- add: git@gitlab.company.name:project_group/project.git
```

To successfully work with remote repositories via ssh, you should understand how dapp searches for access keys.


#### Working with ssh keys

Keys for ssh connects are provided by ssh-agent. The ssh-agent is a daemon that operates via file socket, the path to which is stored in the environment variable SSH_AUTH_SOCK. Dapp mounts this file socket to all assembly containers and sets the environment variable SSH_AUTH_SOCK, i.e. connection to remote git repositories is established with the use of keys that are registered in the running ssh-agent.


The ssh-agent is determined as follows:

- If dapp is started with `--ssh-key` flags (there may be multiple flags):
  - A temporary ssh-agent runs with defined keys, and it is used for all git operations with remote repositories
  - The already running ssh-agent will be ignored in this case
- No `--ssh-key` flags specified and ssh-agent is running:
  - `SSH_AUTH_SOCK` environment variable will be used, and the keys added to this agent will be used for git operations
- No `--ssh-key` flags specified and ssh-agent is not running:
  - If `~/.ssh/id_rsa` file exists, then dapp will run the temporary ssh-agent with the  key from `~/.ssh/id_rsa` file
- If none of the previous options is applicable, then the ssh-agent is not start, and no keys for git operation are available. Build images with remote git paths ends with an error.

## More details: g_a_archive, g_a_post_setup_patch, g_a_latest_patch

Let us review adding files to the resulting image in more detail. As stated earlier, the docker image contains multiple layers. To understand what layers dapp will create, let's consider the building actions based on three sample commits: `1`, `2` and `3`:

- Build of commit No. 1. All files are added to a single layer based on the configuration of the git paths. This is done with the help of the git archive. This is the layer of the `g_a_archive` stage.
- Build of commit No. 2. Another layer is added where the files are changed by applying a patch. This is the layer of the `g_a_latest_patch` stage.
- Build of commit No. 3. Files has already added, so dapp apply patches in the `g_a_latest_patch` stage layer.

Build sequence for these commits may be represented as follows:

| | g_a_archive | --- | g_a_latest_patch |
|---|---|---|---|
| Commit No. 1 is made, build at 10:00 |  files as in commit No. 1 | --- | - |
| Commit No. 2 is made, build at 10:05 |  files as in commit No. 1 | --- | files as in commit No. 2 |
| Commit No. 3 is made, build at 10:15 |  files as in commit No. 1 | --- | files as in commit No. 3 |

An empty space between the layers in this table is not accidental. After a while, the number of commits will grow, and the patch between commit No. 1 and the current commit may become quite large, which will further increase the size of the last layer and the total size of the final image. In order to prevent growth of the last layer, dapp provides another intermediary stage — `g_a_post_setup_patch`.
How does dapp work with these three stages? Now we are going to need more commits to illustrate this, let it be `1`, `2`, `3`, `4`, `5`, `6` and `7`.

- Build of commit No. 1. As before, files are added to a single layer based on the configuration of the git paths. This is done with the help of the git archive. This is the layer of the `g_a_archive` stage.
- Build of commit No. 2. The layer of the `g_a_post_setup_patch` stage is added, where files are changed by applying a patch between commits `1` and `2`.
- Build of commit No. 3. The layer of the `g_a_latest_patch` stage is added, where the patch between `2` and `3` is applied.
- Build of commit No. 4. The size of the patch between `1` and `4` does not exceed 1 MiB, so only the layer of the `g_a_latest_patch` stage is modified by applying the patch between `2` and `4`.
- Build of commit No. 5. The size of the patch between `1` and `5` does not exceed 1 MiB, so only the layer of the `g_a_latest_patch` stage is modified by applying the patch between `2` and `5`.
- Build of commit No. 6. The size of the patch between `1` and `6` exceeds 1 MiB. Now `g_a_post_setup_patch` stage layer is modified.
- Build of commit No. 7. The layer of the `g_a_latest_patch` stage is modified by applying the patch between `6` and `7`.

This means that as commits are added starting from the moment the first build is done, big patches are gradually accumulated into the layer for the `g_a_post_setup_patch` stage, and only patches with moderate size are applied in the layer for the last `g_a_latest_patch` stage. This algorithm reduces the size of the stages cache.

| | g_a_archive | g_a_post_setup_patch | g_a_latest_patch |
|---|---:|---:|---:|
| Commit No. 1 is made, build at 12:00 |  1 |  - | - |
| Commit No. 2 is made, build at 12:05 |  1 |  2 | - |
| Commit No. 3 is made, build at 12:15 |  1 |  2 | 3 |
| Commit No. 4 is made, build at 12:19 |  1 |  2 | 4 |
| Commit No. 5 is made, build at 12:25 |  1 |  2 | 5 |
| Commit No. 6 is made, build at 12:45 |  1 | *6 | - |
| Commit No. 7 is made, build at 12:57 |  1 |  6 | 7 |

\* — the size of the patch for commit `6` exceeded 1 MiB, so this patch is applied in the layer for the `g_a_post_setup_patch` stage.


### Rebuild of `g_a_archive` stage

For various reasons you may want to reset the `g_a_archive` stage. For example, 

 (for example, too much changes have accumulated and rebuild allows to decrease the size of the assembly cache and the resulting image). You can reset the `g_a_archive` stage specifying the "[dapp reset]" or "[reset dapp]" string in the commit's message. Let us assume that, in the previous example commit `4` contains '[dapp reset]' in its message, then the builds would look as follows:


| | g_a_archive | g_a_post_setup_patch | g_a_latest_patch |
|---|---:|---:|---:|
| Commit No. 1 is made, build at 12:00 |  1 |  - | - |
| Commit No. 2 is made, build at 12:05 |  1 |  2 | - |
| Commit No. 3 is made, build at 12:15 |  1 |  2 | 3 |
| Commit No. 4 is made, build at 12:19 |  *4 | - | - |
| Commit No. 5 is made, build at 12:25 |  4 | 5 | - |
| Commit No. 6 is made, build at 12:45 |  4 | 5 | 6 |
| Commit No. 7 is made, build at 12:57 |  4 | 5 | 7 |

\* — commit `4` contains the "[dapp reset]" string in its message, so the `g_a_archive` stage is rebuild.


### git_archive and rebase

dapp supports git rebases. Layers for all `g_a_` stages has a label with commit sha from which this layer was build. If the commit for the stage is lost after rebase then dapp rebuild that stage.
