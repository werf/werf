{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}


{{ header }} Syntax

```bash
werf render [options]
```

{{ header }} Options

```bash
      --dir='': Change to the specified directory to find werf.yaml config
  -h, --help=false: help for render
      --home-dir='': Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --secret-values=[]: Additional helm secret values
      --set=[]: Additional helm sets
      --set-string=[]: Additional helm STRING sets
      --tmp-dir='': Use specified dir to store tmp files and dirs (use system tmp dir by default)
      --values=[]: Additional helm values
```

{{ header }} Environments

```bash
  $WERF_SECRET_KEY  
```

