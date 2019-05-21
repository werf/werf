{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Run container for specified project image

{{ header }} Syntax

```bash
werf run [options] [IMAGE_NAME] [-- COMMAND ARG...]
```

{{ header }} Examples

```bash
  # Run specified image
  $ werf --stages-storage :local run application

  # Run image with predefined docker run options and command for debug
  $ werf --stages-storage :local run --shell

  # Run image with specified docker run options and command
  $ werf --stages-storage :local run --docker-options="-d -p 5000:5000 --restart=always --name registry" -- /app/run.sh

  # Print a resulting docker run command
  $ werf --stages-storage :local run --shell --dry-run
  docker run -ti --rm image-stage-test:1ffe83860127e68e893b6aece5b0b7619f903f8492a285c6410371c87018c6a0 /bin/sh
```

{{ header }} Options

```bash
      --bash=false:
            Use predefined docker options and command for debug
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or 
            ~/.docker (in the order of priority).
            Command needs granted permissions to read and pull images from the specified stages 
            storage.
      --docker-options='':
            Define docker run options
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for run
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-repo=false:
            Allow usage of insecure docker repos (default $WERF_INSECURE_REPO)
      --log-color-mode='auto':
            Set log color mode.
            Supported on, off and auto (based on the stdout's file descriptor referring to a 
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-pretty=true:
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or 
            true).
      --log-project-dir=false:
            Print current project directory path (default $WERF_LOG_PROJECT_DIR)
      --log-terminal-width=-1:
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --shell=false:
            Use predefined docker options and command for debug
      --ssh-key=[]:
            Use only specific ssh keys (Defaults to system ssh-agent or ~/.ssh/{id_rsa|id_dsa}, see 
            https://werf.io/reference/toolbox/ssh.html). Option can be specified multiple times to 
            use multiple keys.
  -s, --stages-storage='':
            Docker Repo to store stages or :local for non-distributed build (only :local is 
            supported for now; default $WERF_STAGES_STORAGE environment).
            More info about stages: https://werf.io/reference/build/stages_and_images.html
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

