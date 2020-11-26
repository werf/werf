---
title: Determinism
sidebar: documentation
permalink: documentation/advanced/configuration/determinism.html
---

Since v1.2 version werf introduces so called _determinism mode_ by default.

Werf will take all configuration files from the current git commit of the local git repository in the determinism mode. It is forbidden to have uncommitted changes to these files by default — werf will exit with an error in such case. With `--non-strict-determinism-inspection` flag werf will only print warnings about uncommitted files, but use file from the current git commit anyway.

It is possible to read the following configuration files from the current project working tree with `--disable-determinism` flag:
 - `werf.yaml`;
 - `.werf/**/*.tmpl` go templates additional files;
 - `.helm/templates`;
 - `.helm/values.yaml`;
 - `.helm/secret-values.yaml`;
 - `.helm/Chart.yaml`.

**Important**. It is not recommended to disable determinism mode, because it increases probability to write a configuration, which will lead to unreproducible builds and and deploys of your application. It is important to use determinism mode to construct configuration which is GitOps friendly.
 
### .Files.Get function

In the determinism mode this function will get file content only from the current git commit.

With `--disable-determinism` flag werf will read the specified file from the current project working tree.

### Env go-templates function

[`{{ env }}` and `{{ expandenv }}`]({{ "documentation/advanced/configuration/supported_go_templates.html" | relative_url }}) functions are only available when `--disable-determinism` has been specified.

### Mount directive

[`mount` directive]({{ "documentation/reference/werf_yaml.html" | relative_url}}) of the stapel builder is only available when `--disable-determinism` has been specified. 

## Dockerfile builder

Werf pass context and Dockerfile itself to the dockerfile builder only from the local git repo commit.

**Important:** All uncommitted changes to the files in the project work tree will be ignored for the dockerfile builder context and `--disable-determinism` flag does not affect this behaviour.

However there is one implicit only way to add files from the project working tree to the dockerfile build context: using [`contextAddFile` directive]({{ "documentation/reference/werf_yaml.html" | relative_url}}):

```
context: app
dockerfile: Dockerfile
contextAddFile:
 - myfile
 - dir/a.out
```

In this configuration werf will create dockerfile builder context which will consists of:
 - directory `app` from the current commit of the local git repo (specified by `context` directive);
 - `myfile` and `dir/a.out` files from the `app` directory of the current project work tree (including uncommitted and untracked by git files).

## Summary

|             | default | `--non-strict-determinism-inspection` | `--disable-determinism` |  
| `werf.yaml` | read from the git commit, uncommitted forbidden | read from the git commit, uncommitted is warning | read from the work tree, uncommitted is ok |
| `.werf/**/*.tmpl`  | read from the git commit, uncommitted forbidden | read from the git commit, uncommitted is warning | read from the work tree, uncommitted is ok |
| `.helm/templates` | read from the git commit, uncommitted forbidden | read from the git commit, uncommitted is warning | read from the work tree, uncommitted is ok |
| `.helm/values.yaml` | read from the git commit, uncommitted forbidden | read from the git commit, uncommitted is warning | read from the work tree, uncommitted is ok |
| `.helm/secret-values.yaml` | read from the git commit, uncommitted forbidden | read from the git commit, uncommitted is warning | read from the work tree, uncommitted is ok |  
| `.helm/Chart.yaml` | read from the git commit, uncommitted forbidden | read from the git commit, uncommitted is warning | read from the work tree, uncommitted is ok |
| Dockerfiles specified in the `werf.yaml` |  read from the git commit, uncommitted forbidden | read from the git commit, uncommitted is warning | read from the git commit (cannot disable determinism) |
| `context` for dockerfiles specified in the `werf.yaml` |  read from the git commit, uncommitted forbidden | read from the git commit, uncommitted is warning | read from the git commit (cannot disable determinism) | 
