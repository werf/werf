{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Run synchronization server

{{ header }} Syntax

```shell
werf synchronization [options]
```

{{ header }} Options

```shell
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
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --host=''
            Bind synchronization server to the specified host (default localhost or $WERF_HOST)
      --kube-config=''
            Kubernetes config file path (default $WERF_KUBE_CONFIG, or $WERF_KUBECONFIG, or         
            $KUBECONFIG)
      --kube-config-base64=''
            Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or            
            $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)
      --kube-context=''
            Kubernetes config context (default $WERF_KUBE_CONTEXT)
      --kubernetes=false
            Use kubernetes lock-manager stages-storage-cache (default $WERF_KUBERNETES)
      --kubernetes-namespace-prefix=''
            Use specified prefix for namespaces created for lock-manager and stages-storage-cache   
            (defaults to `werf-synchronization-` when --kubernetes option is used or                
            $WERF_KUBERNETES_NAMESPACE_PREFIX)
      --local=true
            Use file lock-manager and file stages-storage-cache (true by default or $WERF_LOCAL)
      --local-lock-manager-base-dir=''
            Use specified directory as base for file lock-manager                                   
            (~/.werf/synchronization_server/lock_manager by default or                              
            $WERF_LOCAL_LOCK_MANAGER_BASE_DIR)
      --local-stages-storage-cache-base-dir=''
            Use specified directory as base for file stages-storage-cache                           
            (~/.werf/synchronization_server/stages_storage_cache by default or                      
            $WERF_LOCAL_STAGES_STORAGE_CACHE_BASE_DIR)
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
      --port=''
            Bind synchronization server to the specified port (default 55581 or $WERF_PORT)
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --ttl=''
            Time to live for lock-manager locks and stages-storage-cache records (default $WERF_TTL)
```

