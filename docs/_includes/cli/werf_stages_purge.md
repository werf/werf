{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Purge project stages from stages storage

{{ header }} Syntax

```bash
werf stages purge [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. WERF_DOCKER_CONFIG or DOCKER_CONFIG or ~/.docker 
            will be used by default (in the order of priority).
            Command needs granted permissions to read, pull and delete images from the specified 
            stages storage.
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for purge
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or 
            ~/.werf by default)
      --insecure-repo=false:
            Allow usage of insecure docker repos
  -s, --stages-storage='':
            Docker Repo to store stages or :local for non-distributed build (only :local is 
            supported for now; use WERF_STAGES_STORAGE environment by default).
            More info about stages: https://flant.github.io/werf/reference/build/stages.html
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use WERF_TMP environment or system tmp 
            dir by default)
```

