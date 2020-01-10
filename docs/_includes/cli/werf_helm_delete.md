{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Delete release from Kubernetes with all resources associated with the last release revision

{{ header }} Syntax

```shell
werf helm delete RELEASE_NAME [options]
```

{{ header }} Environments

```shell

```

{{ header }} Options

```shell
      --helm-release-storage-namespace='kube-system':
            Helm release storage namespace (same as --tiller-namespace for regular helm, default    
            $WERF_HELM_RELEASE_STORAGE_NAMESPACE, $TILLER_NAMESPACE or 'kube-system')
      --helm-release-storage-type='configmap':
            helm storage driver to use. One of 'configmap' or 'secret' (default                     
            $WERF_HELM_RELEASE_STORAGE_TYPE or 'configmap')
  -h, --help=false:
            help for delete
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --kube-config='':
            Kubernetes config file path
      --kube-context='':
            Kubernetes config context (default $WERF_KUBE_CONTEXT)
      --no-hooks=false:
            Prevent hooks from running during deletion
      --purge=false:
            Remove the release from the store and make its name free for later use
      --timeout=300:
            Time in seconds to wait for any individual Kubernetes operation (like Jobs for hooks)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

