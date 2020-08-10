{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Env prints out all the environment information in use by Helm.


{{ header }} Syntax

```shell
werf helm-v3 env [flags] [options]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for env
```

