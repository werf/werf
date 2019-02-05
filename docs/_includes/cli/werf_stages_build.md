{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Build stages for images described in the werf.yaml.

The result of build command are built stages pushed into the specified stages repo or locally if 
--stages=:local.

If one or more IMAGE_NAME parameters specified, werf will build only these images stages from 
werf.yaml

{{ header }} Syntax

```bash
werf stages build [IMAGE_NAME...] [options]
```

{{ header }} Environments

```bash
  $WERF_ANSIBLE_ARGS   Pass specified cli args to ansible (ANSIBLE_ARGS)
  $WERF_DOCKER_CONFIG  Force usage of the specified docker config
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
  -h, --help=false:
            help for build
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use WERF_HOME environment or 
            ~/.werf by default)
      --introspect-before-error=false:
            Introspect failed stage in the clean state, before running all assembly instructions 
            of the stage
      --introspect-error=false:
            Introspect failed stage in the state, right after running failed assembly instruction
      --pull-password='':
            Password to authorize pull of base images
      --pull-username='':
            Username to authorize pull of base images
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

