{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

This command inspects a chart (directory, file, or URL) and displays all its content
(values.yaml, Charts.yaml, README)


{{ header }} Syntax

```shell
werf helm-v3 show all [CHART] [flags] [options]
```

{{ header }} Options

```shell
      --ca-file='':
            verify certificates of HTTPS-enabled servers using this CA bundle
      --cert-file='':
            identify HTTPS client using this SSL certificate file
      --devel=false:
            use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set,   
            this is ignored
  -h, --help=false:
            help for all
      --key-file='':
            identify HTTPS client using this SSL key file
      --keyring='~/.gnupg/pubring.gpg':
            location of public keys used for verification
      --password='':
            chart repository password where to locate the requested chart
      --repo='':
            chart repository url where to locate the requested chart
      --username='':
            chart repository username where to locate the requested chart
      --verify=false:
            verify the package before installing it
      --version='':
            specify the exact chart version to install. If this is not specified, the latest        
            version is installed
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

