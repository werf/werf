{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Search provides the ability to search for Helm charts in the various places
they can be stored including the Helm Hub and repositories you have added. Use
search subcommands to search different locations for charts.


{{ header }} Options

```shell
  -h, --help=false:
            help for search
```

