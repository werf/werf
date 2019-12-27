{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Generate hex encryption key. 
For further usage, the encryption key should be saved in $WERF_SECRET_KEY or .werf_secret_key file

{{ header }} Syntax

```shell
werf helm secret generate-secret-key [options]
```

{{ header }} Examples

```shell
  # Export encryption key
  $ export WERF_SECRET_KEY=$(werf helm secret generate-secret-key)

  # Save encryption key in .werf_secret_key file
  $ werf helm secret generate-secret-key > .werf_secret_key
```

{{ header }} Options

```shell
  -h, --help=false:
            help for generate-secret-key
```

