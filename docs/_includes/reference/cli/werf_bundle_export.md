{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Export bundle into the provided directory (or into directory named as a resulting chart in the current working directory). `werf bundle` contains built images defined in the `werf.yaml`, Helm chart, Service values which contain built images tags, any custom values and set values params provided during publish invocation, werf service templates and values.

{{ header }} Syntax

```shell
werf bundle export [IMAGE_NAME...] [options]
```

{{ header }} Options

```shell
      --add-annotation=[]
            Add annotation to deploying resources (can specify multiple).
            Format: annoName=annoValue.
            Also, can be specified with $WERF_ADD_ANNOTATION_* (e.g.                                
            $WERF_ADD_ANNOTATION_1=annoName1=annoValue1,                                            
            $WERF_ADD_ANNOTATION_2=annoName2=annoValue2)
      --add-label=[]
            Add label to deploying resources (can specify multiple).
            Format: labelName=labelValue.
            Also, can be specified with $WERF_ADD_LABEL_* (e.g.                                     
            $WERF_ADD_LABEL_1=labelName1=labelValue1, $WERF_ADD_LABEL_2=labelName2=labelValue2)
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
      --build-report-path=''
            Change build report path and format (by default $WERF_BUILD_REPORT_PATH or              
            ".werf-build-report.json" if not set). Extension must be either .json for JSON format   
            or .env for env-file format. If extension not specified, then .json is used
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
  -d, --destination=''
            Export bundle into the provided directory ($WERF_DESTINATION or chart-name by default)
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
      --disable-auto-host-cleanup=false
            Disable auto host cleanup procedure in main werf commands like werf-build,              
            werf-converge and other (default disabled or WERF_DISABLE_AUTO_HOST_CLEANUP)
      --disable-default-secret-values=false
            Do not use secret values from the default .helm/secret-values.yaml file (default        
            $WERF_DISABLE_DEFAULT_SECRET_VALUES or false)
      --disable-default-values=false
            Do not use values from the default .helm/values.yaml file (default                      
            $WERF_DISABLE_DEFAULT_VALUES or false)
      --docker-config=''
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
            Command needs granted permissions to read, pull and push images into the specified repo 
            and to pull base images
      --docker-server-storage-path=''
            Use specified path to the local docker server storage to check docker storage volume    
            usage while performing garbage collection of local docker images (detect local docker   
            server storage path by default or use $WERF_DOCKER_SERVER_STORAGE_PATH)
      --env=''
            Use specified environment (default $WERF_ENV)
      --final-repo=''
            Container registry storage address (default $WERF_FINAL_REPO)
      --final-repo-container-registry=''
            Choose final-repo container registry implementation.
            The following container registries are supported: ecr, acr, default, dockerhub, gcr,    
            github, gitlab, harbor, quay, selectel.
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
      --final-repo-selectel-account=''
            final-repo Selectel account (default $WERF_FINAL_REPO_SELECTEL_ACCOUNT)
      --final-repo-selectel-password=''
            final-repo Selectel password (default $WERF_FINAL_REPO_SELECTEL_PASSWORD)
      --final-repo-selectel-username=''
            final-repo Selectel username (default $WERF_FINAL_REPO_SELECTEL_USERNAME)
      --final-repo-selectel-vpc=''
            final-repo Selectel VPC (default $WERF_FINAL_REPO_SELECTEL_VPC)
      --final-repo-selectel-vpc-id=''
            final-repo Selectel VPC ID (default $WERF_FINAL_REPO_SELECTEL_VPC_ID)
      --git-work-tree=''
            Use specified git work tree dir (default $WERF_WORK_TREE or lookup for directory that   
            contains .git in the current or parent directories)
      --giterminism-config=''
            Custom path to the giterminism configuration file relative to working directory         
            (default $WERF_GITERMINISM_CONFIG or werf-giterminism.yaml in working directory)
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --ignore-secret-key=false
            Disable secrets decryption (default $WERF_IGNORE_SECRET_KEY)
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
            Loose werf giterminism mode restrictions (NOTE: not all restrictions can be removed,    
            more info https://werf.io/documentation/usage/project_configuration/giterminism.html,   
            default $WERF_LOOSE_GITERMINISM)
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
            github, gitlab, harbor, quay, selectel.
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
      --repo-selectel-account=''
            repo Selectel account (default $WERF_REPO_SELECTEL_ACCOUNT)
      --repo-selectel-password=''
            repo Selectel password (default $WERF_REPO_SELECTEL_PASSWORD)
      --repo-selectel-username=''
            repo Selectel username (default $WERF_REPO_SELECTEL_USERNAME)
      --repo-selectel-vpc=''
            repo Selectel VPC (default $WERF_REPO_SELECTEL_VPC)
      --repo-selectel-vpc-id=''
            repo Selectel VPC ID (default $WERF_REPO_SELECTEL_VPC_ID)
      --report-format=''
            DEPRECATED: use --save-build-report with optional --build-report-path.
            Report format: json or envfile (json or $WERF_REPORT_FORMAT by default) json:
            	{
            	  "Images": {
            		"<WERF_IMAGE_NAME>": {
            			"WerfImageName": "<WERF_IMAGE_NAME>",
            			"DockerRepo": "<REPO>",
            			"DockerTag": "<TAG>"
            			"DockerImageName": "<REPO>:<TAG>",
            			"DockerImageID": "<SHA256>",
            			"DockerImageDigest": "<SHA256>",
            		},
            		...
            	  }
            	}
            envfile:
            	WERF_<FORMATTED_WERF_IMAGE_NAME>_DOCKER_IMAGE_NAME=<REPO>:<TAG>
            	...
            <FORMATTED_WERF_IMAGE_NAME> is werf image name from werf.yaml modified according to the 
            following rules:
            - all characters are uppercase (app -> APP);
            - charset /- is replaced with _ (DEV/APP-FRONTEND -> DEV_APP_FRONTEND)
      --report-path=''
            DEPRECATED: use --save-build-report with optional --build-report-path.
            Report save path ($WERF_REPORT_PATH by default)
  -Z, --require-built-images=false
            Requires all used images to be previously built and exist in repo. Exits with error if  
            needed images are not cached and so require to run build instructions (default          
            $WERF_REQUIRE_BUILT_IMAGES)
      --save-build-report=false
            Save build report (by default $WERF_SAVE_BUILD_REPORT or false). Its path and format    
            configured with --build-report-path
      --secondary-repo=[]
            Specify one or multiple secondary read-only repos with images that will be used as a    
            cache.
            Also, can be specified with $WERF_SECONDARY_REPO_* (e.g. $WERF_SECONDARY_REPO_1=...,    
            $WERF_SECONDARY_REPO_2=...)
      --set=[]
            Set helm values on the command line (can specify multiple or separate values with       
            commas: key1=val1,key2=val2).
            Also, can be defined with $WERF_SET_* (e.g. $WERF_SET_1=key1=val1,                      
            $WERF_SET_2=key2=val2)
      --set-file=[]
            Set values from respective files specified via the command line (can specify multiple   
            or separate values with commas: key1=path1,key2=path2).
            Also, can be defined with $WERF_SET_FILE_* (e.g. $WERF_SET_FILE_1=key1=path1,           
            $WERF_SET_FILE_2=key2=val2)
      --set-string=[]
            Set STRING helm values on the command line (can specify multiple or separate values     
            with commas: key1=val1,key2=val2).
            Also, can be defined with $WERF_SET_STRING_* (e.g. $WERF_SET_STRING_1=key1=val1,        
            $WERF_SET_STRING_2=key2=val2)
  -L, --skip-dependencies-repo-refresh=false
            Do not refresh helm chart repositories locally cached index
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
      --use-custom-tag=''
            Use a tag alias in helm templates instead of an image content-based tag (NOT            
            RECOMMENDED).
            The alias may contain the following shortcuts:
            - %image%, %image_slug% or %image_safe_slug% to use the image name (necessary if there  
            is more than one image in the werf config);
            - %image_content_based_tag% to use a content-based tag.
            For cleaning custom tags and associated content-based tag are treated as one.
            Also, can be defined with $WERF_USE_CUSTOM_TAG (e.g. $WERF_USE_CUSTOM_TAG="%image%-tag")
      --values=[]
            Specify helm values in a YAML file or a URL (can specify multiple).
            Also, can be defined with $WERF_VALUES_* (e.g. $WERF_VALUES_1=.helm/values_1.yaml,      
            $WERF_VALUES_2=.helm/values_2.yaml)
      --virtual-merge=false
            Enable virtual/ephemeral merge commit mode when building current application state      
            ($WERF_VIRTUAL_MERGE by default)
      --without-images=false
            Disable building of images defined in the werf.yaml (if any) and usage of such images   
            in the .helm/templates ($WERF_WITHOUT_IMAGES or false by default — e.g. enable all      
            images defined in the werf.yaml by default)
```

