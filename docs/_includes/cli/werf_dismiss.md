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
https://flant.github.io/werf/reference/deploy/deploy_to_kubernetes.html

{{ header }} Syntax

```bash
werf dismiss [options]
```

{{ header }} Examples

```bash
  # Dismiss project named 'myproject' previously deployed app from 'dev' environment; helm release name and namespace will be named as 'myproject-dev'
  $ werf dismiss --env dev --stages-storage :local

  # Dismiss project with namespace
  $ werf dismiss --env my-feature-branch --stages-storage :local --with-namespace

  # Dismiss project using specified helm release name and namespace
  $ werf dismiss --release myrelease --namespace myns --stages-storage :local
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --env='':
            Use specified environment (default $WERF_ENV)
  -h, --help=false:
            help for dismiss
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME environment or 
            ~/.werf)
      --kube-config='':
            Kubernetes config file path
      --kube-context='':
            Kubernetes config context
      --namespace='':
            Use specified Kubernetes namespace (default %project-%environment template)
      --release='':
            Use specified Helm release name (default %project-%environment template)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP environment or system 
            tmp dir)
      --with-namespace=false:
            Delete Kubernetes Namespace after purging Helm Release
```

