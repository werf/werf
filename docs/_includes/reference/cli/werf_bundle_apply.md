{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Take latest bundle from the specified container registry using specified version tag or version     
mask and apply it as a helm chart into Kubernetes cluster.

{{ header }} Syntax

```shell
werf bundle apply [options]
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
      --atomic=false
            Enable auto rollback of the failed release to the previous deployed release version     
            when current deploy process have failed ($WERF_ATOMIC by default)
  -R, --auto-rollback=false
            Enable auto rollback of the failed release to the previous deployed release version     
            when current deploy process have failed ($WERF_AUTO_ROLLBACK by default)
      --config=""
            Use custom configuration file (default $WERF_CONFIG or werf.yaml in the project         
            directory)
      --config-templates-dir=""
            Custom configuration templates directory (default $WERF_CONFIG_TEMPLATES_DIR or .werf   
            in working directory)
      --container-registry-mirror=[]
            (Buildah-only) Use specified mirrors for docker.io
      --debug-templates=false
            Enable debug mode for Go templates (default $WERF_DEBUG_TEMPLATES or false)
      --deploy-graph-path=""
            Save deploy graph path to the specified file (by default $WERF_DEPLOY_GRAPH_PATH).      
            Extension must be .dot or not specified. If extension not specified, then .dot is used
      --deploy-report-path=""
            Change deploy report path and format (by default $WERF_DEPLOY_REPORT_PATH or            
            ".werf-deploy-report.json" if not set). Extension must be .json for JSON format. If     
            extension not specified, then .json is used
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
      --disable-default-secret-values=false
            Do not use secret values from the default .helm/secret-values.yaml file (default        
            $WERF_DISABLE_DEFAULT_SECRET_VALUES or false)
      --disable-default-values=false
            Do not use values from the default .helm/values.yaml file (default                      
            $WERF_DISABLE_DEFAULT_VALUES or false)
      --docker-config=""
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
            Command needs granted permissions to read, pull and push images into the specified      
            repo, to pull base images
      --env=""
            Use specified environment (default $WERF_ENV)
      --force-adoption=false
            Always adopt resources, even if they belong to a different Helm release (default        
            $WERF_FORCE_ADOPTION or false)
      --git-work-tree=""
            Use specified git work tree dir (default $WERF_WORK_TREE or lookup for directory that   
            contains .git in the current or parent directories)
      --giterminism-config=""
            Custom path to the giterminism configuration file relative to working directory         
            (default $WERF_GITERMINISM_CONFIG or werf-giterminism.yaml in working directory)
      --home-dir=""
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --hooks-status-progress-period=5
            Hooks status progress period in seconds. Set 0 to stop showing hooks status progress.   
            Defaults to $WERF_HOOKS_STATUS_PROGRESS_PERIOD_SECONDS or status progress period value
      --ignore-secret-key=false
            Disable secrets decryption (default $WERF_IGNORE_SECRET_KEY)
      --insecure-helm-dependencies=false
            Allow insecure oci registries to be used in the Chart.yaml dependencies configuration   
            (default $WERF_INSECURE_HELM_DEPENDENCIES)
      --insecure-registry=false
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
      --kube-api-server=""
            Kubernetes API server address (default $WERF_KUBE_API_SERVER)
      --kube-burst-limit=100
            Kubernetes client burst limit (default $WERF_KUBE_BURST_LIMIT or 100)
      --kube-ca-path=""
            Kubernetes API server CA path (default $WERF_KUBE_CA_PATH)
      --kube-config=""
            Kubernetes config file path (default $WERF_KUBE_CONFIG, or $WERF_KUBECONFIG, or         
            $KUBECONFIG)
      --kube-config-base64=""
            Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or            
            $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)
      --kube-context=""
            Kubernetes config context (default $WERF_KUBE_CONTEXT)
      --kube-ignore-pods-logs=false
            Don`t collect logs from PODs in Kubernetes (default $WERF_KUBE_IGNORE_PODS_LOGS)
      --kube-qps-limit=30
            Kubernetes client QPS limit (default $WERF_KUBE_QPS_LIMIT or 30)
      --kube-tls-server=""
            Server name to use for Kubernetes API server certificate validation. If it is not       
            provided, the hostname used to contact the server is used (default                      
            $WERF_KUBE_TLS_SERVER)
      --kube-token=""
            Kubernetes bearer token used for authentication (default $WERF_KUBE_TOKEN)
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
      --namespace=""
            Use specified Kubernetes namespace (default $WERF_NAMESPACE)
      --network-parallelism=30
            Parallelize some network operations (default $WERF_NETWORK_PARALLELISM or 30)
      --no-install-crds=false
            Do not install CRDs from "crds/" directories of installed charts (default               
            $WERF_NO_INSTALL_CRDS)
      --release=""
            Use specified Helm release name (default $WERF_RELEASE)
      --release-label=[]
            Add Helm release labels (can specify multiple). Kind of labels depends or release       
            storage driver.
            Format: labelName=labelValue.
            Also, can be specified with $WERF_RELEASE_LABEL_* (e.g.                                 
            $WERF_RELEASE_LABEL_1=labelName1=labelValue1,                                           
            $WERF_RELEASE_LABEL_2=labelName2=labelValue2)
      --releases-history-max=5
            Max releases to keep in release storage ($WERF_RELEASES_HISTORY_MAX or 5 by default)
      --render-subchart-notes=false
            If set, render subchart notes along with the parent (by default                         
            $WERF_RENDER_SUBCHART_NOTES or false)
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
      --rollback-graph-path=""
            Save rollback graph path to the specified file (by default $WERF_ROLLBACK_GRAPH_PATH).  
            Extension must be .dot or not specified. If extension not specified, then .dot is used
      --save-deploy-report=false
            Save deploy report (by default $WERF_SAVE_DEPLOY_REPORT or false). Its path and format  
            configured with --deploy-report-path
      --secret-values=[]
            Specify helm secret values in a YAML file (can specify multiple). Also, can be defined  
            with $WERF_SECRET_VALUES_* (e.g. $WERF_SECRET_VALUES_ENV=secret_values_test.yaml,       
            $WERF_SECRET_VALUES_DB=secret_values_db.yaml)
      --set=[]
            Set helm values on the command line (can specify multiple or separate values with       
            commas: key1=val1,key2=val2).
            Also, can be defined with $WERF_SET_* (e.g. $WERF_SET_1=key1=val1,                      
            $WERF_SET_2=key2=val2)
      --set-docker-config-json-value=false
            Shortcut to set current docker config into the .Values.dockerconfigjson
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
      --skip-tls-verify-helm-dependencies=false
            Skip TLS certificate validation when accessing a Helm charts repository (default        
            $WERF_SKIP_TLS_VERIFY_HELM_DEPENDENCIES)
      --skip-tls-verify-kube=false
            Skip TLS certificate validation when accessing a Kubernetes cluster (default            
            $WERF_SKIP_TLS_VERIFY_KUBE)
      --skip-tls-verify-registry=false
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --status-progress-period=5
            Status progress period in seconds. Set -1 to stop showing status progress. Defaults to  
            $WERF_STATUS_PROGRESS_PERIOD_SECONDS or 5 seconds
      --tag="latest"
            Provide exact tag version or semver-based pattern, werf will install or upgrade to the  
            latest version of the specified bundle ($WERF_TAG or latest by default)
  -t, --timeout=0
            Resources tracking timeout in seconds ($WERF_TIMEOUT by default)
      --tmp-dir=""
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --values=[]
            Specify helm values in a YAML file or a URL (can specify multiple). Also, can be        
            defined with $WERF_VALUES_* (e.g. $WERF_VALUES_1=values_1.yaml,                         
            $WERF_VALUES_2=values_2.yaml)
```

