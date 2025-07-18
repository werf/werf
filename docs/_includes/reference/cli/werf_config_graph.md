{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Print dependency graph for images in `werf.yaml`.

{{ header }} Syntax

```shell
werf config graph [IMAGE_NAME...] [options]
```

{{ header }} Examples

```shell
  # Print dependency graph
  $ werf config graph
  - image: app1
    dependsOn:
      from: baseImage
  - image: app2
    dependsOn:
      from: baseImage
      import:
      - app1
  - image: baseImage

  # Print dependency graph for a certain image
  $ werf config graph app
  - image: app2
    dependsOn:
      from: baseImage
      import:
      - app1

```

{{ header }} Options

```shell
      --allow-includes-update=false
            Allow use includes latest versions (default $WERF_ALLOW_INCLUDES_UPDATE or false)
      --config=""
            Use custom configuration file (default $WERF_CONFIG or werf.yaml in the project         
            directory)
      --config-templates-dir=""
            Custom configuration templates directory (default $WERF_CONFIG_TEMPLATES_DIR or .werf   
            in working directory)
      --debug-templates=false
            Enable debug mode for Go templates (default $WERF_DEBUG_TEMPLATES or false)
      --dev=false
            Enable development mode (default $WERF_DEV).
            The mode allows working with project files without doing redundant commits during       
            debugging and development
      --dev-branch="_werf-dev"
            Set dev git branch name (default $WERF_DEV_BRANCH or "_werf-dev")
      --dev-ignore=[]
            Add rules to ignore tracked and untracked changes in development mode (can specify      
            multiple).
            Also, can be specified with $WERF_DEV_IGNORE_* (e.g. $WERF_DEV_IGNORE_TESTS=*_test.go,  
            $WERF_DEV_IGNORE_DOCS=path/to/docs)
      --dir=""
            Use specified project directory where project’s werf.yaml and other configuration files 
            should reside (default $WERF_DIR or current working directory)
      --env=""
            Use specified environment (default $WERF_ENV)
      --final-images-only=false
            Process final images only ($WERF_FINAL_IMAGES_ONLY or false by default)
      --git-work-tree=""
            Use specified git work tree dir (default $WERF_WORK_TREE or lookup for directory that   
            contains .git in the current or parent directories)
      --giterminism-config=""
            Custom path to the giterminism configuration file relative to working directory         
            (default $WERF_GITERMINISM_CONFIG or werf-giterminism.yaml in working directory)
      --home-dir=""
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --log-color-mode="auto"
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
      --log-time=false
            Add time to log entries for precise event time tracking (default $WERF_LOG_TIME or      
            false).
      --log-time-format="2006-01-02T15:04:05Z07:00"
            Specify custom log time format (default $WERF_LOG_TIME_FORMAT or RFC3339 format).
      --log-verbose=false
            Enable verbose output (default $WERF_LOG_VERBOSE).
      --loose-giterminism=false
            Loose werf giterminism mode restrictions
      --tmp-dir=""
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

