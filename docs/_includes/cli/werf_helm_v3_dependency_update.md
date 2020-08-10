{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Update the on-disk dependencies to mirror Chart.yaml.

This command verifies that the required charts, as expressed in 'Chart.yaml',
are present in 'charts/' and are at an acceptable version. It will pull down
the latest charts that satisfy the dependencies, and clean up old dependencies.

On successful update, this will generate a lock file that can be used to
rebuild the dependencies to an exact version.

Dependencies are not required to be represented in 'Chart.yaml'. For that
reason, an update command will not remove charts unless they are (a) present
in the Chart.yaml file, but (b) at the wrong version.


{{ header }} Syntax

```shell
werf helm-v3 dependency update CHART [flags] [options]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for update
      --keyring='/home/distorhead/.gnupg/pubring.gpg':
            keyring containing public keys
      --skip-refresh=false:
            do not refresh the local repository cache
      --verify=false:
            verify the packages against signatures
```

