{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Remove a chart repository

{{ header }} Syntax

```shell
werf helm repo remove [NAME] [options]
```

{{ header }} Options

```shell
      --helm-home='~/.helm':
            location of your Helm config. Defaults to $WERF_HELM_HOME, $HELM_HOME or ~/.helm
  -h, --help=false:
            help for remove
```

