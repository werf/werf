{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Purge project images from images repo

{{ header }} Syntax

```bash
werf images purge [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or 
            ~/.docker (in the order of priority).
            Command needs granted permissions to delete images from the specified images repo.
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for purge
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME environment 
            or ~/.werf)
  -i, --images-repo='':
            Docker Repo to store images (default $WERF_IMAGES_REPO environment)
      --insecure-repo=false:
            Allow usage of insecure docker repos
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP environment or system 
            tmp dir)
```

