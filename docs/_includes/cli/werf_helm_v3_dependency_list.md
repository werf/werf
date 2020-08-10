{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

List all of the dependencies declared in a chart.

This can take chart archives and chart directories as input. It will not alter
the contents of a chart.

This will produce an error if the chart cannot be loaded.


{{ header }} Syntax

```shell
werf helm-v3 dependency list CHART [flags] [options]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for list
```

