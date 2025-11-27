{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Rollback the Helm release.

The result of rollback command is an application deployed into Kubernetes. Command will create the new release based on the one we rollback to and wait until all resources of the release will become ready.



{{ header }} Syntax

```shell
werf rollback [options]
```

{{ header }} Examples

```shell
# Rollback the Helm release to the specified revision
werf rollback --revision 10
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
      --delete-propagation=""
            Set default delete propagation strategy (default $WERF_DELETE_PROPAGATION or            
            Foreground).
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
      --hooks-status-progress-period=0
            No-op
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
      --loose-giterminism=false
            Loose werf giterminism mode restrictions
      --namespace=""
            Use specified Kubernetes namespace (default [[ project ]]-[[ env ]] template or         
            deploy.namespace custom template from werf.yaml or $WERF_NAMESPACE)
      --network-parallelism=30
            Parallelize some network operations (default $WERF_NETWORK_PARALLELISM or 30)
      --no-final-tracking=false
            By default disable tracking operations that have no create/update/delete resource       
            operations after them, which are most tracking operations, to speed up the release      
            (default $WERF_NO_FINAL_TRACKING)
      --no-notes=false
            Don`t show release notes at the end of the release (default $WERF_NO_NOTES)
      --no-pod-logs=false
            Disable Pod logs collection and printing (default $WERF_NO_POD_LOGS or false)
      --no-remove-manual-changes=false
            Don`t remove fields added manually to the resource in the cluster if fields aren`t      
            present in the manifest (default $WERF_NO_REMOVE_MANUAL_CHANGES)
      --release=""
            Use specified Helm release name (default [[ project ]]-[[ env ]] template or            
            deploy.helmRelease custom template from werf.yaml or $WERF_RELEASE)
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
      --releases-history-max=5
            Max releases to keep in release storage ($WERF_RELEASES_HISTORY_MAX or 5 by default)
      --revision=0
            Revision number to rollback to (if not specified, rolls back to previous revision)
      --rollback-graph-path=""
            Save rollback graph path to the specified file (by default $WERF_ROLLBACK_GRAPH_PATH).  
            Extension must be .dot or not specified. If extension not specified, then .dot is used
      --rollback-report-path=""
            Change rollback report path and format (by default $WERF_ROLLBACK_REPORT_PATH or        
            ".werf-rollback-report.json" if not set). Extension must be .json for JSON format. If   
            extension not specified, then .json is used
      --runtime-annotations=[]
            Add annotations which will not trigger resource updates to all resources (default       
            $WERF_RUNTIME_ANNOTATIONS)
      --runtime-labels=[]
            Add labels which will not trigger resource updates to all resources (default            
            $WERF_RUNTIME_LABELS)
      --save-rollback-report=false
            Save rollback report (by default $WERF_SAVE_ROLLBACK_REPORT or false). Its path and     
            format configured with --rollback-report-path
      --skip-tls-verify-kube=false
            Skip TLS certificate validation when accessing a Kubernetes cluster (default            
            $WERF_SKIP_TLS_VERIFY_KUBE)
      --status-progress-period=5
            Status progress period in seconds. Set -1 to stop showing status progress. Defaults to  
            $WERF_STATUS_PROGRESS_PERIOD_SECONDS or 5 seconds
  -t, --timeout=0
            Resources tracking timeout in seconds ($WERF_TIMEOUT by default)
      --tmp-dir=""
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

