{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Add a chart repository

{{ header }} Syntax

```shell
werf helm repo add [NAME] [URL] [options]
```

{{ header }} Options

```shell
      --ca-file='':
            verify certificates of HTTPS-enabled servers using this CA bundle
      --cert-file='':
            identify HTTPS client using this SSL certificate file
      --helm-home='~/.helm':
            location of your Helm config. Defaults to $WERF_HELM_HOME, $HELM_HOME or ~/.helm
  -h, --help=false:
            help for add
      --key-file='':
            identify HTTPS client using this SSL key file
      --no-update=false:
            raise error if repo is already registered
      --password='':
            chart repository password
      --username='':
            chart repository username
```

