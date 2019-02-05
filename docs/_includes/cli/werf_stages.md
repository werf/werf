{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Work with stages, which are cache for images

{{ header }} Options

```bash
  -h, --help=false:
            help for stages
```

