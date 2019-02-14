{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Encrypt provided data.

Provide data onto stdin by default.

Data can be provided in file by specifying --file-path option. Option --values should be specified 
in the case when values yaml file provided

{{ header }} Syntax

```bash
werf helm secret generate [options]
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
      --file-path='':
            Encode file data by specified path
  -h, --help=false:
            help for generate
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default WERF_HOME environment or 
            ~/.werf)
      --output-file-path='':
            Save encoded data by specified file path
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default WERF_TMP environment or system 
            tmp dir)
      --values=false:
            Encode specified FILE_PATH (--file-path) as secret values file
```

