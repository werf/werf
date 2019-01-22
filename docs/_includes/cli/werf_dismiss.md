{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}


{{ header }} Syntax

```bash
werf dismiss [options]
```

{{ header }} Options

```bash
      --dir='': Change to the specified directory to find werf.yaml config
      --environment='': Use specified environment (use CI_ENVIRONMENT_SLUG by default). Environment is a required parameter and should be specified with option or CI_ENVIRONMENT_SLUG variable.
  -h, --help=false: help for dismiss
      --home-dir='': Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --kube-context='': Kubernetes config context
      --namespace='': Use specified Kubernetes namespace (use %project-%environment template by default)
      --release='': Use specified Helm release name (use %project-%environment template by default)
      --tmp-dir='': Use specified dir to store tmp files and dirs (use system tmp dir by default)
      --with-namespace=false: Delete Kubernetes Namespace after purging Helm Release
```

