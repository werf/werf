{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Generate hex key that can be used as WERF_SECRET_KEY.

16-bytes key will be generated (AES-128).

{{ header }} Syntax

```bash
werf secret keygen [options]
```

{{ header }} Options

```bash
  -h, --help=false:
            help for keygen
```

