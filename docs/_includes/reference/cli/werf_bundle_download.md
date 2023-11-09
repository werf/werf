{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Take latest bundle from the specified container registry using specified version tag or version     
mask and unpack it into provided directory (or into directory named as a resulting chart in the     
current working directory).

{{ header }} Syntax

```shell
werf bundle download [options]
```

{{ header }} Options

```shell
  -d, --destination=''
            Download bundle into the provided directory ($WERF_DESTINATION or chart-name by default)
      --docker-config=''
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
      --home-dir=''
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --insecure-helm-dependencies=false
            Allow insecure oci registries to be used in the .helm/Chart.yaml dependencies           
            configuration (default $WERF_INSECURE_HELM_DEPENDENCIES)
      --insecure-registry=false
            Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)
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
      --skip-tls-verify-registry=false
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --tag='latest'
            Provide exact tag version or semver-based pattern, werf will install or upgrade to the  
            latest version of the specified bundle ($WERF_TAG or latest by default)
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

