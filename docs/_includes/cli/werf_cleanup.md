{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Safely cleanup unused project images and stages.

First step is 'werf images cleanup' command, which will delete unused images from images repo.      
Second step is 'werf stages cleanup' command, which will delete unused stages from stages storage   
to be in sync with the images repo.

It is safe to run this command periodically (daily is enough) by automated cleanup job in parallel  
with other werf commands such as build, deploy and host cleanup.

{{ header }} Syntax

```shell
werf cleanup [options]
```

{{ header }} Examples

```shell
  $ werf cleanup --stages-storage :local --images-repo registry.mydomain.com/myproject
```

{{ header }} Options

```shell
      --dir='':
            Change to the specified directory to find werf.yaml config
      --docker-config='':
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
            Command needs granted permissions to read, pull and delete images from the specified    
            stages storage and images repo
      --dry-run=false:
            Indicate what the command would do without actually doing that
      --git-commit-strategy-expiry-days=-1:
            Keep images published with the git-commit tagging strategy in the images repo for the   
            specified maximum days since image published. Republished image will be kept specified  
            maximum days since new publication date. No days limit by default, -1 disables the      
            limit. Value can be specified by the $WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS
      --git-commit-strategy-limit=-1:
            Keep max number of images published with the git-commit tagging strategy in the images  
            repo. No limit by default, -1 disables the limit. Value can be specified by the         
            $WERF_GIT_COMMIT_STRATEGY_LIMIT
      --git-tag-strategy-expiry-days=-1:
            Keep images published with the git-tag tagging strategy in the images repo for the      
            specified maximum days since image published. Republished image will be kept specified  
            maximum days since new publication date. No days limit by default, -1 disables the      
            limit. Value can be specified by the $WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS
      --git-tag-strategy-limit=-1:
            Keep max number of images published with the git-tag tagging strategy in the images     
            repo. No limit by default, -1 disables the limit. Value can be specified by the         
            $WERF_GIT_TAG_STRATEGY_LIMIT
  -h, --help=false:
            help for cleanup
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
  -i, --images-repo='':
            Docker Repo to store images (default $WERF_IMAGES_REPO)
      --images-repo-mode='multirepo':
            Define how to store images in Repo: multirepo or monorepo (defaults to                  
            $WERF_IMAGES_REPO_MODE or multirepo)
      --insecure-registry=false:
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
      --kube-config='':
            Kubernetes config file path
      --kube-context='':
            Kubernetes config context (default $WERF_KUBE_CONTEXT)
      --log-color-mode='auto':
            Set log color mode.
            Supported on, off and auto (based on the stdout’s file descriptor referring to a        
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
      --skip-tls-verify-registry=false:
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
  -s, --stages-storage='':
            Docker Repo to store stages or :local for non-distributed build (only :local is         
            supported for now; default $WERF_STAGES_STORAGE environment).
            More info about stages: https://werf.io/documentation/reference/stages_and_images.html
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --without-kube=false:
            Do not skip deployed Kubernetes images (default $WERF_KUBE_CONTEXT)
```

