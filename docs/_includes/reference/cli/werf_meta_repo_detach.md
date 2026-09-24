{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Remove the per-project meta-repo marker from --repo so werf no longer forces --meta-repo usage. Metadata already stored in the meta-repo is NOT moved back.

{{ header }} Syntax

```shell
werf meta-repo detach [options]
```

{{ header }} Options

```shell
      --allow-includes-update=false
            Allow use includes latest versions (default $WERF_ALLOW_INCLUDES_UPDATE or false)
      --config=""
            Use custom configuration file (default $WERF_CONFIG or werf.yaml in the project         
            directory)
      --config-render-path=""
            Custom path for storing rendered configuration file
      --config-templates-dir=""
            Custom configuration templates directory (default $WERF_CONFIG_TEMPLATES_DIR or .werf   
            in working directory)
      --container-registry-mirror=[]
            (Buildah-only) Use specified mirrors for docker.io
      --debug-templates=false
            Enable debug mode for Go templates (default $WERF_DEBUG_TEMPLATES or false)
      --dev=false
            Enable development mode (default $WERF_DEV).
            The mode allows working with project files without doing redundant commits during       
            debugging and development
      --dev-branch="_werf-dev"
            Set dev git branch name (default $WERF_DEV_BRANCH or "_werf-dev")
      --dev-ignore=[]
            Add rules to ignore tracked and untracked changes in development mode (can specify      
            multiple).
            Also, can be specified with $WERF_DEV_IGNORE_* (e.g. $WERF_DEV_IGNORE_TESTS=*_test.go,  
            $WERF_DEV_IGNORE_DOCS=path/to/docs)
      --dir=""
            Use specified project directory where project’s werf.yaml and other configuration files 
            should reside (default $WERF_DIR or current working directory)
      --docker-config=""
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
            Command needs granted permissions to read and write images to --repo
      --env=""
            Use specified environment (default $WERF_ENV)
      --git-work-tree=""
            Use specified git work tree dir (default $WERF_WORK_TREE or lookup for directory that   
            contains .git in the current or parent directories)
      --giterminism-config=""
            Custom path to the giterminism configuration file relative to working directory         
            (default $WERF_GITERMINISM_CONFIG or werf-giterminism.yaml in working directory)
      --home-dir=""
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-registry=false
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
      --log-color-mode="auto"
            Set log color mode.
            Supported on, off and auto (based on the stdout’s file descriptor referring to a        
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-debug=false
            Enable debug (default $WERF_LOG_DEBUG).
      --log-pretty=true
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or   
            true).
      --log-project-dir=false
            Print current project directory path (default $WERF_LOG_PROJECT_DIR)
      --log-quiet=false
            Disable explanatory output (default $WERF_LOG_QUIET).
      --log-terminal-width=-1
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --log-time=false
            Add time to log entries for precise event time tracking (default $WERF_LOG_TIME or      
            false).
      --log-time-format="2006-01-02T15:04:05Z07:00"
            Specify custom log time format (default $WERF_LOG_TIME_FORMAT or RFC3339 format).
      --log-verbose=false
            Enable verbose output (default $WERF_LOG_VERBOSE).
      --loose-giterminism=false
            Loose werf giterminism mode restrictions
      --platform=[]
            Enable platform emulation when building images with werf, format: OS/ARCH[/VARIANT]     
            ($WERF_PLATFORM or $DOCKER_DEFAULT_PLATFORM by default)
      --repo=""
            Container registry storage address (default $WERF_REPO)
      --repo-container-registry=""
            Choose repo container registry implementation.
            The following container registries are supported: ecr, acr, default, dockerhub, gcr,    
            github, gitlab, harbor, quay.
            Default $WERF_REPO_CONTAINER_REGISTRY or auto mode (detect container registry by repo   
            address).
      --repo-docker-hub-password=""
            repo Docker Hub password (default $WERF_REPO_DOCKER_HUB_PASSWORD)
      --repo-docker-hub-token=""
            repo Docker Hub token (default $WERF_REPO_DOCKER_HUB_TOKEN)
      --repo-docker-hub-username=""
            repo Docker Hub username (default $WERF_REPO_DOCKER_HUB_USERNAME)
      --repo-github-token=""
            repo GitHub token (default $WERF_REPO_GITHUB_TOKEN)
      --repo-harbor-password=""
            repo Harbor password (default $WERF_REPO_HARBOR_PASSWORD)
      --repo-harbor-username=""
            repo Harbor username (default $WERF_REPO_HARBOR_USERNAME)
      --repo-quay-token=""
            repo quay.io token (default $WERF_REPO_QUAY_TOKEN)
      --skip-tls-verify-registry=false
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --ssh-key=[]
            Use only specific ssh key(s).
            Can be specified with $WERF_SSH_KEY_* (e.g. $WERF_SSH_KEY_REPO=~/.ssh/repo_rsa,         
            $WERF_SSH_KEY_NODEJS=~/.ssh/nodejs_rsa).
            Defaults to $WERF_SSH_KEY_*, system ssh-agent or ~/.ssh/{id_rsa|id_dsa}
      --tmp-dir=""
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

