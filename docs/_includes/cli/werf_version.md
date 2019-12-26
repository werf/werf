{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Print version

{{ header }} Syntax

```shell
werf version [options]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for version
```

