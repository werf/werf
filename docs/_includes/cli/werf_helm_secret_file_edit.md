{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Edit or create new secret file.
Encryption key should be in $WERF_SECRET_KEY or .werf_secret_key file

{{ header }} Syntax

```bash
werf helm secret file edit FILE_PATH [options]
```

{{ header }} Examples

```bash
  # Create/edit existing secret file
  $ werf helm secret file edit .helm/secret/privacy
```

{{ header }} Environments

```bash
  $WERF_SECRET_KEY  Use specified secret key to extract secrets for the deploy. Recommended way to 
                    set secret key in CI-system. 
                    
                    Secret key also can be defined in files:
                    * ~/.werf/global_secret_key (globally),
                    * .werf_secret_key (per project)
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
  -h, --help=false:
            help for edit
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

