{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Build out the charts/ directory from the requirements.lock file.

Build is used to reconstruct a chart's dependencies to the state specified in
the lock file.

If no lock file is found, 'werf helm dependency build' will mirror the behavior of
the 'werf helm dependency update' command. This means it will update the on-disk
dependencies to mirror the requirements.yaml file and generate a lock file.


{{ header }} Syntax

```shell
werf helm dependency build [options]
```

{{ header }} Options

```shell
      --dir='':
            Change to the specified directory to find werf.yaml config
      --helm-home='~/.helm':
            location of your Helm config. Defaults to $WERF_HELM_HOME, $HELM_HOME or ~/.helm
  -h, --help=false:
            help for build
      --keyring='$HOME/.gnupg/pubring.gpg':
            keyring containing public keys
      --verify=false:
            verify the packages against signatures
```

