{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Generate encryption key

{{ header }} Syntax

```bash
werf secret keygen [options]
```

{{ header }} Options

```bash
  -h, --help=false: help for keygen
```

