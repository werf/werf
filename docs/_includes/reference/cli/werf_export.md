{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Export images to an arbitrary repository according to a template specified by the `--tag` option (build if needed).

The tag may contain the following shortcuts:
- `image`, `image_slug` or `image_safe_slug` to use the image name (necessary if there is more than one image in the werf config);
- `image_content_based_tag` to use a content-based tag.
All meta-information related to werf is removed from the exported images, and then images are completely under the user's responsibility.

{{ header }} Syntax

```shell
werf export [IMAGE_NAME...] [options]
```

{{ header }} Examples

```shell
  # Export images to Docker Hub and GitHub container registry
  $ werf export \
      --tag index.docker.io/company/project:%image%-latest \
      --tag ghcr.io/company/project/%image%:latest

  # Export images with extra labels
  $ werf export \
      --tag registry.werf.io/company/project/%image%:latest \
      --add-label io.artifacthub.package.readme-url=https://raw.githubusercontent.com/werf/werf/main/README.md \
      --add-label org.opencontainers.image.created=2023-03-13T11:55:24Z \
      --add-label org.opencontainers.image.description="Official image to run werf in containers"
```

{{ header }} Options

```shell
      --add-label=[]
            Add label to exported images (can specify multiple).
            Format: labelName=labelValue.
            Also, can be specified with $WERF_EXPORT_ADD_LABEL_* (e.g.                              
            $WERF_EXPORT_ADD_LABEL_1=labelName1=labelValue1,                                        
            $WERF_EXPORT_ADD_LABEL_2=labelName2=labelValue2)
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
      --container-registry-mirror=[]
            (Buildah-only) Use specified mirrors for docker.io
      --dev=false
            Enable development mode (default $WERF_DEV).
            The mode allows working with project files without doing redundant commits during       
            debugging and development
      --dev-branch='_werf-dev'
            Set dev git branch name (default $WERF_DEV_BRANCH or "_werf-dev")
      --dev-ignore=[]
            Add rules to ignore tracked and untracked changes in development mode (can specify      
            multiple).
            Also, can be specified with $WERF_DEV_IGNORE_* (e.g. $WERF_DEV_IGNORE_TESTS=*_test.go,  
            $WERF_DEV_IGNORE_DOCS=path/to/docs)
      --dir=''
            Use specified project directory where project’s werf.yaml and other configuration files 
            should reside (default $WERF_DIR or current working directory)
      --docker-config=''
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
            Command needs granted permissions to read and pull images from the specified repo
      --dry-run=false
            Indicate what the command would do without actually doing that (default $WERF_DRY_RUN)
      --env=''
            Use specified environment (default $WERF_ENV)
      --final-repo=''
            Container registry storage address (default $WERF_FINAL_REPO)
      --final-repo-container-registry=''
            Choose final-repo container registry implementation.
            The following container registries are supported: ecr, acr, default, dockerhub, gcr,    
            github, gitlab, harbor, quay.
            Default $WERF_FINAL_REPO_CONTAINER_REGISTRY or auto mode (detect container registry by  
            repo address).
      --final-repo-docker-hub-password=''
            final-repo Docker Hub password (default $WERF_FINAL_REPO_DOCKER_HUB_PASSWORD)
      --final-repo-docker-hub-token=''
            final-repo Docker Hub token (default $WERF_FINAL_REPO_DOCKER_HUB_TOKEN)
      --final-repo-docker-hub-username=''
            final-repo Docker Hub username (default $WERF_FINAL_REPO_DOCKER_HUB_USERNAME)
      --final-repo-github-token=''
            final-repo GitHub token (default $WERF_FINAL_REPO_GITHUB_TOKEN)
      --final-repo-harbor-password=''
            final-repo Harbor password (default $WERF_FINAL_REPO_HARBOR_PASSWORD)
      --final-repo-harbor-username=''
            final-repo Harbor username (default $WERF_FINAL_REPO_HARBOR_USERNAME)
      --final-repo-quay-token=''
            final-repo quay.io token (default $WERF_FINAL_REPO_QUAY_TOKEN)
      --git-work-tree=''
            Use specified git work tree dir (default $WERF_WORK_TREE or lookup for directory that   
            contains .git in the current or parent directories)
      --giterminism-config=''
            Custom path to the giterminism configuration file relative to working directory         
            (default $WERF_GITERMINISM_CONFIG or werf-giterminism.yaml in working directory)
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-helm-dependencies=false
            Allow insecure oci registries to be used in the .helm/Chart.yaml dependencies           
            configuration (default $WERF_INSECURE_HELM_DEPENDENCIES)
      --insecure-registry=false
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
      --introspect-before-error=false
            Introspect failed stage in the clean state, before running all assembly instructions of 
            the stage
      --introspect-error=false
            Introspect failed stage in the state, right after running failed assembly instruction
      --introspect-stage=[]
            Introspect a specific stage. The option can be used multiple times to introspect        
            several stages.
            
            There are the following formats to use:
            * specify IMAGE_NAME/STAGE_NAME to introspect stage STAGE_NAME of either image or       
            artifact IMAGE_NAME
            * specify STAGE_NAME or */STAGE_NAME for the introspection of all existing stages with  
            name STAGE_NAME
            
            IMAGE_NAME is the name of an image or artifact described in werf.yaml, the nameless     
            image specified with ~.
            STAGE_NAME should be one of the following: from, beforeInstall,                         
            dependenciesBeforeInstall, gitArchive, install, dependenciesAfterInstall, beforeSetup,  
            dependenciesBeforeSetup, setup, dependenciesAfterSetup, gitCache, gitLatestPatch,       
            dockerInstructions, dockerfile
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
      --log-time=false
            Add time to log entries for precise event time tracking (default $WERF_LOG_TIME or      
            false).
      --log-time-format='2006-01-02T15:04:05Z07:00'
            Specify custom log time format (default $WERF_LOG_TIME_FORMAT or RFC3339 format).
      --log-verbose=false
            Enable verbose output (default $WERF_LOG_VERBOSE).
      --loose-giterminism=false
            Loose werf giterminism mode restrictions
  -p, --parallel=true
            Run in parallel (default $WERF_PARALLEL or true)
      --parallel-tasks-limit=5
            Parallel tasks limit, set -1 to remove the limitation (default                          
            $WERF_PARALLEL_TASKS_LIMIT or 5)
      --platform=[]
            Enable platform emulation when building images with werf, format: OS/ARCH[/VARIANT]     
            ($WERF_PLATFORM or $DOCKER_DEFAULT_PLATFORM by default)
      --repo=''
            Container registry storage address (default $WERF_REPO)
      --repo-container-registry=''
            Choose repo container registry implementation.
            The following container registries are supported: ecr, acr, default, dockerhub, gcr,    
            github, gitlab, harbor, quay.
            Default $WERF_REPO_CONTAINER_REGISTRY or auto mode (detect container registry by repo   
            address).
      --repo-docker-hub-password=''
            repo Docker Hub password (default $WERF_REPO_DOCKER_HUB_PASSWORD)
      --repo-docker-hub-token=''
            repo Docker Hub token (default $WERF_REPO_DOCKER_HUB_TOKEN)
      --repo-docker-hub-username=''
            repo Docker Hub username (default $WERF_REPO_DOCKER_HUB_USERNAME)
      --repo-github-token=''
            repo GitHub token (default $WERF_REPO_GITHUB_TOKEN)
      --repo-harbor-password=''
            repo Harbor password (default $WERF_REPO_HARBOR_PASSWORD)
      --repo-harbor-username=''
            repo Harbor username (default $WERF_REPO_HARBOR_USERNAME)
      --repo-quay-token=''
            repo quay.io token (default $WERF_REPO_QUAY_TOKEN)
  -Z, --require-built-images=false
            Requires all used images to be previously built and exist in repo. Exits with error if  
            needed images are not cached and so require to run build instructions (default          
            $WERF_REQUIRE_BUILT_IMAGES)
      --secondary-repo=[]
            Specify one or multiple secondary read-only repos with images that will be used as a    
            cache.
            Also, can be specified with $WERF_SECONDARY_REPO_* (e.g. $WERF_SECONDARY_REPO_1=...,    
            $WERF_SECONDARY_REPO_2=...)
      --skip-tls-verify-registry=false
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --ssh-key=[]
            Use only specific ssh key(s).
            Can be specified with $WERF_SSH_KEY_* (e.g. $WERF_SSH_KEY_REPO=~/.ssh/repo_rsa,         
            $WERF_SSH_KEY_NODEJS=~/.ssh/nodejs_rsa).
            Defaults to $WERF_SSH_KEY_*, system ssh-agent or ~/.ssh/{id_rsa|id_dsa}
  -S, --synchronization=''
            Address of synchronizer for multiple werf processes to work with a single repo.
            
            Default:
             - $WERF_SYNCHRONIZATION, or
             - :local if --repo is not specified, or
             - https://synchronization.werf.io if --repo has been specified.
            
            The same address should be specified for all werf processes that work with a single     
            repo. :local address allows execution of werf processes from a single host only
      --tag=[]
            Set a tag template (can specify multiple).
            It is necessary to use image name shortcut %image% or %image_slug% if multiple images   
            are exported (e.g. REPO:TAG-%image% or REPO-%image%:TAG)
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --virtual-merge=false
            Enable virtual/ephemeral merge commit mode when building current application state      
            ($WERF_VIRTUAL_MERGE by default)
```

