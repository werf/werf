{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
List deployed releases

{{ header }} Syntax

```shell
werf release list [options]
```

{{ header }} Options

```shell
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
      --output-format="table"
            Output format. Options: table, yaml, json (default $WERF_OUTPUT_FORMAT or "table")
      --release-storage=""
            How releases should be stored (default $WERF_RELEASE_STORAGE)
      --release-storage-sql-connection=""
            SQL Connection String for Helm SQL Storage (default                                     
            $WERF_RELEASE_STORAGE_SQL_CONNECTION)
      --skip-tls-verify-kube=false
            Skip TLS certificate validation when accessing a Kubernetes cluster (default            
            $WERF_SKIP_TLS_VERIFY_KUBE)
      --tmp-dir=""
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

