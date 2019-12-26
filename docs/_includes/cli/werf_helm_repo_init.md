{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Init default chart repositories configuration

{{ header }} Syntax

```shell
werf helm repo init [options]
```

{{ header }} Options

```shell
      --helm-home='~/.helm':
            location of your Helm config. Defaults to $WERF_HELM_HOME, $HELM_HOME or ~/.helm
  -h, --help=false:
            help for init
      --local-repo-url='http://127.0.0.1:8879/charts':
            URL for local repository
      --skip-refresh=false:
            do not refresh (download) the local repository cache
      --stable-repo-url='https://kubernetes-charts.storage.googleapis.com':
            URL for stable repository
```

