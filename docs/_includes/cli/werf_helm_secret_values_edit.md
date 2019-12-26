{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Edit or create new secret values file.
Encryption key should be in $WERF_SECRET_KEY or .werf_secret_key file

{{ header }} Syntax

```shell
werf helm secret values edit FILE_PATH [options]
```

{{ header }} Examples

```shell
  # Create/edit existing secret values file
  $ werf helm secret values edit .helm/secret-values.yaml
```

{{ header }} Environments

```shell
  $WERF_SECRET_KEY  Use specified secret key to extract secrets for the deploy. Recommended way to  
                    set secret key in CI-system. 
                    
                    Secret key also can be defined in files:
                    * ~/.werf/global_secret_key (globally),
                    * .werf_secret_key (per project)
```

{{ header }} Options

```shell
      --dir='':
            Change to the specified directory to find werf.yaml config
  -h, --help=false:
            help for edit
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

