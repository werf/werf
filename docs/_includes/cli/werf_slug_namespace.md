{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Prints name suitable for Kubernetes Namespace based on the specified NAME

{{ header }} Syntax

```bash
werf slug namespace NAME [options]
```

{{ header }} Examples

```bash
  $ werf slug namespace feature-fix-2
  feature-fix-2

  $ werf slug namespace 'branch/one/!@#4.4-3'
  branch-one-4-4-3-4fe08955

  $ werf slug namespace My_branch
  my-branch-8ebf2d1d
```

{{ header }} Options

```bash
  -h, --help=false:
            help for namespace
```

