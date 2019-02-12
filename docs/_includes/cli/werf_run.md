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
            Specify docker config directory path. WERF_DOCKER_CONFIG or DOCKER_CONFIG or ~/.docker 
            will be used by default (in the order of priority).
            Command needs granted permissions to read and pull images from the specified stages 
            storage.
      --docker-options='':
            Define docker run options
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for run
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or 
            ~/.werf by default)
      --insecure-repo=false:
            Allow usage of insecure docker repos
      --shell=false:
            Use predefined docker options and command for debug
      --ssh-key=[]:
            Use only specific ssh keys (system ssh-agent or default keys will be used by default, 
            see https://flant.github.io/werf/reference/toolbox/ssh.html). Option can be specified 
            multiple times to use multiple keys.
  -s, --stages-storage='':
            Docker Repo to store stages or :local for non-distributed build (only :local is 
            supported for now; use WERF_STAGES_STORAGE environment by default).
            More info about stages: https://flant.github.io/werf/reference/build/stages.html
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use WERF_TMP environment or system tmp 
            dir by default)
```

