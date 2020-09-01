{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Read the current directory and generate an index file based on the charts found.

This tool is used for creating an 'index.yaml' file for a chart repository. To
set an absolute URL to the charts, use '--url' flag.

To merge the generated index with an existing index file, use the '--merge'
flag. In this case, the charts found in the current directory will be merged
into the existing index, with local charts taking priority over existing charts.


{{ header }} Syntax

```shell
werf helm-v3 repo index [DIR] [flags] [options]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for index
      --merge='':
            merge the generated index into the given index
      --url='':
            url of chart repository
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

