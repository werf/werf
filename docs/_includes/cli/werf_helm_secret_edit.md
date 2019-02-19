{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Edit or create new secret file.

The file can be raw secret file (by default) or secret values yaml file (with option --values)

{{ header }} Syntax

```bash
werf helm secret edit FILE_PATH [options]
```

{{ header }} Environments

```bash
  $WERF_SECRET_KEY  Use specified secret key to extract secrets for the deploy; recommended way to 
                    set secret key in CI-system
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
  -h, --help=false:
            help for edit
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME environment 
            or ~/.werf)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP environment or system 
            tmp dir)
      --values=false:
            Edit FILE_PATH as secret values file
```

