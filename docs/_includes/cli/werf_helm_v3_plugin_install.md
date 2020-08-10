{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

This command allows you to install a plugin from a url to a VCS repo or a local path.


{{ header }} Syntax

```shell
werf helm-v3 plugin install [options] <path|url>... [flags]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for install
      --version='':
            specify a version constraint. If this is not specified, the latest version is installed
```

