{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Prints name suitable for Kubernetes Namespace based on the specified NAME

{{ header }} Syntax

```bash
werf slug namespace NAME [options]
```

{{ header }} Options

```bash
  -h, --help=false: help for namespace
```

