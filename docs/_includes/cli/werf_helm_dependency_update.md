{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Update the on-disk dependencies to mirror the requirements.yaml file.

This command verifies that the required charts, as expressed in 'requirements.yaml',
are present in 'charts/' and are at an acceptable version. It will pull down
the latest charts that satisfy the dependencies, and clean up old dependencies.

On successful update, this will generate a lock file that can be used to
rebuild the requirements to an exact version.

Dependencies are not required to be represented in 'requirements.yaml'. For that
reason, an update command will not remove charts unless they are (a) present
in the requirements.yaml file, but (b) at the wrong version.


{{ header }} Syntax

```shell
werf helm dependency update [options]
```

{{ header }} Options

```shell
      --dir='':
            Change to the specified directory to find werf.yaml config
      --helm-home='~/.helm':
            location of your Helm config. Defaults to $WERF_HELM_HOME, $HELM_HOME or ~/.helm
  -h, --help=false:
            help for update
      --keyring='$HOME/.gnupg/pubring.gpg':
            keyring containing public keys
      --skip-refresh=false:
            do not refresh the local repository cache
      --verify=false:
            verify the packages against signatures
```

