{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Generate werf environment variables for specified CI system.

Currently supported only GitLab CI

{{ header }} Syntax

```shell
werf ci-env CI_SYSTEM [options]
```

{{ header }} Examples

```shell
  # Load generated werf environment variables on GitLab job runner
  $ source <(werf ci-env gitlab)
```

{{ header }} Options

```shell
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
            Command will copy specified or default (~/.docker) config to the temporary directory    
            and may perform additional login with new config
  -h, --help=false:
            help for ci-env
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-registry=false:
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
      --log-color-mode='auto':
            Set log color mode.
            Supported on, off and auto (based on the stdout’s file descriptor referring to a        
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-debug=false:
            Enable debug (default $WERF_LOG_DEBUG).
      --log-pretty=true:
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or   
            true).
      --log-quiet=false:
            Disable explanatory output (default $WERF_LOG_QUIET).
      --log-terminal-width=-1:
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --log-verbose=false:
            Enable verbose output (default $WERF_LOG_VERBOSE).
      --skip-tls-verify-registry=false:
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --tagging-strategy='stages-signature':
            stages-signature: always use '--tag-by-stages-signature' option to tag all published    
            images by corresponding stages-signature; tag-or-branch: generate auto                  
            '--tag-git-branch' or '--tag-git-tag' tag by specified CI_SYSTEM environment variables
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

