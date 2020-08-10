{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Build out the charts/ directory from the Chart.lock file.

Build is used to reconstruct a chart's dependencies to the state specified in
the lock file. This will not re-negotiate dependencies, as 'helm dependency update'
does.

If no lock file is found, 'helm dependency build' will mirror the behavior
of 'helm dependency update'.


{{ header }} Syntax

```shell
werf helm-v3 dependency build CHART [flags] [options]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for build
      --keyring='/home/distorhead/.gnupg/pubring.gpg':
            keyring containing public keys
      --verify=false:
            verify the packages against signatures
```

