{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Deploy Helm chart specified by path

{{ header }} Syntax

```shell
werf helm deploy-chart CHART_DIR|CHART_REFERENCE RELEASE_NAME [options]
```

{{ header }} Examples

```shell
  # Deploy raw helm chart from current directory
  $ werf helm deploy-chart . myrelease

  # Deploy helm chart by chart reference
  $ werf helm deploy-chart stable/nginx-ingress myrelease

```

{{ header }} Options

```shell
      --ca-file='':
            verify certificates of HTTPS-enabled servers using this CA bundle (if using CHART as a  
            chart reference)
      --cert-file='':
            identify HTTPS client using this SSL certificate file (if using CHART as a chart        
            reference)
      --devel=false:
            use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set,   
            this is ignored (if using CHART as a chart reference)
      --dir='':
            Change to the specified directory to find werf.yaml config
      --helm-home='~/.helm':
            location of your Helm config. Defaults to $WERF_HELM_HOME, $HELM_HOME or ~/.helm
      --helm-release-storage-namespace='kube-system':
            Helm release storage namespace (same as --tiller-namespace for regular helm, default    
            $WERF_HELM_RELEASE_STORAGE_NAMESPACE, $TILLER_NAMESPACE or 'kube-system')
      --helm-release-storage-type='configmap':
            helm storage driver to use. One of 'configmap' or 'secret' (default                     
            $WERF_HELM_RELEASE_STORAGE_TYPE or 'configmap')
  -h, --help=false:
            help for deploy-chart
      --home-dir='':
            Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)
      --hooks-status-progress-period=5:
            Hooks status progress period in seconds. Set 0 to stop showing hooks status progress.   
            Defaults to $WERF_HOOKS_STATUS_PROGRESS_PERIOD_SECONDS or status progress period value
      --key-file='':
            identify HTTPS client using this SSL key file (if using CHART as a chart reference)
      --keyring='$HOME/.gnupg/pubring.gpg':
            keyring containing public keys (if using CHART as a chart reference)
      --kube-config='':
            Kubernetes config file path
      --kube-context='':
            Kubernetes config context (default $WERF_KUBE_CONTEXT)
      --log-color-mode='auto':
            Set log color mode.
            Supported on, off and auto (based on the stdout’s file descriptor referring to a        
            terminal) modes.
            Default $WERF_LOG_COLOR_MODE or auto mode.
      --log-pretty=true:
            Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or   
            true).
      --log-terminal-width=-1:
            Set log terminal width.
            Defaults to:
            * $WERF_LOG_TERMINAL_WIDTH
            * interactive terminal width or 140
      --namespace='':
            Namespace to install release into
      --password='':
            chart repository password (if using CHART as a chart reference)
      --prov=false:
            fetch the provenance file, but don't perform verification (if using CHART as a chart    
            reference)
      --repo='':
            chart repository url where to locate the requested chart (if using CHART as a chart     
            reference)
      --set=[]:
            Set helm values on the command line (can specify multiple or separate values with       
            commas: key1=val1,key2=val2)
      --set-string=[]:
            Set STRING helm values on the command line (can specify multiple or separate values     
            with commas: key1=val1,key2=val2)
      --status-progress-period=5:
            Status progress period in seconds. Set -1 to stop showing status progress. Defaults to  
            $WERF_STATUS_PROGRESS_PERIOD_SECONDS or 5 seconds
      --three-way-merge-mode='':
            Set three way merge mode for release.
            Supported 'enabled', 'disabled' and 'onlyNewReleases', see docs for more info           
            https://werf.io/documentation/reference/deploy_process/experimental_three_way_merge.html
  -t, --timeout=0:
            Resources tracking timeout in seconds
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)
      --username='':
            chart repository username (if using CHART as a chart reference)
      --values=[]:
            Specify helm values in a YAML file or a URL (can specify multiple)
      --verify=false:
            verify the package against its signature (if using CHART as a chart reference)
      --version='':
            specific version of a chart. Without this, the latest version is fetched (if using      
            CHART as a chart reference)
```

