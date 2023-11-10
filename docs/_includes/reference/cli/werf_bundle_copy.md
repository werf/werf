{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Take latest bundle from the specified container registry using specified version tag and copy it    
either into a different tag within the same container registry or into another container registry.

{{ header }} Syntax

```shell
werf bundle copy [options]
```

{{ header }} Options

```shell
      --docker-config=''
            Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or  
            ~/.docker (in the order of priority)
            Command needs granted permissions to read, pull and push images into the specified repos
      --from=''
            Source address of the bundle to copy, specify bundle archive using schema               
            `archive:PATH_TO_ARCHIVE.tar.gz`, specify remote bundle with schema                     
            `[docker://]REPO:TAG` or without schema.
  -C, --helm-compatible-chart=true
            Set chart name in the Chart.yaml of the published chart to the last path component of   
            container registry repo (for REGISTRY/PATH/TO/REPO address chart name will be REPO,     
            more info https://helm.sh/docs/topics/registries/#oci-feature-deprecation-and-behavior-c
            hanges-with-v370). In helm compatibility mode chart is fully conforming with the helm   
            OCI registry requirements. Default true or $WERF_HELM_COMPATIBLE_CHART.
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
      --platform=[]
            Enable platform emulation when building images with werf, format: OS/ARCH[/VARIANT]     
            ($WERF_PLATFORM or $DOCKER_DEFAULT_PLATFORM by default)
      --rename-chart=''
            Force setting of chart name in the Chart.yaml of the published chart to the specified   
            value (can be set by the $WERF_RENAME_CHART, no rename by default, could not be used    
            together with the `--helm-compatible-chart` option).
      --repo=''
            Deprecated param, use --from=ADDR instead. Source address of bundle which should be     
            copied.
      --skip-tls-verify-registry=false
            Skip TLS certificate validation when accessing a registry (default                      
            $WERF_SKIP_TLS_VERIFY_REGISTRY)
      --tag=''
            Deprecated param, use --from=REPO:TAG instead. Provide from tag version of the bundle   
            to copy ($WERF_TAG or latest by default).
      --tmp-dir=''
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --to=''
            Destination address of the bundle to copy, specify bundle archive using schema          
            `archive:PATH_TO_ARCHIVE.tar.gz`, specify remote bundle with schema                     
            `[docker://]REPO:TAG` or without schema.
      --to-tag=''
            Deprecated param, use --to=REPO:TAG instead. Provide to tag version of the bundle to    
            copy ($WERF_TO_TAG or same as --tag by default).
```

