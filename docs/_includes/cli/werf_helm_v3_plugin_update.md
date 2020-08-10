{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
update one or more Helm plugins

{{ header }} Syntax

```shell
werf helm-v3 plugin update <plugin>... [flags] [options]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for update
```

