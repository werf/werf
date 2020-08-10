{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Update gets the latest information about charts from the respective chart repositories.
Information is cached locally, where it is used by commands like 'helm search'.


{{ header }} Syntax

```shell
werf helm-v3 repo update [flags] [options]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for update
```

