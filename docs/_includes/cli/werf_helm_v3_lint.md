{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

This command takes a path to a chart and runs a series of tests to verify that
the chart is well-formed.

If the linter encounters things that will cause the chart to fail installation,
it will emit [ERROR] messages. If it encounters issues that break with convention
or recommendation, it will emit [WARNING] messages.


{{ header }} Syntax

```shell
werf helm-v3 lint PATH [flags] [options]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for lint
      --set=[]:
            set values on the command line (can specify multiple or separate values with commas:    
            key1=val1,key2=val2)
      --set-file=[]:
            set values from respective files specified via the command line (can specify multiple   
            or separate values with commas: key1=path1,key2=path2)
      --set-string=[]:
            set STRING values on the command line (can specify multiple or separate values with     
            commas: key1=val1,key2=val2)
      --strict=false:
            fail on lint warnings
  -f, --values=[]:
            specify values in a YAML file or a URL (can specify multiple)
      --with-subcharts=false:
            lint dependent charts
```

