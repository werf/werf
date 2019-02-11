{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Generate bash completion scripts

{{ header }} Syntax

```bash
werf completion [options]
```

{{ header }} Examples

```bash
  # Load completion run
  $ source <(werf completion)
```

{{ header }} Options

```bash
  -h, --help=false:
            help for completion
```

