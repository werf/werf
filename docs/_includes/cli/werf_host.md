{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Work with werf cache and data of all projects on the host machine

{{ header }} Options

```bash
  -h, --help=false:
            help for host
```

