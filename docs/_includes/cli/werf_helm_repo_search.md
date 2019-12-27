{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Search reads through all of the repositories configured on the system, and
looks for matches


{{ header }} Syntax

```shell
werf helm repo search [keyword] [options]
```

{{ header }} Options

```shell
      --col-width=60:
            specifies the max column width of output
      --helm-home='~/.helm':
            location of your Helm config. Defaults to $WERF_HELM_HOME, $HELM_HOME or ~/.helm
  -h, --help=false:
            help for search
  -r, --regexp=false:
            use regular expressions for searching
  -v, --version='':
            search using semantic versioning constraints
  -l, --versions=false:
            show the long listing, with each version of each chart on its own line
```

