{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Take locally extracted bundle or download bundle from the specified container registry using        
specified version tag or version mask and render it as Kubernetes manifests.

{{ header }} Syntax

```shell
werf bundle render [options]
```

{{ header }} Environments

```shell
  $WERF_SECRET_KEY  Use specified secret key to extract secrets for the deploy. Recommended way to  
                    set secret key in CI-system.
                    
                    Secret key also can be defined in files:
                    * ~/.werf/global_secret_key (globally),
                    * .werf_secret_key (per project)
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
  -b, --bundle-dir=""
            Get extracted bundle from directory instead of registry (default $WERF_BUNDLE_DIR)
      --container-registry-mirror=[]
            (Buildah-only) Use specified mirrors for docker.io
      --debug-templates=false
            Enable debug mode for Go templates (default $WERF_DEBUG_TEMPLATES or false)
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
      --home-dir=""
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --ignore-secret-key=false
            Disable secrets decryption (default $WERF_IGNORE_SECRET_KEY)
      --include-crds=true
            Include CRDs in the templated output (default $WERF_INCLUDE_CRDS)
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
      --kube-qps-limit=30
            Kubernetes client QPS limit (default $WERF_KUBE_QPS_LIMIT or 30)
      --kube-tls-server=""
            Server name to use for Kubernetes API server certificate validation. If it is not       
            provided, the hostname used to contact the server is used (default                      
            $WERF_KUBE_TLS_SERVER)
      --kube-token=""
            Kubernetes bearer token used for authentication (default $WERF_KUBE_TOKEN)
      --kube-version=""
            Set specific Capabilities.KubeVersion (default $WERF_KUBE_VERSION)
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
      --log-quiet=true
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
      --namespace=""
            Use specified Kubernetes namespace (default $WERF_NAMESPACE)
      --network-parallelism=30
            Parallelize some network operations (default $WERF_NETWORK_PARALLELISM or 30)
      --output=""
            Write render output to the specified file instead of stdout ($WERF_RENDER_OUTPUT by     
            default)
      --release=""
            Use specified Helm release name (default $WERF_RELEASE)
      --release-storage-sql-connection=""
            SQL Connection String for Helm SQL Storage (default                                     
            $WERF_RELEASE_STORAGE_SQL_CONNECTION)
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
  -s, --show-only=[]
            only show manifests rendered from the given templates
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
      --tag="latest"
            Provide exact tag version or semver-based pattern, werf will render the latest version  
            of the specified bundle ($WERF_TAG or latest by default)
      --tmp-dir=""
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --validate=false
            Validate your manifests against the Kubernetes cluster you are currently pointing at    
            (default $WERF_VALIDATE)
      --values=[]
            Specify helm values in a YAML file or a URL (can specify multiple). Also, can be        
            defined with $WERF_VALUES_* (e.g. $WERF_VALUES_1=values_1.yaml,                         
            $WERF_VALUES_2=values_2.yaml)
```

