{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Generate bash completion scripts

{{ header }} Syntax

```shell
werf completion [options]
```

{{ header }} Examples

```shell
  # Load completion run
  $ source <(werf completion)
```

{{ header }} Options

```shell
  -h, --help=false:
            help for completion
```

