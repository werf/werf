{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
remove one or more chart repositories

{{ header }} Syntax

```shell
werf helm-v3 repo remove [REPO1 [REPO2 ...]] [flags] [options]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for remove
```

