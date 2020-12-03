{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Safely cleanup unused project images.

The command works according to special rules called cleanup policies, which the user defines in     
werf.yaml ([https://werf.io/documentation/reference/werf_yaml.html#configuring-cleanup-policies]({{ "/documentation/reference/werf_yaml.html#configuring-cleanup-policies" | relative_url }})).

It is safe to run this command periodically (daily is enough) by automated cleanup job in parallel  
with other werf commands such as build, converge and host cleanup.

{{ header }} Syntax

```shell
werf cleanup [options]
```

{{ header }} Examples

```shell
  $ werf cleanup --repo registry.mydomain.com/myproject/werf
```

{{ header }} Options

```shell
      --config=''
            Use custom configuration file (default $WERF_CONFIG or werf.yaml in working directory)
      --config-templates-dir=''
            Custom configuration templates directory (default $WERF_CONFIG_TEMPLATES_DIR or .werf   
            in working directory)
      --dir=''
            Use custom working directory (default $WERF_DIR or current directory)
      --disable-giterminism=false
            Disable werf giterminism mode (more info                                                
            https://werf.io/v1.2-alpha/documentation/advanced/configuration/giterminism.html,       
            default $WERF_DISABLE_GITERMINISM)
      --docker-config=''
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
            Command needs granted permissions to read, pull and delete images from the specified    
            repo
      --dry-run=false
            Indicate what the command would do without actually doing that (default $WERF_DRY_RUN)
      --env=''
            Use specified environment (default $WERF_ENV)
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-registry=false
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
      --keep-stages-built-within-last-n-hours=2
            Keep stages that were built within last hours (default                                  
            $WERF_KEEP_STAGES_BUILT_WITHIN_LAST_N_HOURS or 2)
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
      --non-strict-giterminism-inspection=false
            Change some errors to warnings during giterminism inspection (more info                 
            https://werf.io/v1.2-alpha/documentation/advanced/configuration/giterminism.html,       
            default $WERF_NON_STRICT_GITERMINISM_INSPECTION)
  -p, --parallel=true
            Run in parallel (default $WERF_PARALLEL)
      --parallel-tasks-limit=10
            Parallel tasks limit, set -1 to remove the limitation (default                          
            $WERF_PARALLEL_TASKS_LIMIT or 5)
      --repo=''
            Docker Repo to store stages (default $WERF_REPO)
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
      --repo-implementation=''
            Choose repo implementation.
            The following docker registry implementations are supported: ecr, acr, default,         
            dockerhub, gcr, github, gitlab, harbor, quay.
            Default $WERF_REPO_IMPLEMENTATION or auto mode (detect implementation by a registry).
      --repo-quay-token=''
            quay.io token (default $WERF_REPO_QUAY_TOKEN)
      --scan-context-namespace-only=false
            Scan for used images only in namespace linked with context for each available context   
            in kube-config (or only for the context specified with option --kube-context). When     
            disabled will scan all namespaces in all contexts (or only for the context specified    
            with option --kube-context). (Default $WERF_SCAN_CONTEXT_NAMESPACE_ONLY)
      --secondary-repo=[]
            Specify one or multiple secondary read-only repo with images that will be used as a     
            cache
      --skip-tls-verify-registry=false
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
  -S, --synchronization=''
            Address of synchronizer for multiple werf processes to work with a single repo.
            
            Default:
            * $WERF_SYNCHRONIZATION or
            * :local if --repo is not specified or
            * kubernetes://werf-synchronization if --repo is specified
            
            The same address should be specified for all werf processes that work with a single     
            repo. :local address allows execution of werf processes from a single host only
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --without-kube=false
            Do not skip deployed Kubernetes images (default $WERF_WITHOUT_KUBE)
```

