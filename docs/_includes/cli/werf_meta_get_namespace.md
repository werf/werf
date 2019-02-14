{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Print Kubernetes Namespace that will be used in current configuration with specified params

{{ header }} Syntax

```bash
werf meta get-namespace [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --env='':
            Use specified environment (default WERF_DEPLOY_ENVIRONMENT)
  -h, --help=false:
            help for get-namespace
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default WERF_HOME environment or 
            ~/.werf)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default WERF_TMP environment or system 
            tmp dir)
```

