{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Decrypt data.

Provide encrypted data onto stdin by default.

Data can be provided in a file by specifying --file-path option. Option --values should be 
specified in the case when secret values yaml file provided.

{{ header }} Syntax

```bash
werf secret extract [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --file-path='':
            Decode file data by specified path
  -h, --help=false:
            help for extract
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --output-file-path='':
            Save decoded data by specified file path
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use system tmp dir by default)
      --values=false:
            Decode specified FILE_PATH (--file-path) as secret values file
```

{{ header }} Environments

```bash
  $WERF_SECRET_KEY  
```

