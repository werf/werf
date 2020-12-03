{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
List image and artifact names defined in werf.yaml

{{ header }} Syntax

```shell
werf config list [options]
```

{{ header }} Options

```shell
      --config=''
            Use custom configuration file (default $WERF_CONFIG or werf.yaml in working directory)
      --config-templates-dir=''
            Custom configuration templates directory (default $WERF_CONFIG_TEMPLATES_DIR or .werf   
            in working directory)
      --dir=''
            Use custom working directory (default $WERF_DIR or current directory)
      --env=''
            Use specified environment (default $WERF_ENV)
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --images-only=false
            Show image names without artifacts
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
      --loose-giterminism=false
            Loose werf giterminism mode restrictions (NOTE: not all restrictions can be removed,    
            more info                                                                               
            https://werf.io/v1.2-alpha/documentation/advanced/configuration/giterminism.html,       
            default $WERF_LOOSE_GITERMINISM)
      --non-strict-giterminism-inspection=false
            Change some errors to warnings during giterminism inspection (more info                 
            https://werf.io/v1.2-alpha/documentation/advanced/configuration/giterminism.html,       
            default $WERF_NON_STRICT_GITERMINISM_INSPECTION)
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

