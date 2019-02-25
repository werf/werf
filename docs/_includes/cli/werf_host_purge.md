{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Purge werf images, stages, cache and other data of all projects on host machine.

The data include:
* Old service tmp dirs, which werf creates during every build, publish, deploy and other commands.
* Local cache:
  * Remote git clones cache.
  * Git worktree cache.
* Shared context:
  * Mounts which persists between several builds (mounts from build_dir).

WARNING: Do not run this command during any other werf command is working on the host machine. This 
command is supposed to be run manually.

{{ header }} Syntax

```bash
werf host purge [options]
```

{{ header }} Options

```bash
      --disable-pretty-log=false:
            Disable emojis, auto line wrapping and replace log process border characters with 
            spaces (default $WERF_DISABLE_PRETTY_LOG).
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or 
            ~/.docker (in the order of priority).
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for purge
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-repo=false:
            Allow usage of insecure docker repos (default $WERF_INSECURE_REPO)
      --log-color-mode='auto':
            Set log color mode.
            Supported on, off and auto (based on the stdout's file descriptor referring to a 
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

