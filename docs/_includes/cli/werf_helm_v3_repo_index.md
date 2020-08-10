{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Read the current directory and generate an index file based on the charts found.

This tool is used for creating an 'index.yaml' file for a chart repository. To
set an absolute URL to the charts, use '--url' flag.

To merge the generated index with an existing index file, use the '--merge'
flag. In this case, the charts found in the current directory will be merged
into the existing index, with local charts taking priority over existing charts.


{{ header }} Syntax

```shell
werf helm-v3 repo index [DIR] [flags] [options]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for index
      --merge='':
            merge the generated index into the given index
      --url='':
            url of chart repository
```

