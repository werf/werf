{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

This command takes a path to a chart and runs a series of tests to verify that
the chart is well-formed.

If the linter encounters things that will cause the chart to fail installation,
it will emit [ERROR] messages. If it encounters issues that break with convention
or recommendation, it will emit [WARNING] messages.


{{ header }} Syntax

```shell
werf helm-v3 lint PATH [flags] [options]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for lint
      --set=[]:
            set values on the command line (can specify multiple or separate values with commas:    
            key1=val1,key2=val2)
      --set-file=[]:
            set values from respective files specified via the command line (can specify multiple   
            or separate values with commas: key1=path1,key2=path2)
      --set-string=[]:
            set STRING values on the command line (can specify multiple or separate values with     
            commas: key1=val1,key2=val2)
      --strict=false:
            fail on lint warnings
  -f, --values=[]:
            specify values in a YAML file or a URL (can specify multiple)
      --with-subcharts=false:
            lint dependent charts
```

{{ header }} Options inherited from parent commands

```shell
      --hooks-status-progress-period=5:
            Hooks status progress period in seconds. Set 0 to stop showing hooks status progress.   
            Defaults to $WERF_HOOKS_STATUS_PROGRESS_PERIOD_SECONDS or status progress period value
      --kube-config='':
            Kubernetes config file path (default $WERF_KUBE_CONFIG or $WERF_KUBECONFIG or           
            $KUBECONFIG)
      --kube-config-base64='':
            Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or            
            $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)
      --kube-context='':
            Kubernetes config context (default $WERF_KUBE_CONTEXT)
  -n, --namespace='':
            namespace scope for this request
      --status-progress-period=5:
            Status progress period in seconds. Set -1 to stop showing status progress. Defaults to  
            $WERF_STATUS_PROGRESS_PERIOD_SECONDS or 5 seconds
```

