{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Prints name suitable for Helm Release based on the specified NAME

{{ header }} Syntax

```bash
werf slug release NAME [options]
```

{{ header }} Examples

```bash
  $ werf slug release my_release-NAME
  my_release-NAME

  The result has been trimmed to fit maximum bytes limit:
  $ werf slug release looooooooooooooooooooooooooooooooooooooooooong_string
  looooooooooooooooooooooooooooooooooooooooooong-stri-b150a895
```

{{ header }} Options

```bash
  -h, --help=false:
            help for release
```

