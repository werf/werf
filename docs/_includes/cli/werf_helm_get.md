{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Get release templates and values by release name

{{ header }} Syntax

```shell
werf helm get RELEASE_NAME [options]
```

{{ header }} Options

```shell
      --helm-release-storage-namespace='kube-system':
            Helm release storage namespace (same as --tiller-namespace for regular helm, default    
            $WERF_HELM_RELEASE_STORAGE_NAMESPACE, $TILLER_NAMESPACE or 'kube-system')
      --helm-release-storage-type='configmap':
            helm storage driver to use. One of 'configmap' or 'secret' (default                     
            $WERF_HELM_RELEASE_STORAGE_TYPE or 'configmap')
  -h, --help=false:
            help for get
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --kube-config='':
            Kubernetes config file path
      --kube-context='':
            Kubernetes config context (default $WERF_KUBE_CONTEXT)
      --log-color-mode='auto':
            Set log color mode.
            Supported on, off and auto (based on the stdout’s file descriptor referring to a        
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-debug=false:
            Enable debug (default $WERF_LOG_DEBUG).
      --log-pretty=true:
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or   
            true).
      --log-quiet=false:
            Disable explanatory output (default $WERF_LOG_QUIET).
      --log-terminal-width=-1:
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --log-verbose=false:
            Enable verbose output (default $WERF_LOG_VERBOSE).
      --revision=0:
            Get the named release by revision (use latest revision by default)
      --template='':
            Go template for formatting the output, eg: {{.Release.Name}}
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
```

