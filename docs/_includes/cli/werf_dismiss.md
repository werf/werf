{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Delete application from Kubernetes.

Helm Release will be purged and optionally Kubernetes Namespace.

Environment is a required param for the dismiss by default, because it is needed to construct Helm 
Release name and Kubernetes Namespace. Either --env or CI_ENVIRONMENT_SLUG should be specified for 
command.

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
            Use specified environment (use CI_ENVIRONMENT_SLUG by default). Environment is a 
            required parameter and should be specified with option or CI_ENVIRONMENT_SLUG variable.
  -h, --help=false:
            help for dismiss
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --kube-context='':
            Kubernetes config context
      --namespace='':
            Use specified Kubernetes namespace (use %project-%environment template by default)
      --release='':
            Use specified Helm release name (use %project-%environment template by default)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use system tmp dir by default)
      --with-namespace=false:
            Delete Kubernetes Namespace after purging Helm Release
```

