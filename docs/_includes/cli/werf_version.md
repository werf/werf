{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Print version

{{ header }} Syntax

```bash
werf version [options]
```

{{ header }} Options

```bash
  -h, --help=false:
            help for version
```

