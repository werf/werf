{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Purge project stages from stages storage

{{ header }} Syntax

```shell
werf stages purge [options]
```

{{ header }} Options

```shell
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
            Command needs granted permissions to read, pull and delete images from the specified    
            stages storage
      --dry-run=false:
            Indicate what the command would do without actually doing that
      --force=false:
            Remove containers that are based on deleting werf docker images
  -h, --help=false:
            help for purge
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
      --log-project-dir=false:
            Print current project directory path (default $WERF_LOG_PROJECT_DIR)
      --log-terminal-width=-1:
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --skip-tls-verify-registry=false:
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
  -s, --stages-storage='':
            Docker Repo to store stages or :local for non-distributed build (only :local is         
            supported for now; default $WERF_STAGES_STORAGE environment).
            More info about stages: https://werf.io/documentation/reference/stages_and_images.html
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

