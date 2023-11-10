{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Generate werf environment variables for specified CI system.

Currently supported only GitLab (gitlab) and GitHub (github) CI systems

{{ header }} Syntax

```shell
werf ci-env CI_SYSTEM [options]
```

{{ header }} Examples

```shell
  # Load generated werf environment variables on GitLab job runner
  $ . $(werf ci-env gitlab --as-file)

  # Load generated werf environment variables on GitLab job runner using powershell
  $ Invoke-Expression -Command "werf ci-env gitlab --as-file --shell powershell" | Out-String -OutVariable WERF_CI_ENV_SCRIPT_PATH
  $ . $WERF_CI_ENV_SCRIPT_PATH.Trim()

  # Load generated werf environment variables on GitLab job runner using cmd.exe
  $ FOR /F "tokens=*" %g IN ('werf ci-env gitlab --as-file --shell cmdexe') do (SET WERF_CI_ENV_SCRIPT_PATH=%g)
  $ %WERF_CI_ENV_SCRIPT_PATH%
```

{{ header }} Options

```shell
      --as-env-file=false
            Create the .env file and print the path for sourcing (default $WERF_AS_ENV_FILE).
      --as-file=false
            Create the script and print the path for sourcing (default $WERF_AS_FILE).
      --config=''
            Use custom configuration file (default $WERF_CONFIG or werf.yaml in working directory)
      --config-templates-dir=''
            Custom configuration templates directory (default $WERF_CONFIG_TEMPLATES_DIR or .werf   
            in working directory)
      --dev=false
            Enable development mode (default $WERF_DEV).
            The mode allows working with project files without doing redundant commits during       
            debugging and development
      --dev-branch='_werf-dev'
            Set dev git branch name (default $WERF_DEV_BRANCH or "_werf-dev")
      --dev-ignore=[]
            Add rules to ignore tracked and untracked changes in development mode (can specify      
            multiple).
            Also, can be specified with $WERF_DEV_IGNORE_* (e.g. $WERF_DEV_IGNORE_TESTS=*_test.go,  
            $WERF_DEV_IGNORE_DOCS=path/to/docs)
      --dir=''
            Use specified project directory where project’s werf.yaml and other configuration files 
            should reside (default $WERF_DIR or current working directory)
      --docker-config=''
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
            Command will copy specified or default (~/.docker) config to the temporary directory    
            and may perform additional login with new config.
      --env=''
            Use specified environment (default $WERF_ENV)
      --git-work-tree=''
            Use specified git work tree dir (default $WERF_WORK_TREE or lookup for directory that   
            contains .git in the current or parent directories)
      --giterminism-config=''
            Custom path to the giterminism configuration file relative to working directory         
            (default $WERF_GITERMINISM_CONFIG or werf-giterminism.yaml in working directory)
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
      --log-time=false
            Add time to log entries for precise event time tracking (default $WERF_LOG_TIME or      
            false).
      --log-time-format='2006-01-02T15:04:05Z07:00'
            Specify custom log time format (default $WERF_LOG_TIME_FORMAT or RFC3339 format).
      --log-verbose=false
            Enable verbose output (default $WERF_LOG_VERBOSE).
      --login-to-registry=true
            Log in to CI-specific registry automatically if possible (default                       
            $WERF_LOGIN_TO_REGISTRY).
      --loose-giterminism=false
            Loose werf giterminism mode restrictions (NOTE: not all restrictions can be removed,    
            more info https://werf.io/documentation/usage/project_configuration/giterminism.html,   
            default $WERF_LOOSE_GITERMINISM)
  -o, --output-file-path=''
            Write to custom file (default $WERF_OUTPUT_FILE_PATH).
      --platform=[]
            Enable platform emulation when building images with werf, format: OS/ARCH[/VARIANT]     
            ($WERF_PLATFORM or $DOCKER_DEFAULT_PLATFORM by default)
      --shell=''
            Set to cmdexe, powershell or use the default behaviour that is compatible with any unix 
            shell (default $WERF_SHELL).
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

