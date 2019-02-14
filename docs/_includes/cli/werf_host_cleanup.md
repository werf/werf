{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Cleanup old unused werf cache and data of all projects on host machine.

The data include:
* Lost docker containers and images from interrupted builds.
* Old service tmp dirs, which werf creates during every build, publish, deploy and other commands.
* Local cache:
  * Remote git clones cache.
  * Git worktree cache.

It is safe to run this command periodically by automated cleanup job in parallel with other werf 
commands such as build, deploy, stages and images cleanup.

{{ header }} Syntax

```bash
werf host cleanup [options]
```

{{ header }} Options

```bash
      --docker-config='':
            Specify docker config directory path. Default WERF_DOCKER_CONFIG or DOCKER_CONFIG or 
            ~/.docker (in the order of priority).
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for cleanup
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default WERF_HOME environment or 
            ~/.werf)
      --insecure-repo=false:
            Allow usage of insecure docker repos
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default WERF_TMP environment or system 
            tmp dir)
```

