{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Prints name suitable for Docker Tag based on the specified NAME

{{ header }} Syntax

```bash
werf slug tag NAME [options]
```

{{ header }} Options

```bash
  -h, --help=false: help for tag
```

