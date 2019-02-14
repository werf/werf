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

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. Default WERF_DOCKER_CONFIG or DOCKER_CONFIG or 
            ~/.docker (in the order of priority).
            Command needs granted permissions to read, pull and delete images from the specified 
            stages storage, read images from the specified images repo.
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for cleanup
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default WERF_HOME environment or 
            ~/.werf)
  -i, --images-repo='':
            Docker Repo to store images (default WERF_IMAGES_REPO environment)
      --insecure-repo=false:
            Allow usage of insecure docker repos
  -s, --stages-storage=':local':
            Docker Repo to store stages or :local for non-distributed build (only :local is 
            supported for now; default WERF_STAGES_STORAGE environment).
            More info about stages: https://flant.github.io/werf/reference/build/stages.html
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default WERF_TMP environment or system 
            tmp dir)
```

