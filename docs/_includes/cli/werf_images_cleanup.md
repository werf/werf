{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Cleanup project images from images repo

{{ header }} Syntax

```shell
werf images cleanup [options]
```

{{ header }} Options

```shell
      --allow-git-shallow-clone=false
            Sign the intention of using shallow clone despite restrictions (default                 
            $WERF_ALLOW_GIT_SHALLOW_CLONE)
      --config=''
            Use custom configuration file (default $WERF_CONFIG or werf.yaml in working directory)
      --config-templates-dir=''
            Change to the custom configuration templates directory (default                         
            $WERF_CONFIG_TEMPLATES_DIR or .werf in working directory)
      --dir=''
            Use custom working directory (default $WERF_DIR or current directory)
      --docker-config=''
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
            Command needs granted permissions to delete images from the specified images repo
      --dry-run=false
            Indicate what the command would do without actually doing that (default $WERF_DRY_RUN)
      --git-commit-strategy-expiry-days=-1
            Keep images published with the git-commit tagging strategy in the images repo for the   
            specified maximum days since image published. Republished image will be kept specified  
            maximum days since new publication date. No days limit by default, -1 disables the      
            limit. Value can be specified by the $WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS
      --git-commit-strategy-limit=-1
            Keep max number of images published with the git-commit tagging strategy in the images  
            repo. No limit by default, -1 disables the limit. Value can be specified by the         
            $WERF_GIT_COMMIT_STRATEGY_LIMIT
      --git-history-based-cleanup=true
            Use git history based cleanup (default $WERF_GIT_HISTORY_BASED_CLEANUP)
      --git-history-based-cleanup-v1.2=false
            Use git history based cleanup and delete images tags without related image metadata     
            (default $WERF_GIT_HISTORY_BASED_CLEANUP_v1_2)
      --git-history-synchronization=false
            Synchronize git branches and tags with remote origin (default                           
            $WERF_GIT_HISTORY_SYNCHRONIZATION)
      --git-tag-strategy-expiry-days=-1
            Keep images published with the git-tag tagging strategy in the images repo for the      
            specified maximum days since image published. Republished image will be kept specified  
            maximum days since new publication date. No days limit by default, -1 disables the      
            limit. Value can be specified by the $WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS
      --git-tag-strategy-limit=-1
            Keep max number of images published with the git-tag tagging strategy in the images     
            repo. No limit by default, -1 disables the limit. Value can be specified by the         
            $WERF_GIT_TAG_STRATEGY_LIMIT
  -h, --help=false
            help for cleanup
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
  -i, --images-repo=''
            Docker Repo to store images (default $WERF_IMAGES_REPO)
      --images-repo-docker-hub-password=''
            Docker Hub password for images repo (default $WERF_IMAGES_REPO_DOCKER_HUB_PASSWORD,     
            $WERF_REPO_DOCKER_HUB_PASSWORD)
      --images-repo-docker-hub-token=''
            Docker Hub token for images repo (default $WERF_IMAGES_REPO_DOCKER_HUB_TOKEN,           
            $WERF_REPO_DOCKER_HUB_TOKEN)
      --images-repo-docker-hub-username=''
            Docker Hub username for images repo (default $WERF_IMAGES_REPO_DOCKER_HUB_USERNAME,     
            $WERF_REPO_DOCKER_HUB_USERNAME)
      --images-repo-github-token=''
            GitHub token for images repo (default $WERF_IMAGES_REPO_GITHUB_TOKEN,                   
            $WERF_REPO_GITHUB_TOKEN)
      --images-repo-implementation=''
            Choose repo implementation for images repo.
            The following docker registry implementations are supported: ecr, acr, default,         
            dockerhub, gcr, github, gitlab, harbor, quay.
            Default $WERF_IMAGES_REPO_IMPLEMENTATION, $WERF_REPO_IMPLEMENTATION or auto mode        
            (detect implementation by a registry).
      --images-repo-mode='auto'
            Define how to store in images repo: multirepo or monorepo.
            Default $WERF_IMAGES_REPO_MODE or auto mode
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
  -p, --parallel=true
            Run in parallel (default $WERF_PARALLEL)
      --parallel-tasks-limit=10
            Parallel tasks limit, set -1 to remove the limitation (default                          
            $WERF_PARALLEL_TASKS_LIMIT or 5)
      --repo-docker-hub-password=''
            Common Docker Hub password for any stages storage or images repo specified for the      
            command (default $WERF_REPO_DOCKER_HUB_PASSWORD)
      --repo-docker-hub-token=''
            Common Docker Hub token for any stages storage or images repo specified for the command 
            (default $WERF_REPO_DOCKER_HUB_TOKEN)
      --repo-docker-hub-username=''
            Common Docker Hub username for any stages storage or images repo specified for the      
            command (default $WERF_REPO_DOCKER_HUB_USERNAME)
      --repo-github-token=''
            Common GitHub token for any stages storage or images repo specified for the command     
            (default $WERF_REPO_GITHUB_TOKEN)
      --repo-implementation=''
            Choose common repo implementation for any stages storage or images repo specified for   
            the command.
            The following docker registry implementations are supported: ecr, acr, default,         
            dockerhub, gcr, github, gitlab, harbor, quay.
            Default $WERF_REPO_IMPLEMENTATION or auto mode (detect implementation by a registry).
      --scan-context-namespace-only=false
            Scan for used images only in namespace linked with context for each available context   
            in kube-config (or only for the context specified with option --kube-context). When     
            disabled will scan all namespaces in all contexts (or only for the context specified    
            with option --kube-context). (Default $WERF_SCAN_CONTEXT_NAMESPACE_ONLY)
      --skip-tls-verify-registry=false
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
  -s, --stages-storage=''
            Docker Repo to store stages or :local for non-distributed build (only :local is         
            supported for now; default $WERF_STAGES_STORAGE environment)
      --stages-storage-repo-docker-hub-password=''
            Docker Hub password for stages storage (default                                         
            $WERF_STAGES_STORAGE_REPO_DOCKER_HUB_PASSWORD, $WERF_REPO_DOCKER_HUB_PASSWORD)
      --stages-storage-repo-docker-hub-token=''
            Docker Hub token for stages storage (default                                            
            $WERF_STAGES_STORAGE_REPO_DOCKER_HUB_TOKEN, $WERF_REPO_DOCKER_HUB_TOKEN)
      --stages-storage-repo-docker-hub-username=''
            Docker Hub username for stages storage (default                                         
            $WERF_STAGES_STORAGE_REPO_DOCKER_HUB_USERNAME, $WERF_REPO_DOCKER_HUB_USERNAME)
      --stages-storage-repo-github-token=''
            GitHub token for stages storage (default $WERF_STAGES_STORAGE_REPO_GITHUB_TOKEN,        
            $WERF_REPO_GITHUB_TOKEN)
      --stages-storage-repo-implementation=''
            Choose repo implementation for stages storage.
            The following docker registry implementations are supported: ecr, acr, default,         
            dockerhub, gcr, github, gitlab, harbor, quay.
            Default $WERF_STAGES_STORAGE_REPO_IMPLEMENTATION, $WERF_REPO_IMPLEMENTATION or auto     
            mode (detect implementation by a registry).
  -S, --synchronization=''
            Address of synchronizer for multiple werf processes to work with a single stages        
            storage (default :local if --stages-storage=:local or kubernetes://werf-synchronization 
            if non-local stages-storage specified or $WERF_SYNCHRONIZATION if set). The same        
            address should be specified for all werf processes that work with a single stages       
            storage. :local address allows execution of werf processes from a single host only.
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --without-kube=false
            Do not skip deployed Kubernetes images (default $WERF_WITHOUT_KUBE)
```

