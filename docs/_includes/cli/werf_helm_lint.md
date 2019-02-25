{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Run lint procedure for the Werf chart

{{ header }} Syntax

```bash
werf helm lint [options]
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
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or 
            ~/.docker (in the order of priority).
      --env='':
            Use specified environment (default $WERF_ENV)
  -h, --help=false:
            help for lint
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --secret-values=[]:
            Additional helm secret values
      --set=[]:
            Additional helm sets
      --set-string=[]:
            Additional helm STRING sets
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --values=[]:
            Additional helm values
```

