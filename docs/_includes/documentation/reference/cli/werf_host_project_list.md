{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
List project names based on local storage

{{ header }} Syntax

```shell
werf host project list [options]
```

{{ header }} Options

```shell
      --disable-gitermenism=false
            Disable werf gitermenism mode (more info                                                
            https://werf.io/v1.2-alpha/documentation/advanced/configuration/gitermenism.html,       
            default $WERF_DISABLE_GITERMENISM)
      --docker-config=''
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --log-color-mode='auto'
            Set log color mode.
            Supported on, off and auto (based on the stdout’s file descriptor referring to a        
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-debug=false
            Enable debug (default $WERF_LOG_DEBUG).
      --log-pretty=true
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or   
            true).
      --log-quiet=false
            Disable explanatory output (default $WERF_LOG_QUIET).
      --log-terminal-width=-1
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --log-verbose=false
            Enable verbose output (default $WERF_LOG_VERBOSE).
  -q, --names-only=false
            Only show project names
      --non-strict-gitermenism-inspection=false
            Change some errors to warnings during gitermenism inspection (more info                 
            https://werf.io/v1.2-alpha/documentation/advanced/configuration/gitermenism.html,       
            default $WERF_NON_STRICT_GITERMENISM_INSPECTION)
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

