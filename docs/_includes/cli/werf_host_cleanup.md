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

```shell
werf host cleanup [options]
```

{{ header }} Options

```shell
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for cleanup
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-registry=false:
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
      --log-color-mode='auto':
            Set log color mode.
            Supported on, off and auto (based on the stdout’s file descriptor referring to a        
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-pretty=true:
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or   
            true).
      --log-terminal-width=-1:
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --skip-tls-verify-registry=false:
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

