{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
List werf releases

{{ header }} Syntax

```shell
werf helm ls [FILTER] [options]
```

{{ header }} Environments

```shell

```

{{ header }} Options

```shell
  -a, --all=false:
            Show all releases, not just the ones marked DEPLOYED
      --col-width=60:
            Specifies the max column width of output
      --deleted=false:
            Show deleted releases
      --deleting=false:
            Show releases that are currently being deleted
      --helm-release-storage-namespace='kube-system':
            Helm release storage namespace (same as --tiller-namespace for regular helm, default    
            $WERF_HELM_RELEASE_STORAGE_NAMESPACE, $TILLER_NAMESPACE or 'kube-system')
      --helm-release-storage-type='configmap':
            helm storage driver to use. One of 'configmap' or 'secret' (default                     
            $WERF_HELM_RELEASE_STORAGE_TYPE or 'configmap')
  -h, --help=false:
            help for ls
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --kube-config='':
            Kubernetes config file path
      --kube-context='':
            Kubernetes config context (default $WERF_KUBE_CONTEXT)
  -m, --max=256:
            Maximum number of releases to fetch
      --namespace='':
            Show releases within a specific namespace
  -o, --offset='':
            Next release name in the list, used to offset from start value
      --output='table':
            Output the specified format (json, yaml or table)
      --pending=false:
            Show pending releases
  -r, --reverse=false:
            Reverse the sort order (descending by default)
  -q, --short=false:
            Output short listing format
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

