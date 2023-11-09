{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
This command lists all of the releases for a specified namespace (uses current namespace context if namespace not specified).

By default, it lists only releases that are deployed or failed. Flags like `--uninstalled` and `--all` will alter this behavior. Such flags can be combined: `--uninstalled --failed`.

By default, items are sorted alphabetically. Use the `-d` flag to sort by release date.

If the `--filter` flag is provided, it will be treated as a filter. Filters are regular expressions (Perl compatible) that are applied to the list of releases. Only items that match the filter will be returned.
```
    $ helm list --filter 'ara[a-z]+'
    NAME                UPDATED                                  CHART
    maudlin-arachnid    2020-06-18 14:17:46.125134977 +0000 UTC  alpine-0.1.0
```
If no results are found, `helm list` will exit `0`, but with no output (or in the case of no `-q` flag, only headers).

By default, up to 256 items may be returned. To limit this, use the `--max` flag. Setting `--max` to `0` will not return all results. Rather, it will return the server's default, which may be much higher than 256. Pairing the `--max` flag with the `--offset` flag allows you to page through results.


{{ header }} Syntax

```shell
werf helm list [flags] [options]
```

{{ header }} Options

```shell
  -a, --all=false
            show all releases without any filter applied
  -A, --all-namespaces=false
            list releases across all namespaces
  -d, --date=false
            sort by release date
      --deployed=false
            show deployed releases. If no other is specified, this will be automatically enabled
      --failed=false
            show failed releases
  -f, --filter=''
            a regular expression (Perl compatible). Any releases that match the expression will be  
            included in the results
  -m, --max=256
            maximum number of releases to fetch
      --no-headers=false
            don`t print headers when using the default output format
      --offset=0
            next release index in the list, used to offset from start value
  -o, --output=table
            prints the output in the specified format. Allowed values: table, json, yaml
      --pending=false
            show pending releases
  -r, --reverse=false
            reverse the sort order
  -l, --selector=''
            Selector (label query) to filter on, supports `=`, `==`, and `!=`.(e.g. -l              
            key1=value1,key2=value2). Works only for secret(default) and configmap storage backends.
  -q, --short=false
            output short (quiet) listing format
      --superseded=false
            show superseded releases
      --time-format=''
            format time using golang time formatter. Example: --time-format "2006-01-02             
            15:04:05Z0700"
      --uninstalled=false
            show uninstalled releases (if `helm uninstall --keep-history` was used)
      --uninstalling=false
            show releases that are currently being uninstalled
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

