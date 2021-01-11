---
title: Giterminism
sidebar: documentation
permalink: documentation/advanced/configuration/giterminism.html
change_canonical: true
---

Since v1.2 version werf introduces so called _giterminism mode_ by default (this word constructed from `git` and `determinism`, which means "determined by the git").

Werf will take all configuration files from the current git commit of the local git repository in the giterminism mode. It is also forbidden to use some configuration or templating directives (such as `mount` or `env`).  However it is possible to loosen giterminism restrictions explicitly with [`werf-giterminism.yaml` configuration file](#werf-giterminismyaml).

## `werf-giterminism.yaml`

`werf-giterminism.yaml` configuration file describes which restrictions should be ignored for current configuration (allow usage of some uncommitted files, mount directives, environment variables etc.).

**Important**. This file should be committed into the project git repository to take effect. 

**Important**. It is not recommended loosen giterminism restrictions, because it increases probability to construct a configuration, which will lead to unreproducible builds and deploys of your application. It is important to minimize usage of `werf-giterminism.yaml` to construct a configuration, which is GitOps friendly, durable and easily reproducible.

```yaml
giterminismConfigVersion: 1
config:  # giterminism configuration for werf.yaml
  allowUncommitted: true
  allowUncommittedTemplates:
    - /**/*/
    - .werf/template.tmpl
  goTemplateRendering:
    allowEnvVariables:                        # {{ env "VARIABLE_X" }}
      - /CI_*/
      - VARIABLE_X
    allowUncommittedFiles:                    # {{ .Files.Get|Glob|... "PATH1" }}
      - /**/*/
      - .werf/nginx.conf
  stapel:
    allowFromLatest: true
    git:
      allowBranch: true
    mount:
      allowBuildDir: true                     # from: build_dir
      allowFromPaths:                         # fromPath: PATH
        - PATH1
        - PATH2
  dockerfile:
    allowUncommitted:
      - /**/*/
      - myapp/Dockerfile
    allowUncommittedDockerignoreFiles:
      - /**/*/
      - myapp/.dockerignore
    allowContextAddFile:
      - aaa
      - bbb
helm: # giterminism configuration for helm
  allowUncommittedFiles:
    - /templates/**/*/
    - values.yaml
    - Chart.yaml
```

### .Files.Get function

In the giterminism mode this function will get file content only from the current git commit.

It is possible to specify list of files which will be read from the current project working tree instead of git commit by using [`config.goTemplateRendering.allowUncommittedFiles`](#werf-giterminismyaml) `werf-giterminism.yaml` configuration file directive (globs are supported). 

### Env go-templates function

[`{{ env }}` and `{{ expandenv }}`]({{ "documentation/advanced/configuration/supported_go_templates.html" | relative_url }}) functions are only available when [`config.goTemplateRendering.allowEnvVariables`](#werf-giterminismyaml) `werf-giterminism.yaml` configuration file directive has been specified (globs are supported).

### Mount directive

[`mount` directive]({{ "documentation/reference/werf_yaml.html" | relative_url}}) of the stapel builder is only available when [`config.stapel.mount`](#werf-giterminismyaml) `werf-giterminism.yaml` configuration file directives has been specified (depending of the type of mount).

## Dockerfile builder

Werf pass build context, `Dockerfile` and `.dockerignore` to the dockerfile builder only from the local git repo commit.

There is only one implicit way to add files from the project working tree to the dockerfile build context: using [`contextAddFile` directive]({{ "documentation/reference/werf_yaml.html" | relative_url}}) directive of `werf.yaml` configuration file:

```yaml
# werf.yaml configuration file
context: app
dockerfile: Dockerfile
contextAddFile:
 - myfile
 - dir/a.out
```

In this configuration werf will create dockerfile builder context which will consists of:
 - directory `app` from the current commit of the local git repo (specified by `context` directive);
 - `myfile` and `dir/a.out` files from the `app` directory of the current project work tree (including uncommitted and untracked by git files).

[`config.dockerfile.allowContextAddFile`](#werf-giterminismyaml) directive  of the `werf-giterminism.yaml` configuration file should be specified to allow usage of `contextAddFile` directive of `werf.yaml` configuration file:

```yaml
# werf-giterminism.yaml configuration file
giterminismConfigVersion: 1
config:
  dockerfile:
    allowContextAddFile:
      - myfile
      - dir/a.out
```
