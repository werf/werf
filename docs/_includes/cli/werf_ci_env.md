{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Generate werf environment variables for specified CI system.

Currently supported only GitLab CI

{{ header }} Syntax

```bash
werf ci-env CI_SYSTEM [options]
```

{{ header }} Examples

```bash
  # Load generated werf environment variables on gitlab job runner
  $ source <(werf ci-env gitlab --tagging-strategy tag-or-branch)
```

{{ header }} Options

```bash
  -h, --help=false:
            help for ci-env
      --tagging-strategy='':
            tag-or-branch: generate auto '--tag-git-branch' or '--tag-git-tag' tag by specified 
            CI_SYSTEM environment variables
```

