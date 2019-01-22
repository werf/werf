{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}


{{ header }} Syntax

```bash
werf build [DIMG_NAME...] [options]
```

{{ header }} Options

```bash
      --dir='': Change to the specified directory to find werf.yaml config
  -h, --help=false: help for build
      --home-dir='': Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --introspect-before-error=false: Introspect failed stage in the clean state, before running all assembly instructions of the stage
      --introspect-error=false: Introspect failed stage in the state, right after running failed assembly instruction
      --pull-password='': Docker registry password to authorize pull of base images
      --pull-username='': Docker registry username to authorize pull of base images
      --registry-password='': Docker registry password to authorize pull of base images
      --registry-username='': Docker registry username to authorize pull of base images
      --ssh-key=[]: Enable only specified ssh keys (use system ssh-agent by default)
      --tmp-dir='': Use specified dir to store tmp files and dirs (use system tmp dir by default)
```

{{ header }} Environments

```bash
  $WERF_ANSIBLE_ARGS                
  $WERF_DOCKER_CONFIG               
  $WERF_IGNORE_CI_DOCKER_AUTOLOGIN  
  $WERF_HOME                        
  $WERF_TMP                         
```

