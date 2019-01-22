{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Prints name suitable for Helm Release based on the specified NAME

{{ header }} Syntax

```bash
werf slug release NAME [options]
```

{{ header }} Options

```bash
  -h, --help=false: help for release
```

