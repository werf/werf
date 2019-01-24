{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Edit or create new secret file.

The file can be raw secret file (by default) or secret values yaml file (with option --values).

{{ header }} Syntax

```bash
werf secret edit FILE_PATH [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
  -h, --help=false:
            help for edit
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use system tmp dir by default)
      --values=false:
            Edit FILE_PATH as secret values file
```

{{ header }} Environments

```bash
  $WERF_SECRET_KEY  
  $WERF_TMP         
```

