{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Prints name suitable for Docker Tag based on the specified NAME

{{ header }} Syntax

```bash
werf slug tag NAME [options]
```

{{ header }} Examples

```bash
  $ werf slug tag helo/ehlo
  helo-ehlo-b6f6ab1f

  $ werf slug tag 16.04
  16.04
```

{{ header }} Options

```bash
  -h, --help=false:
            help for tag
```

