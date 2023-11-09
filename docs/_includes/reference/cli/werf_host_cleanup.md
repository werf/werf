{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Cleanup old unused werf cache and data of all projects on host machine.

The data include:
* Lost Docker containers and images from interrupted builds.
* Old service tmp dirs, which werf creates during every `build`, `converge` and other commands.
* Local cache:
  * remote Git clones cache;
  * Git worktree cache.

It is safe to run this command periodically by automated cleanup job in parallel with other werf commands such as `build`, `converge` and `cleanup`.

{{ header }} Syntax

```shell
werf host cleanup [options]
```

{{ header }} Options

```shell
      --allowed-docker-storage-volume-usage=70
            Set allowed percentage of docker storage volume usage which will cause cleanup of least 
            recently used local docker images (default 70% or                                       
            $WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE)
      --allowed-docker-storage-volume-usage-margin=5
            During cleanup of least recently used local docker images werf would delete images      
            until volume usage becomes below "allowed-docker-storage-volume-usage -                 
            allowed-docker-storage-volume-usage-margin" level (default 5% or                        
            $WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE_MARGIN)
      --allowed-local-cache-volume-usage=70
            Set allowed percentage of local cache (~/.werf/local_cache by default) volume usage     
            which will cause cleanup of least recently used data from the local cache (default 70%  
            or $WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE)
      --allowed-local-cache-volume-usage-margin=5
            During cleanup of least recently used local docker images werf would delete images      
            until volume usage becomes below "allowed-docker-storage-volume-usage -                 
            allowed-docker-storage-volume-usage-margin" level (default 5% or                        
            $WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN)
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
      --disable-auto-host-cleanup=false
            Disable auto host cleanup procedure in main werf commands like werf-build,              
            werf-converge and other (default disabled or WERF_DISABLE_AUTO_HOST_CLEANUP)
      --docker-config=''
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
      --docker-server-storage-path=''
            Use specified path to the local docker server storage to check docker storage volume    
            usage while performing garbage collection of local docker images (detect local docker   
            server storage path by default or use $WERF_DOCKER_SERVER_STORAGE_PATH)
      --dry-run=false
            Indicate what the command would do without actually doing that (default $WERF_DRY_RUN)
      --force=false
            Force deletion of images which are being used by some containers (default $WERF_FORCE)
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-registry=false
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
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
      --loose-giterminism=false
            Loose werf giterminism mode restrictions (NOTE: not all restrictions can be removed,    
            more info https://werf.io/documentation/usage/project_configuration/giterminism.html,   
            default $WERF_LOOSE_GITERMINISM)
      --platform=[]
            Enable platform emulation when building images with werf, format: OS/ARCH[/VARIANT]     
            ($WERF_PLATFORM or $DOCKER_DEFAULT_PLATFORM by default)
  -N, --project-name=''
            Set a specific project name (default $WERF_PROJECT_NAME)
      --skip-tls-verify-registry=false
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

