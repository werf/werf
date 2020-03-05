{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
remove image record from the list of managed images which will be preserved during cleanup procedure

{{ header }} Syntax

```shell
werf managed-images rm [options]
```

{{ header }} Options

```shell
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
            Command needs granted permissions to read and write images to the specified stages      
            storage
  -h, --help=false:
            help for rm
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
  -i, --images-repo='':
            Docker Repo to store images (default $WERF_IMAGES_REPO)
      --images-repo-mode='multirepo':
            Define how to store images in Repo: multirepo or monorepo (defaults to                  
            $WERF_IMAGES_REPO_MODE or multirepo)
      --insecure-registry=false:
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
      --log-color-mode='auto':
            Set log color mode.
            Supported on, off and auto (based on the stdout’s file descriptor referring to a        
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-debug=false:
            Enable debug (default $WERF_LOG_DEBUG).
      --log-pretty=true:
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or   
            true).
      --log-project-dir=false:
            Print current project directory path (default $WERF_LOG_PROJECT_DIR)
      --log-quiet=false:
            Disable explanatory output (default $WERF_LOG_QUIET).
      --log-terminal-width=-1:
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --log-verbose=false:
            Enable verbose output (default $WERF_LOG_VERBOSE).
  -N, --project-name='':
            Use specified project name (default $WERF_PROJECT_NAME)
      --skip-tls-verify-registry=false:
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --ssh-key=[]:
            Use only specific ssh keys (Defaults to system ssh-agent or ~/.ssh/{id_rsa|id_dsa}, see 
            https://werf.io/documentation/reference/toolbox/ssh.html).
            Option can be specified multiple times to use multiple keys
  -s, --stages-storage='':
            Docker Repo to store stages or :local for non-distributed build (only :local is         
            supported for now; default $WERF_STAGES_STORAGE environment).
            More info about stages: https://werf.io/documentation/reference/stages_and_images.html
      --synchronization=':local':
            Address of synchronizer for multiple werf processes to work with a single stages        
            storage (default :local or $WERF_SYNCHRONIZATION if set). The same address should be    
            specified for all werf processes that work with a single stages storage. :local address 
            allows execution of werf processes from a single host only.
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

