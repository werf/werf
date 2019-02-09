{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Cleanup project stages from stages storage for the images, that do not exist in the specified 
images repo

{{ header }} Syntax

```bash
werf stages cleanup [options]
```

{{ header }} Environments

```bash
  $WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY  Redefine default 
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. WERF_DOCKER_CONFIG or DOCKER_CONFIG or ~/.docker 
            will be used by default (in the order of priority).
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for cleanup
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or 
            ~/.werf by default)
  -i, --images='':
            Docker Repo to store images (use WERF_IMAGES_REPO environment by default)
  -s, --stages='':
            Docker Repo to store stages or :local for non-distributed build (only :local is 
            supported for now)
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use WERF_TMP environment or system tmp 
            dir by default)
```

