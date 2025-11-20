{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Take latest bundle from the specified container registry using specified version tag or version     
mask and plan its deployment as a helm chart into Kubernetes cluster.

{{ header }} Syntax

```shell
werf bundle plan [options]
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
      --container-registry-mirror=[]
            (Buildah-only) Use specified mirrors for docker.io
      --debug-templates=false
            Enable debug mode for Go templates (default $WERF_DEBUG_TEMPLATES or false)
      --deploy-graph-path=""
            Save deploy graph path to the specified file (by default $WERF_DEPLOY_GRAPH_PATH).      
            Extension must be .dot or not specified. If extension not specified, then .dot is used
      --diff-context-lines=3
            Show N lines of context around diffs ($WERF_DIFF_CONTEXT_LINES by default)
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
      --exit-code=false
            If true, returns exit code 0 if no changes, exit code 2 if any changes planned or exit  
            code 1 in case of an error (default $WERF_EXIT_CODE or false)
      --force-adoption=false
            Always adopt resources, even if they belong to a different Helm release (default        
            $WERF_FORCE_ADOPTION or false)
      --home-dir=""
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --ignore-secret-key=false
            Disable secrets decryption (default $WERF_IGNORE_SECRET_KEY)
      --insecure-helm-dependencies=false
            Allow insecure oci registries to be used in the Chart.yaml dependencies configuration   
            (default $WERF_INSECURE_HELM_DEPENDENCIES)
      --insecure-registry=false
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
      --kube-api-server=""
            Kubernetes API server address (default $WERF_KUBE_API_SERVER)
      --kube-auth-password=""
            Basic auth password for Kubernetes API (default $WERF_KUBE_AUTH_PASSWORD)
      --kube-auth-provider=""
            Auth provider name for authentication in Kubernetes API (default                        
            $WERF_KUBE_AUTH_PROVIDER)
      --kube-auth-provider-config=[]
            Auth provider config for authentication in Kubernetes API (default                      
            $WERF_KUBE_AUTH_PROVIDER_CONFIG)
      --kube-auth-username=""
            Basic auth username for Kubernetes API (default $WERF_KUBE_AUTH_USERNAME)
      --kube-burst-limit=100
            Kubernetes client burst limit (default $WERF_KUBE_BURST_LIMIT or 100)
      --kube-ca-data=""
            Pass Kubernetes API server TLS CA data (default $WERF_KUBE_CA_DATA)
      --kube-ca-path=""
            Kubernetes API server CA path (default $WERF_KUBE_CA_PATH)
      --kube-cert=""
            Path to PEM-encoded TLS client cert for connecting to Kubernetes API (default           
            $WERF_KUBE_CERT
      --kube-cert-data=""
            Pass PEM-encoded TLS client cert for connecting to Kubernetes API (default              
            $WERF_KUBE_CERT_DATA)
      --kube-config=""
            Kubernetes config file path (default $WERF_KUBE_CONFIG, or $WERF_KUBECONFIG, or         
            $KUBECONFIG)
      --kube-config-base64=""
            Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or            
            $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)
      --kube-context=""
            Kubernetes config context (default $WERF_KUBE_CONTEXT)
      --kube-context-cluster=""
            Use cluster from Kubeconfig for current context (default $WERF_KUBE_CONTEXT_CLUSTER)
      --kube-context-user=""
            Use user from Kubeconfig for current context (default $WERF_KUBE_CONTEXT_USER)
      --kube-impersonate-group=[]
            Sets Impersonate-Group headers when authenticating in Kubernetes. Can be also set with  
            $WERF_KUBE_IMPERSONATE_GROUP_* environment variables
      --kube-impersonate-uid=""
            Sets Impersonate-Uid header when authenticating in Kubernetes (default                  
            $WERF_KUBE_IMPERSONATE_UID)
      --kube-impersonate-user=""
            Sets Impersonate-User header when authenticating in Kubernetes (default                 
            $WERF_KUBE_IMPERSONATE_USER)
      --kube-key=""
            Path to PEM-encoded TLS client key for connecting to Kubernetes API (default            
            $WERF_KUBE_KEY)
      --kube-key-data=""
            Pass PEM-encoded TLS client key for connecting to Kubernetes API (default               
            $WERF_KUBE_KEY_DATA)
      --kube-proxy-url=""
            Proxy URL to use for proxying all requests to Kubernetes API (default                   
            $WERF_KUBE_PROXY_URL)
      --kube-qps-limit=30
            Kubernetes client QPS limit (default $WERF_KUBE_QPS_LIMIT or 30)
      --kube-request-timeout=0s
            Timeout for all requests to Kubernetes API (default $WERF_KUBE_REQUEST_TIMEOUT)
      --kube-tls-server=""
            Server name to use for Kubernetes API server certificate validation. If it is not       
            provided, the hostname used to contact the server is used (default                      
            $WERF_KUBE_TLS_SERVER)
      --kube-token=""
            Kubernetes bearer token used for authentication (default $WERF_KUBE_TOKEN)
      --kube-token-path=""
            Path to file with bearer token for authentication in Kubernetes (default                
            $WERF_KUBE_TOKEN_PATH)
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
      --namespace=""
            Use specified Kubernetes namespace (default $WERF_NAMESPACE)
      --network-parallelism=30
            Parallelize some network operations (default $WERF_NETWORK_PARALLELISM or 30)
      --no-final-tracking=false
            By default disable tracking operations that have no create/update/delete resource       
            operations after them, which are most tracking operations, to speed up the release      
            (default $WERF_NO_FINAL_TRACKING)
      --no-install-crds=false
            Do not install CRDs from "crds/" directories of installed charts (default               
            $WERF_NO_INSTALL_CRDS)
      --no-remove-manual-changes=false
            Don`t remove fields added manually to the resource in the cluster if fields aren`t      
            present in the manifest (default $WERF_NO_REMOVE_MANUAL_CHANGES)
      --provenance-keyring=""
            Path to keyring containing public keys to verify chart provenance (default              
            $WERF_PROVENANCE_KEYRING)
      --provenance-strategy=""
            Strategy for provenance verifying (default $WERF_PROVENANCE_STRATEGY).
      --release=""
            Use specified Helm release name (default $WERF_RELEASE)
      --release-info-annotations=[]
            Add annotations to release metadata (default $WERF_RELEASE_INFO_ANNOTATIONS)
      --release-label=[]
            Add Helm release labels (can specify multiple). Kind of labels depends or release       
            storage driver.
            Format: labelName=labelValue.
            Also, can be specified with $WERF_RELEASE_LABEL_* (e.g.                                 
            $WERF_RELEASE_LABEL_1=labelName1=labelValue1,                                           
            $WERF_RELEASE_LABEL_2=labelName2=labelValue2)
      --release-storage=""
            How releases should be stored (default $WERF_RELEASE_STORAGE)
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
      --runtime-annotations=[]
            Add annotations which will not trigger resource updates to all resources (default       
            $WERF_RUNTIME_ANNOTATIONS)
      --runtime-labels=[]
            Add labels which will not trigger resource updates to all resources (default            
            $WERF_RUNTIME_LABELS)
      --secret-key=""
            Secret key (default $WERF_SECRET_KEY)
      --secret-values=[]
            Specify helm secret values in a YAML file (can specify multiple). Also, can be defined  
            with $WERF_SECRET_VALUES_* (e.g. $WERF_SECRET_VALUES_ENV=.helm/secret_values_test.yaml, 
            $WERF_SECRET_VALUES_DB=.helm/secret_values_db.yaml)
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
      --set-json=[]
            Set new values, where the key is the value path and the value is JSON (can specify      
            multiple or separate values with commas: key1=val1,key2=val2).
            Also, can be defined with $WERF_SET_JSON_* (e.g. $WERF_SET_JSON_1=key1=val1,            
            $WERF_SET_JSON_2=key2=val2)
      --set-literal=[]
            Set new values, where the key is the value path and the value is the value. The value   
            will always become a literal string (can specify multiple or separate values with       
            commas: key1=val1,key2=val2).)
            Also, can be defined with $WERF_SET_LITERAL_* (e.g. $WERF_SET_LITERAL_1=key1=val1,      
            $WERF_SET_LITERAL_2=key2=val2)
      --set-runtime-json=[]
            Set new keys in $.Runtime, where the key is the value path and the value is JSON. This  
            is meant to be generated inside the program, so use --set-json instead, unless you know 
            what you are doing. Can specify multiple or separate values with commas:                
            key1=val1,key2=val2.
            Also, can be defined with $WERF_SET_RUNTIME_JSON_* (e.g.                                
            $WERF_SET_RUNTIME_JSON_1=key1=val1, $WERF_SET_RUNTIME_JSON_2=key2=val2)
      --set-string=[]
            Set STRING helm values on the command line (can specify multiple or separate values     
            with commas: key1=val1,key2=val2).
            Also, can be defined with $WERF_SET_STRING_* (e.g. $WERF_SET_STRING_1=key1=val1,        
            $WERF_SET_STRING_2=key2=val2)
      --show-insignificant-diffs=false
            Show insignificant diff lines ($WERF_SHOW_INSIGNIFICANT_DIFFS by default)
      --show-sensitive-diffs=false
            Show sensitive diff lines ($WERF_SHOW_SENSITIVE_DIFFS by default)
      --show-verbose-crd-diffs=false
            Show verbose CRD diff lines ($WERF_SHOW_VERBOSE_CRD_DIFFS by default)
      --show-verbose-diffs=true
            Show verbose diff lines ($WERF_SHOW_VERBOSE_DIFFS by default)
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
            Provide exact tag version or semver-based pattern, werf will install or upgrade to the  
            latest version of the specified bundle ($WERF_TAG or latest by default)
      --templates-allow-dns=false
            Allow performing DNS requests in templating (default $WERF_TEMPLATES_ALLOW_DNS)
      --tmp-dir=""
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --values=[]
            Specify helm values in a YAML file or a URL (can specify multiple). Also, can be        
            defined with $WERF_VALUES_* (e.g. $WERF_VALUES_1=.helm/values_1.yaml,                   
            $WERF_VALUES_2=.helm/values_2.yaml)
```

