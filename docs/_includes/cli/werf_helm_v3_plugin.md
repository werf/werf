{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Manage client-side Helm plugins.


{{ header }} Options

```shell
  -h, --help=false:
            help for plugin
```

