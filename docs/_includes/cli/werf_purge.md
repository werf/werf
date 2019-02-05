{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Purge all project images from images repo and stages from stages repo (or locally).

First step is 'werf images purge', which will delete all project images from images repo. Second 
step is 'werf stages purge', which will delete all stages from stages repo (or locally).

Command allows deletion of all images of the project at once, meant to be used manually.

WARNING: Images from images repo, that are being used in Kubernetes cluster will also be deleted.

{{ header }} Syntax

```bash
werf purge [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for purge
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or 
            ~/.werf by default)
  -i, --images='':
            Docker Repo to store images (use WERF_IMAGES_REPO environment by default)
  -p, --images-password='':
            Images Docker repo username (granted permission to read images info and delete images, 
            use WERF_IMAGES_PASSWORD environment by default)
  -u, --images-username='':
            Images Docker repo username (granted permission to read images info and delete images, 
            use WERF_IMAGES_USERNAME environment by default)
  -s, --stages='':
            Docker Repo to store stages or :local for non-distributed build (only :local is 
            supported for now)
      --stages-password='':
            Stages Docker repo password
      --stages-username='':
            Stages Docker repo username
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use WERF_TMP environment or system tmp 
            dir by default)
```

