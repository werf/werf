{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Delete application from Kubernetes.

Helm Release will be purged and optionally Kubernetes Namespace.

Environment is a required param for the dismiss by default, because it is needed to construct Helm  
Release name and Kubernetes Namespace. Either --env or $WERF_ENV should be specified for command.

Read more info about Helm Release name, Kubernetes Namespace and how to change it:                  
[https://werf.io/documentation/reference/deploy_process/deploy_into_kubernetes.html](https://werf.io/documentation/reference/deploy_process/deploy_into_kubernetes.html)

{{ header }} Syntax

```shell
werf dismiss [options]
```

{{ header }} Examples

```shell
  # Dismiss project named 'myproject' previously deployed app from 'dev' environment; helm release name and namespace will be named as 'myproject-dev'
  $ werf dismiss --env dev

  # Dismiss project with namespace
  $ werf dismiss --env my-feature-branch --with-namespace

  # Dismiss project using specified helm release name and namespace
  $ werf dismiss --release myrelease --namespace myns
```

{{ header }} Options

```shell
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
      --env='':
            Use specified environment (default $WERF_ENV)
      --helm-release-storage-namespace='kube-system':
            Helm release storage namespace (same as --tiller-namespace for regular helm, default    
            $WERF_HELM_RELEASE_STORAGE_NAMESPACE, $TILLER_NAMESPACE or 'kube-system')
      --helm-release-storage-type='configmap':
            helm storage driver to use. One of 'configmap' or 'secret' (default                     
            $WERF_HELM_RELEASE_STORAGE_TYPE or 'configmap')
  -h, --help=false:
            help for dismiss
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --kube-config='':
            Kubernetes config file path
      --kube-context='':
            Kubernetes config context (default $WERF_KUBE_CONTEXT)
      --log-project-dir=false:
            Print current project directory path (default $WERF_LOG_PROJECT_DIR)
      --namespace='':
            Use specified Kubernetes namespace (default [[ project ]]-[[ env ]] template or         
            deploy.namespace custom template from werf.yaml)
      --release='':
            Use specified Helm release name (default [[ project ]]-[[ env ]] template or            
            deploy.helmRelease custom template from werf.yaml)
      --releases-history-max=0:
            Max releases to keep in release storage. Can be set by environment variable             
            $WERF_RELEASES_HISTORY_MAX. By default werf keeps all releases.
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --with-hooks=true:
            Delete Helm Release hooks getting from existing revisions
      --with-namespace=false:
            Delete Kubernetes Namespace after purging Helm Release
```

