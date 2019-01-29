{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Prints Helm Release name that will be used in current configuration with specified params

{{ header }} Syntax

```bash
werf get-release [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --env='':
            Use specified environment (use CI_ENVIRONMENT_SLUG by default). Environment is a 
            required parameter and should be specified with option or CI_ENVIRONMENT_SLUG variable.
  -h, --help=false:
            help for get-release
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use system tmp dir by default)
```

