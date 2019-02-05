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
  $ werf run application

  # Run image with predefined docker run options and command for debug
  $ werf run --shell

  # Run image with specified docker run options and command
  $ werf run --docker-options="-d -p 5000:5000 --restart=always --name registry" -- /app/run.sh

  # Print a resulting docker run command
  $ werf run --shell --dry-run
  docker run -ti --rm image-stage-test:1ffe83860127e68e893b6aece5b0b7619f903f8492a285c6410371c87018c6a0 /bin/sh
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-options='':
            Define docker run options
      --dry-run=false:
            Indicate what the command would do without actually doing that
  -h, --help=false:
            help for run
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or 
            ~/.werf by default)
      --shell=false:
            Use predefined docker options and command for debug
      --ssh-key=[]:
            Enable only specified ssh keys (use system ssh-agent by default)
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

