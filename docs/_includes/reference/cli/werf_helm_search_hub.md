{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Search for Helm charts in the Artifact Hub or your own hub instance.

Artifact Hub is a web-based application that enables finding, installing, and publishing packages and configurations for CNCF projects, including publicly available distributed charts Helm charts. It is a Cloud Native Computing Foundation sandbox project. You can browse the hub at [https://artifacthub.io/](https://artifacthub.io/).

The `[KEYWORD]` argument accepts either a keyword string, or quoted string of rich query options. For rich query options documentation, see
[https://artifacthub.github.io/hub/api/?urls.primaryName=Monocular%20compatible%20search%20API#/Monocular/get_api_chartsvc_v1_charts_search](https://artifacthub.github.io/hub/api/?urls.primaryName=Monocular%20compatible%20search%20API#/Monocular/get_api_chartsvc_v1_charts_search).

Previous versions of Helm used an instance of Monocular as the default `endpoint`, so for backwards compatibility Artifact Hub is compatible with the Monocular search API. Similarly, when setting the `endpoint` flag, the specified endpoint must also be implement a Monocular compatible search API endpoint. Note that when specifying a Monocular instance as the `endpoint`, rich queries are not supported. For API details, see [https://github.com/helm/monocular](https://github.com/helm/monocular).


{{ header }} Syntax

```shell
werf helm search hub [KEYWORD] [flags] [options]
```

{{ header }} Options

```shell
      --endpoint='https://hub.helm.sh'
            Hub instance to query for charts
      --fail-on-no-result=false
            search fails if no results are found
      --list-repo-url=false
            print charts repository URL
      --max-col-width=50
            maximum column width for output table
  -o, --output=table
            prints the output in the specified format. Allowed values: table, json, yaml
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

