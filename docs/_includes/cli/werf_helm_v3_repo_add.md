{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
add a chart repository

{{ header }} Syntax

```shell
werf helm-v3 repo add [NAME] [URL] [flags] [options]
```

{{ header }} Options

```shell
      --ca-file='':
            verify certificates of HTTPS-enabled servers using this CA bundle
      --cert-file='':
            identify HTTPS client using this SSL certificate file
  -h, --help=false:
            help for add
      --insecure-skip-tls-verify=false:
            skip tls certificate checks for the repository
      --key-file='':
            identify HTTPS client using this SSL key file
      --no-update=false:
            raise error if repo is already registered
      --password='':
            chart repository password
      --username='':
            chart repository username
```

