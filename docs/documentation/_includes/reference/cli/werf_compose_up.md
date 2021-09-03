{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Run docker-compose up command with forwarded image names

Image environment name format: $WERF_&lt;FORMATTED_WERF_IMAGE_NAME&gt;_DOCKER_IMAGE_NAME                  
($WERF_DOCKER_IMAGE_NAME for nameless image).
&lt;FORMATTED_WERF_IMAGE_NAME&gt; is werf image name from werf.yaml modified according to the following   
rules:
- all characters are uppercase (app -&gt; APP);
- charset /- is replaced with _ (DEV/APP-FRONTEND -&gt; DEV_APP_FRONTEND).

If one or more IMAGE_NAME parameters specified, werf will build and forward only these images.

Given the following werf configuration:

# werf.yaml
project: x
configVersion: 1
---
image: frontend
dockerfile: frontend.Dockerfile
---
image: geodata-backend
dockerfile: backend.Dockerfile

Use described images as follows in your docker compose configuration:

# docker-compose.yaml
services:
  frontend:
    image: $WERF_FRONTEND_DOCKER_IMAGE_NAME
  backend:
    image: $WERF_GEODATA_BACKEND_DOCKER_IMAGE_NAME


{{ header }} Syntax

```shell
werf compose up [IMAGE_NAME...] [options] [--docker-compose-options="OPTIONS"] [--docker-compose-command-options="OPTIONS"] [--] [SERVICE...]
```

{{ header }} Examples

```shell
  # Run docker-compose up with forwarded image names
  $ werf compose up

  # Follow git HEAD and run docker-compose up for each new commit
  $ werf compose up --follow --docker-compose-command-options="-d"

  # Print docker-compose command without executing
  $ werf compose up --docker-compose-options="-f docker-compose-test.yml" --docker-compose-command-options="--abort-on-container-exit -t 20" --dry-run --quiet
  export WERF_APP_DOCKER_IMAGE_NAME=localhost:5000/project:570c59946a7f77873d361efd25a637c4ccde86abf3d3186add19bded-1604928781528
  docker-compose -f docker-compose-test.yml up --abort-on-container-exit -t 20
```

{{ header }} Options

```shell
      --allowed-docker-storage-volume-usage=70
            Set allowed percentage of docker storage volume usage which will cause cleanup of least 
            recently used local docker images (default 70% or                                       
            $WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE)
      --allowed-docker-storage-volume-usage-margin=5
            During cleanup of least recently used local docker images werf would delete images      
            until volume usage becomes below "allowed-docker-storage-volume-usage -                 
            allowed-docker-storage-volume-usage-margin" level (default 5% or                        
            $WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE_MARGIN)
      --allowed-local-cache-volume-usage=70
            Set allowed percentage of local cache (~/.werf/local_cache by default) volume usage     
            which will cause cleanup of least recently used data from the local cache (default 70%  
            or $WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE)
      --allowed-local-cache-volume-usage-margin=5
            During cleanup of least recently used local docker images werf would delete images      
            until volume usage becomes below "allowed-docker-storage-volume-usage -                 
            allowed-docker-storage-volume-usage-margin" level (default 5% or                        
            $WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN)
      --cache-repo=[]
            Specify one or multiple cache repos with images that will be used as a cache. Cache     
            will be populated when pushing newly built images into the primary repo and when        
            pulling existing images from the primary repo. Cache repo will be used to pull images   
            and to get manifests before making requests to the primary repo.
            Also, can be specified with $WERF_CACHE_REPO_* (e.g. $WERF_CACHE_REPO_1=...,            
            $WERF_CACHE_REPO_2=...)
      --config=''
            Use custom configuration file (default $WERF_CONFIG or werf.yaml in working directory)
      --config-templates-dir=''
            Custom configuration templates directory (default $WERF_CONFIG_TEMPLATES_DIR or .werf   
            in working directory)
      --dev=false
            Enable development mode (default $WERF_DEV).
            The mode allows working with project files without doing redundant commits during       
            debugging and development
      --dev-branch-prefix='werf-dev-'
            Set dev git branch prefix (default $WERF_DEV_BRANCH_PREFIX or werf-dev-)
      --dev-ignore=[]
            Add rules to ignore tracked and untracked changes in development mode (can specify      
            multiple).
            Also, can be specified with $WERF_DEV_IGNORE_* (e.g. $WERF_DEV_IGNORE_TESTS=*_test.go,  
            $WERF_DEV_IGNORE_DOCS=path/to/docs)
      --dir=''
            Use specified project directory where project’s werf.yaml and other configuration files 
            should reside (default $WERF_DIR or current working directory)
      --disable-auto-host-cleanup=false
            Disable auto host cleanup procedure in main werf commands like werf-build,              
            werf-converge and other (default disabled or WERF_DISABLE_AUTO_HOST_CLEANUP)
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
      --docker-server-storage-path=''
            Use specified path to the local docker server storage to check docker storage volume    
            usage while performing garbage collection of local docker images (detect local docker   
            server storage path by default or use $WERF_DOCKER_SERVER_STORAGE_PATH)
      --dry-run=false
            Indicate what the command would do without actually doing that (default $WERF_DRY_RUN)
      --env=''
            Use specified environment (default $WERF_ENV)
      --final-repo=''
            Docker Repo to store only those stages which are going to be used by the Kubernetes     
            cluster, in other word final images (default $WERF_FINAL_REPO)
      --final-repo-container-registry=''
            Choose repo container registry for .
            The following container registries are supported: ecr, acr, default, dockerhub, gcr,    
            github, gitlab, harbor, quay.
            Default $WERF_FINAL_REPO_CONTAINER_REGISTRY or auto mode (detect container registry by  
            repo address).
      --final-repo-docker-hub-password=''
            Docker Hub password for  (default $WERF_FINAL_REPO_DOCKER_HUB_PASSWORD)
      --final-repo-docker-hub-token=''
            Docker Hub token for  (default $WERF_FINAL_REPO_DOCKER_HUB_TOKEN)
      --final-repo-docker-hub-username=''
            Docker Hub username for  (default $WERF_FINAL_REPO_DOCKER_HUB_USERNAME)
      --final-repo-github-token=''
            GitHub token for  (default $WERF_FINAL_REPO_GITHUB_TOKEN)
      --final-repo-harbor-password=''
            Harbor password for  (default $WERF_FINAL_REPO_HARBOR_PASSWORD)
      --final-repo-harbor-username=''
            Harbor username for  (default $WERF_FINAL_REPO_HARBOR_USERNAME)
      --final-repo-quay-token=''
            quay.io token for  (default $WERF_FINAL_REPO_QUAY_TOKEN)
      --follow=false
            Enable follow mode (default $WERF_FOLLOW).
            The mode allows restarting the command on a new commit.
            In development mode (--dev), werf restarts the command on any changes (including        
            untracked files) in the git repository worktree
      --git-work-tree=''
            Use specified git work tree dir (default $WERF_WORK_TREE or lookup for directory that   
            contains .git in the current or parent directories)
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-registry=false
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
      --kube-config=''
            Kubernetes config file path (default $WERF_KUBE_CONFIG, or $WERF_KUBECONFIG, or         
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
      --platform=''
            Enable platform emulation when building images with werf. The only supported option for 
            now is linux/amd64.
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

