{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Work with managed images which will be preserved during cleanup procedure

{{ header }} Options

```shell
  -h, --help=false:
            help for managed-images
```

