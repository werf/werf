{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Delete application from Kubernetes.

Helm Release will be purged and optionally Kubernetes Namespace.

Environment is a required param for the dismiss by default, because it is needed to construct Helm  
Release name and Kubernetes Namespace. Either --env or $WERF_ENV should be specified for command

{{ header }} Syntax

```shell
werf dismiss [options]
```

{{ header }} Examples

```shell
  # Dismiss project named 'myproject' previously deployed app from 'dev' environment; helm release name and namespace will be named as 'myproject-dev'
  $ werf dismiss --env dev

  # Dismiss project with namespace
  $ werf dismiss --env my-feature-branch --with-namespace

  # Dismiss project using specified helm release name and namespace
  $ werf dismiss --release myrelease --namespace myns
```

{{ header }} Options

```shell
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
      --env=''
            Use specified environment (default $WERF_ENV)
      --helm-release-storage-namespace='kube-system'
            Helm release storage namespace (same as --tiller-namespace for regular helm, default    
            $WERF_HELM_RELEASE_STORAGE_NAMESPACE, $TILLER_NAMESPACE or 'kube-system')
      --helm-release-storage-type='configmap'
            helm storage driver to use. One of 'configmap' or 'secret' (default                     
            $WERF_HELM_RELEASE_STORAGE_TYPE or 'configmap')
  -h, --help=false
            help for dismiss
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
      --namespace=''
            Use specified Kubernetes namespace (default [[ project ]]-[[ env ]] template or         
            deploy.namespace custom template from werf.yaml or $WERF_NAMESPACE)
      --release=''
            Use specified Helm release name (default [[ project ]]-[[ env ]] template or            
            deploy.helmRelease custom template from werf.yaml or $WERF_RELEASE)
      --releases-history-max=0
            Max releases to keep in release storage. Can be set by environment variable             
            $WERF_RELEASES_HISTORY_MAX. By default werf keeps all releases.
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
      --with-hooks=true
            Delete Helm Release hooks getting from existing revisions (default $WERF_WITH_HOOKS or  
            true)
      --with-namespace=false
            Delete Kubernetes Namespace after purging Helm Release (default $WERF_WITH_NAMESPACE)
```

