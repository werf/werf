{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
list installed Helm plugins

{{ header }} Syntax

```shell
werf helm-v3 plugin list [flags] [options]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for list
```

