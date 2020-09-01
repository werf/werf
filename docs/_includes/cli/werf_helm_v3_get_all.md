{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

This command prints a human readable collection of information about the
notes, hooks, supplied values, and generated manifest file of the given release.


{{ header }} Syntax

```shell
werf helm-v3 get all RELEASE_NAME [flags] [options]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for all
      --revision=0:
            get the named release with revision
      --template='':
            go template for formatting the output, eg: {{.Release.Name}}
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

