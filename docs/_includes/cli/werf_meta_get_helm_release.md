{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Print Helm Release name that will be used in current configuration with specified params

{{ header }} Syntax

```bash
werf meta get-helm-release [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --env='':
            Use specified environment (use WERF_DEPLOY_ENVIRONMENT by default)
  -h, --help=false:
            help for get-helm-release
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or 
            ~/.werf by default)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use WERF_TMP environment or system tmp 
            dir by default)
```

