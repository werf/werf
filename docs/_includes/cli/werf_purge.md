{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Purge all project images from images repo and stages from stages storage.

First step is 'werf images purge', which will delete all project images from images repo. Second 
step is 'werf stages purge', which will delete all stages from stages storage.

WARNING: Do not run this command during any other werf command is working on the host machine. This 
command is supposed to be run manually. Images from images repo, that are being used in Kubernetes 
cluster will also be deleted.

{{ header }} Syntax

```bash
werf purge [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --disable-pretty-log=false:
            Disable emojis, auto line wrapping and replace log process border characters with 
            spaces (default $WERF_DISABLE_PRETTY_LOG).
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or 
            ~/.docker (in the order of priority).
            Command needs granted permissions to delete images from the specified stages storage 
            and images repo.
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for purge
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
  -i, --images-repo='':
            Docker Repo to store images (default $WERF_IMAGES_REPO)
      --insecure-repo=false:
            Allow usage of insecure docker repos
      --log-color-mode='auto':
            Set log color mode.
            Supported on, off and auto (based on the stdout's file descriptor referring to a 
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
  -s, --stages-storage='':
            Docker Repo to store stages or :local for non-distributed build (only :local is 
            supported for now; default $WERF_STAGES_STORAGE environment).
            More info about stages: https://werf.io/reference/build/stages.html
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP or system tmp dir)
```

