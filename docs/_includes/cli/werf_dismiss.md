{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Delete application from Kubernetes.

Helm Release will be purged and optionally Kubernetes Namespace.

Environment is a required param for the dismiss by default, because it is needed to construct Helm 
Release name and Kubernetes Namespace. Either --env or WERF_DEPLOY_ENVIRONMENT should be specified 
for command.

Read more info about Helm Release name, Kubernetes Namespace and how to change it: 
https://flant.github.io/werf/reference/deploy/deploy_to_kubernetes.html

{{ header }} Syntax

```bash
werf dismiss [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --env='':
            Use specified environment (use WERF_DEPLOY_ENVIRONMENT by default)
  -h, --help=false:
            help for dismiss
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or 
            ~/.werf by default)
      --kube-context='':
            Kubernetes config context
      --namespace='':
            Use specified Kubernetes namespace (use %project-%environment template by default)
      --release='':
            Use specified Helm release name (use %project-%environment template by default)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use WERF_TMP environment or system tmp 
            dir by default)
      --with-namespace=false:
            Delete Kubernetes Namespace after purging Helm Release
```

