{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Search reads through all of the repositories configured on the system, and looks for matches. Search of these repositories uses the metadata stored on the system.

It will display the latest stable versions of the charts found. If you specify the `--devel` flag, the output will include pre-release versions. If you want to search using a version constraint, use `--version`.

Examples:
```
# Search for stable release versions matching the keyword 'nginx'
$ helm search repo nginx

# Search for release versions matching the keyword 'nginx', including pre-release versions
$ helm search repo nginx --devel

# Search for the latest stable release for nginx-ingress with a major version of 1
$ helm search repo nginx-ingress --version ^1.0.0
```
Repositories are managed with `helm repo` commands.


{{ header }} Syntax

```shell
werf helm search repo [keyword] [flags] [options]
```

{{ header }} Options

```shell
      --devel=false
            use development versions (alpha, beta, and release candidate releases), too. Equivalent 
            to version `>0.0.0-0`. If --version is set, this is ignored
      --fail-on-no-result=false
            search fails if no results are found
      --max-col-width=50
            maximum column width for output table
  -o, --output=table
            prints the output in the specified format. Allowed values: table, json, yaml
  -r, --regexp=false
            use regular expressions for searching repositories you have added
      --version=''
            search using semantic versioning constraints on repositories you have added
  -l, --versions=false
            show the long listing, with each version of each chart on its own line, for             
            repositories you have added
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

