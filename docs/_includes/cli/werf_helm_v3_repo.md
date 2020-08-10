{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

This command consists of multiple subcommands to interact with chart repositories.

It can be used to add, remove, list, and index chart repositories.


{{ header }} Options

```shell
  -h, --help=false:
            help for repo
```

