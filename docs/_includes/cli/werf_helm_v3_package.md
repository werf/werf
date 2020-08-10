{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

This command packages a chart into a versioned chart archive file. If a path
is given, this will look at that path for a chart (which must contain a
Chart.yaml file) and then package that directory.

Versioned chart archives are used by Helm package repositories.

To sign a chart, use the '--sign' flag. In most cases, you should also
provide '--keyring path/to/secret/keys' and '--key keyname'.

  $ helm package --sign ./mychart --key mykey --keyring ~/.gnupg/secring.gpg

If '--keyring' is not specified, Helm usually defaults to the public keyring
unless your environment is otherwise configured.


{{ header }} Syntax

```shell
werf helm-v3 package [CHART_PATH] [...] [flags] [options]
```

{{ header }} Options

```shell
      --app-version='':
            set the appVersion on the chart to this version
  -u, --dependency-update=false:
            update dependencies from "Chart.yaml" to dir "charts/" before packaging
  -d, --destination='.':
            location to write the chart.
  -h, --help=false:
            help for package
      --key='':
            name of the key to use when signing. Used if --sign is true
      --keyring='/home/distorhead/.gnupg/pubring.gpg':
            location of a public keyring
      --sign=false:
            use a PGP private key to sign this package
      --version='':
            set the version on the chart to this semver version
```

