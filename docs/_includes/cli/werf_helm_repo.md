{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

This command consists of multiple subcommands to interact with chart repositories.
It can be used:
* to init, add, update, remove and list chart repositories 
* to search and fetch charts


{{ header }} Options

```shell
  -h, --help=false:
            help for repo
```

