{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
This command takes a release name and uninstalls the release.

It removes all of the resources associated with the last release of the chart as well as the release history, freeing it up for future use.

Use the `--dry-run` flag to see which releases will be uninstalled without actually uninstalling them.

{{ header }} Syntax

```shell
werf helm uninstall RELEASE_NAME [...] [flags] [options]
```

{{ header }} Options

```shell
      --cascade='background'
            Must be "background", "orphan", or "foreground". Selects the deletion cascading         
            strategy for the dependents. Defaults to background.
      --description=''
            add a custom description
      --dry-run=false
            simulate a uninstall
      --ignore-not-found=false
            Treat "release not found" as a successful uninstall
      --keep-history=false
            remove all associated resources and mark the release as deleted, but retain the release 
            history
      --no-hooks=false
            prevent hooks from running during uninstallation
      --timeout=5m0s
            time to wait for any individual Kubernetes operation (like Jobs for hooks)
      --wait=false
            if set, will wait until all the resources are deleted before returning. It will wait    
            for as long as --timeout
```

{{ header }} Options inherited from parent commands

```shell
      --hooks-status-progress-period=5
            Hooks status progress period in seconds. Set 0 to stop showing hooks status progress.   
            Defaults to $WERF_HOOKS_STATUS_PROGRESS_PERIOD_SECONDS or status progress period value
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
  -n, --namespace=''
            namespace scope for this request
      --status-progress-period=5
            Status progress period in seconds. Set -1 to stop showing status progress. Defaults to  
            $WERF_STATUS_PROGRESS_PERIOD_SECONDS or 5 seconds
```

