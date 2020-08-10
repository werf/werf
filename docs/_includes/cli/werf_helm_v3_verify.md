{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Verify that the given chart has a valid provenance file.

Provenance files provide cryptographic verification that a chart has not been
tampered with, and was packaged by a trusted provider.

This command can be used to verify a local chart. Several other commands provide
'--verify' flags that run the same validation. To generate a signed package, use
the 'helm package --sign' command.


{{ header }} Syntax

```shell
werf helm-v3 verify PATH [flags] [options]
```

{{ header }} Options

```shell
  -h, --help=false:
            help for verify
      --keyring='/home/distorhead/.gnupg/pubring.gpg':
            keyring containing public keys
```

