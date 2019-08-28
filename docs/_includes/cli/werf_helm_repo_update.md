{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Update gets the latest information about charts from the respective chart repositories.
Information is cached locally, where it is used by commands like 'werf helm repo search'.


{{ header }} Syntax

```bash
werf helm repo update [options]
```

{{ header }} Options

```bash
      --helm-home='/home/aigrychev/.helm':
            location of your Helm config. Defaults to $WERF_HELM_HOME or $HELM_HOME
  -h, --help=false:
            help for update
      --strict=false:
            fail on update warnings
```

