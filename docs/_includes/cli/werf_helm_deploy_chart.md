{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Deploy Helm chart specified by path.

If specified Helm chart is a Werf chart with additional values and contains werf-chart.yaml, then 
werf will pass all additinal values and data into helm

{{ header }} Syntax

```bash
werf helm deploy-chart PATH RELEASE_NAME [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
  -h, --help=false:
            help for deploy-chart
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or 
            ~/.werf by default)
      --kube-context='':
            Kubernetes config context
      --namespace='':
            Namespace to install release into
      --set=[]:
            Additional helm sets
      --set-string=[]:
            Additional helm STRING sets
  -t, --timeout=0:
            Resources tracking timeout in seconds
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use WERF_TMP environment or system tmp 
            dir by default)
      --values=[]:
            Additional helm values
```

