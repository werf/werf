{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Search the Helm Hub or an instance of Monocular for Helm charts.

The Helm Hub provides a centralized search for publicly available distributed
charts. It is maintained by the Helm project. It can be visited at
[https://hub.helm.sh](https://hub.helm.sh)

Monocular is a web-based application that enables the search and discovery of
charts from multiple Helm Chart repositories. It is the codebase that powers the
Helm Hub. You can find it at [https://github.com/helm/monocular](https://github.com/helm/monocular)


{{ header }} Syntax

```shell
werf helm-v3 search hub [keyword] [flags] [options]
```

{{ header }} Options

```shell
      --endpoint='https://hub.helm.sh':
            monocular instance to query for charts
  -h, --help=false:
            help for hub
      --max-col-width=50:
            maximum column width for output table
  -o, --output=table:
            prints the output in the specified format. Allowed values: table, json, yaml
```

