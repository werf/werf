{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Run docker-compose run command with forwarded image names

Image environment name format: $WERF_&lt;FORMATTED_WERF_IMAGE_NAME&gt;_DOCKER_IMAGE_NAME                  
($WERF_DOCKER_IMAGE_NAME for nameless image).
&lt;FORMATTED_WERF_IMAGE_NAME&gt; is werf image name from werf.yaml modified according to the following   
rules:
- all characters are uppercase (app -&gt; APP);
- charset /- is replaced with _ (DEV/APP-FRONTEND -&gt; DEV_APP_FRONTEND).

If one or more IMAGE_NAME parameters specified, werf will build and forward only these images

{{ header }} Syntax

```shell
werf compose run [IMAGE_NAME...] [options] [--docker-compose-options="OPTIONS"] [--docker-compose-command-options="OPTIONS"] -- SERVICE [COMMAND] [ARGS...]
```

{{ header }} Examples

```shell
  # Print docker-compose command without executing
  $ werf compose run --docker-compose-options="-f docker-compose-test.yml" --docker-compose-command-options="-e TOKEN=123" --dry-run --quiet -- test
  export WERF_TEST_DOCKER_IMAGE_NAME=test:03dc2e0bceb09833f54fcab39e89e6e4137316ebbe544aeec8184420-1620123105753
  docker-compose -f docker-compose-test.yml up -e TOKEN=123 -- test
```

{{ header }} Options

```shell
      --config=''
            Use custom configuration file (default $WERF_CONFIG or werf.yaml in working directory)
      --config-templates-dir=''
            Custom configuration templates directory (default $WERF_CONFIG_TEMPLATES_DIR or .werf   
            in working directory)
      --dev=false
            Enable development mode (default $WERF_DEV).
            The mode allows working with project files without doing redundant commits during       
            debugging and development
      --dev-mode='simple'
            Set development mode (default $WERF_DEV_MODE or simple).
            Two development modes are supported:
            - simple: for working with the worktree state of the git repository
            - strict: for working with the index state of the git repository
      --dir=''
            Use specified project directory where project’s werf.yaml and other configuration files 
            should reside (default $WERF_DIR or current working directory)
      --docker-compose-bin-path=''
            Define docker-compose bin path (default $WERF_DOCKER_COMPOSE_BIN_PATH)
      --docker-compose-command-options=''
            Define docker-compose command options (default $WERF_DOCKER_COMPOSE_COMMAND_OPTIONS)
      --docker-compose-options=''
            Define docker-compose options (default $WERF_DOCKER_COMPOSE_OPTIONS)
      --docker-config=''
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
            Command needs granted permissions to read and pull images from the specified repo
      --dry-run=false
            Indicate what the command would do without actually doing that (default $WERF_DRY_RUN)
      --env=''
            Use specified environment (default $WERF_ENV)
      --git-work-tree=''
            Use specified git work tree dir (default $WERF_WORK_TREE or lookup for directory that   
            contains .git in the current or parent directories)
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-registry=false
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
      --kube-config=''
            Kubernetes config file path (default $WERF_KUBE_CONFIG or $WERF_KUBECONFIG or           
            $KUBECONFIG)
      --kube-config-base64=''
            Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or            
            $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)
      --kube-context=''
            Kubernetes config context (default $WERF_KUBE_CONTEXT)
      --log-color-mode='auto'
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
      --log-verbose=false
            Enable verbose output (default $WERF_LOG_VERBOSE).
      --loose-giterminism=false
            Loose werf giterminism mode restrictions (NOTE: not all restrictions can be removed,    
            more info https://werf.io/documentation/advanced/giterminism.html, default              
            $WERF_LOOSE_GITERMINISM)
      --repo=''
            Docker Repo to store stages (default $WERF_REPO)
      --repo-container-registry=''
            Choose repo container registry.
            The following container registries are supported: ecr, acr, default, dockerhub, gcr,    
            github, gitlab, harbor, quay.
            Default $WERF_REPO_CONTAINER_REGISTRY or auto mode (detect container registry by repo   
            address).
      --repo-docker-hub-password=''
            Docker Hub password (default $WERF_REPO_DOCKER_HUB_PASSWORD)
      --repo-docker-hub-token=''
            Docker Hub token (default $WERF_REPO_DOCKER_HUB_TOKEN)
      --repo-docker-hub-username=''
            Docker Hub username (default $WERF_REPO_DOCKER_HUB_USERNAME)
      --repo-github-token=''
            GitHub token (default $WERF_REPO_GITHUB_TOKEN)
      --repo-harbor-password=''
            Harbor password (default $WERF_REPO_HARBOR_PASSWORD)
      --repo-harbor-username=''
            Harbor username (default $WERF_REPO_HARBOR_USERNAME)
      --repo-quay-token=''
            quay.io token (default $WERF_REPO_QUAY_TOKEN)
      --secondary-repo=[]
            Specify one or multiple secondary read-only repos with images that will be used as a    
            cache.
            Also, can be specified with $WERF_SECONDARY_REPO_* (e.g. $WERF_SECONDARY_REPO_1=...,    
            $WERF_SECONDARY_REPO_2=...)
  -Z, --skip-build=false
            Disable building of docker images, cached images in the repo should exist in the repo   
            if werf.yaml contains at least one image description (default $WERF_SKIP_BUILD)
      --skip-tls-verify-registry=false
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --ssh-key=[]
            Use only specific ssh key(s).
            Can be specified with $WERF_SSH_KEY_* (e.g. $WERF_SSH_KEY_REPO=~/.ssh/repo_rsa,         
            $WERF_SSH_KEY_NODEJS=~/.ssh/nodejs_rsa).
            Defaults to $WERF_SSH_KEY_*, system ssh-agent or ~/.ssh/{id_rsa|id_dsa}, see            
            https://werf.io/documentation/reference/toolbox/ssh.html
  -S, --synchronization=''
            Address of synchronizer for multiple werf processes to work with a single repo.
            
            Default:
             - $WERF_SYNCHRONIZATION, or
             - :local if --repo is not specified, or
             - https://synchronization.werf.io if --repo has been specified.
            
            The same address should be specified for all werf processes that work with a single     
            repo. :local address allows execution of werf processes from a single host only
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --virtual-merge=false
            Enable virtual/ephemeral merge commit mode when building current application state      
            ($WERF_VIRTUAL_MERGE by default)
      --virtual-merge-from-commit=''
            Commit hash for virtual/ephemeral merge commit with new changes introduced in the pull  
            request ($WERF_VIRTUAL_MERGE_FROM_COMMIT by default)
      --virtual-merge-into-commit=''
            Commit hash for virtual/ephemeral merge commit which is base for changes introduced in  
            the pull request ($WERF_VIRTUAL_MERGE_INTO_COMMIT by default)
```

