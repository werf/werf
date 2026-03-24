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
      --allowed-backend-storage-volume-usage=70
            Set the cleanup threshold for backend (Docker or Buildah) storage. Plain numbers define 
            the maximum allowed usage percentage, while values with units define the minimum        
            required free space, e.g. 70 or 10GB (default 70 or                                     
            $WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE)
      --allowed-backend-storage-volume-usage-margin=5
            Set the cleanup margin for backend (Docker or Buildah) storage. In percentage mode the  
            margin is subtracted from the usage threshold, while in bytes mode it is added to the   
            minimum required free space threshold, e.g. 5 or 2GB (default 5 or                      
            $WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE_MARGIN)
      --allowed-local-cache-volume-usage=70
            Set the cleanup threshold for local cache (~/.werf/local_cache by default). Plain       
            numbers define the maximum allowed usage percentage, while values with units define the 
            minimum required free space, e.g. 70 or 10GB (default 70 or                             
            $WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE)
      --allowed-local-cache-volume-usage-margin=5
            Set the cleanup margin for local cache. In percentage mode the margin is subtracted     
            from the usage threshold, while in bytes mode it is added to the minimum required free  
            space threshold, e.g. 5 or 2GB (default 5 or                                            
            $WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN)
      --backend-storage-path=""
            Use specified path to the local backend (Docker or Buildah) storage to check backend    
            storage volume usage while performing garbage collection of local backend images        
            (detect local backend storage path by default or use $WERF_BACKEND_STORAGE_PATH)
      --container-registry-mirror=[]
            (Buildah-only) Use specified mirrors for docker.io
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
      --disable-auto-host-cleanup=false
            Disable auto host cleanup procedure in main werf commands like werf-build,              
            werf-converge and other (default disabled or WERF_DISABLE_AUTO_HOST_CLEANUP)
      --docker-config=""
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
      --dry-run=false
            Indicate what the command would do without actually doing that (default $WERF_DRY_RUN)
      --force=false
            Force deletion of images which are being used by some containers (default $WERF_FORCE)
      --home-dir=""
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-registry=false
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
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
      --platform=[]
            Enable platform emulation when building images with werf, format: OS/ARCH[/VARIANT]     
            ($WERF_PLATFORM or $DOCKER_DEFAULT_PLATFORM by default)
  -N, --project-name=""
            Set a specific project name (default $WERF_PROJECT_NAME)
      --skip-tls-verify-registry=false
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --tmp-dir=""
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

