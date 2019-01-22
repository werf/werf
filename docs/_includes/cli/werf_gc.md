{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}


{{ header }} Syntax

```bash
werf gc [options]
```

{{ header }} Options

```bash
  -h, --help=false: help for gc
      --home-dir='': Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --tmp-dir='': Use specified dir to store tmp files and dirs (use system tmp dir by default)
```

