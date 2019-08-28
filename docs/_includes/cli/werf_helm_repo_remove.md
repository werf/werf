{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Remove a chart repository

{{ header }} Syntax

```bash
werf helm repo remove [NAME] [options]
```

{{ header }} Options

```bash
      --helm-home='/home/aigrychev/.helm':
            location of your Helm config. Defaults to $WERF_HELM_HOME or $HELM_HOME
  -h, --help=false:
            help for remove
```

