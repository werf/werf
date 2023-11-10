{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Start a migration of your existing Helm 2 release to Helm 3.

{{ header }} Syntax

```shell
werf helm migrate2to3 [options]
```

{{ header }} Options

```shell
      --helm2-release-storage-namespace=''
            Helm 2 release storage namespace (same as --tiller-namespace for regular helm 2,        
            defaults to $WERF_HELM2_RELEASE_STORAGE_NAMESPACE, or                                   
            $WERF_HELM_RELEASE_STORAGE_NAMESPACE, or $TILLER_NAMESPACE, or "kube-system")
      --helm2-release-storage-type=''
            Helm w storage driver to use. One of "configmap" or "secret", defaults to               
            $WERF_HELM2_RELEASE_STORAGE_TYPE, or $WERF_HELM_RELEASE_STORAGE_TYPE, or "configmap")
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-helm-dependencies=false
            Allow insecure oci registries to be used in the .helm/Chart.yaml dependencies           
            configuration (default $WERF_INSECURE_HELM_DEPENDENCIES)
      --kube-config=''
            Kubernetes config file path (default $WERF_KUBE_CONFIG, or $WERF_KUBECONFIG, or         
            $KUBECONFIG)
      --kube-config-base64=''
            Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or            
            $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)
      --kube-context=''
            Kubernetes config context (default $WERF_KUBE_CONTEXT)
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
      --release=''
            Existing helm 2 release name which should be migrated to helm 3 (default                
            $WERF_RELEASE). Option also sets target name for a new helm 3 release, use              
            --target-release option (or $WERF_TARGET_RELEASE) to specify a different helm 3 release 
            name.
      --target-namespace=''
            Target kubernetes namespace for a new helm 3 release (default $WERF_NAMESPACE)
      --target-release=''
            Target helm 3 release name (optional, default $WERF_TARGET_RELEASE, or the value of     
            --release option, or $WERF_RELEASE)
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

{{ header }} Options inherited from parent commands

```shell
      --hooks-status-progress-period=5
            Hooks status progress period in seconds. Set 0 to stop showing hooks status progress.   
            Defaults to $WERF_HOOKS_STATUS_PROGRESS_PERIOD_SECONDS or status progress period value
  -n, --namespace=''
            namespace scope for this request
      --status-progress-period=5
            Status progress period in seconds. Set -1 to stop showing status progress. Defaults to  
            $WERF_STATUS_PROGRESS_PERIOD_SECONDS or 5 seconds
```

